package tasks

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/spot"
)

var (
	hamCancel context.CancelFunc
	har       *hamAlertReader

	errEmptyMsg = fmt.Errorf("empty message from HamAlert server")
	badMsg      = "unexpected message from HamAlert server %s"
)

// our own type so we can make this "quitable"
type hamAlertReader struct {
	c   net.Conn
	r   *bufio.Reader
	ctx context.Context
}

func NewHamAlertReader() (*hamAlertReader, error) {
	// connect
	con, err := net.Dial("tcp", config.ClusterServices.HamAlert.HostPort)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	return &hamAlertReader{c: con, r: bufio.NewReader(con)}, nil
}

func (har *hamAlertReader) stopped() bool {
	return har.ctx != nil && har.ctx.Err() != nil
}

// isHamAlertPrompt reports the DXSpider idle prompt, for example
// "AD8FD de HamAlert 25-10-2026 0057>".
func isHamAlertPrompt(msg string) bool {
	msg = strings.TrimSpace(msg)
	return strings.Contains(msg, " de HamAlert ") && strings.HasSuffix(msg, ">")
}

func (har *hamAlertReader) readString(delim byte) (*string, error) {
	msg, err := har.r.ReadString(delim)
	if err != nil {
		if har.stopped() {
			return nil, nil
		}
		if err == io.EOF {
			var empty string
			return &empty, nil
		}
		log.Printf("%+v", err)
		return nil, err
	}

	return &msg, nil
}

func (har *hamAlertReader) readAndExpect(delim byte, expect string) error {
	msg, err := har.readString(delim)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	if msg == nil {
		return errEmptyMsg
	}
	if !strings.Contains(*msg, expect) {
		err := fmt.Errorf(badMsg, *msg)
		log.Printf("%+v", err)
		return err
	}

	return nil
}

func (har *hamAlertReader) writeString(s string) error {
	_, err := fmt.Fprintln(har.c, s)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// gatherHamAlerts does the real work of inetgrating with HamAlert to collect spot
// only returns if we couldn't connect to HamAlert or connection is closed
func (har *hamAlertReader) gatherHamAlerts() error {
	var err error
	var msg *string

	// login
	err = har.readAndExpect('\n', "HamAlert")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// hamalert username
	err = har.readAndExpect(':', "login:")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	err = har.writeString(config.ClusterServices.HamAlert.Username)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// hamalert password
	err = har.readAndExpect(':', "password:")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	err = har.writeString(config.ClusterServices.HamAlert.Password)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// verify greeting
	msg, err = har.readString('>')
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	if msg == nil {
		return errEmptyMsg
	}
	if !strings.Contains(*msg, fmt.Sprintf("%s de HamAlert", strings.ToUpper(config.ClusterServices.HamAlert.Username))) {
		err := fmt.Errorf("authentication failure %s", *msg)
		log.Printf("%+v", err)
		return err
	}

	// throw out next line, part of greeting
	_, err = har.readString('\n')
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// prime with a little history
	err = har.writeString("show/dx 5")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// wait for new lines to show up and add spot entries for them
	re := regexp.MustCompile(`^DX de\s+([^:]+):\s+([^\s]+)\s+([^\s]+)\s+(.*)([0-9]{4})Z`)
	for {
		if har.stopped() {
			return nil
		}

		msg, err = har.readString('\n')
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		if msg == nil {
			return nil
		}

		//log.Printf("%s", *msg)

		// parse for elements
		match := re.FindStringSubmatch(*msg)
		if match == nil {
			if isHamAlertPrompt(*msg) {
				continue
			}
			if strings.HasPrefix(strings.ToLower(*msg), "no spots") {
				select {
				case <-har.ctx.Done():
					return nil
				case <-time.After(30 * time.Second):
				}
				continue
			}

			err := fmt.Errorf("could not match on message '%s'", *msg)
			log.Printf("%+v", err)
			return err
		}

		//log.Printf("%+v", match)

		err = spot.Add(match[5], match[3], match[2], match[4], match[1])
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}
}

// StartHamAlerts starts the collection of spots from HamAlert
func StartHamAlerts() {
	mutexQuitChannels.Lock()
	if hamCancel != nil {
		mutexQuitChannels.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	hamCancel = cancel
	mutexQuitChannels.Unlock()

	go runHamAlerts(ctx)
}

// runHamAlerts stays connected until ctx is cancelled.
func runHamAlerts(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		setTaskStatus(TaskHamAlert, TaskStatusOK)

		reader, err := NewHamAlertReader()
		if err != nil {
			log.Printf("%+v", err)
		} else {
			reader.ctx = ctx
			if !trackHamAlertReader(ctx, reader) {
				_ = reader.c.Close()
				return
			}

			err = reader.gatherHamAlerts()
			if err != nil {
				log.Printf("%+v", err)
			}
		}

		if ctx.Err() != nil {
			return
		}

		setTaskStatus(TaskHamAlert, TaskStatusFailed)

		if err != nil {
			// some connection problem, pause before retry
			select {
			case <-ctx.Done():
				return
			case <-time.After(30 * time.Second):
			}
		}
	}
}

// trackHamAlertReader records the live connection unless stop already won the race.
func trackHamAlertReader(ctx context.Context, reader *hamAlertReader) bool {
	mutexQuitChannels.Lock()
	defer mutexQuitChannels.Unlock()

	if ctx.Err() != nil || hamCancel == nil {
		return false
	}

	har = reader
	return true
}

// StopHamAlerts shutdowns the collection of spots from HamAlert
func StopHamAlerts() {
	mutexQuitChannels.Lock()
	cancel := hamCancel
	reader := har
	hamCancel = nil
	har = nil
	if cancel != nil {
		cancel()
	}
	mutexQuitChannels.Unlock()

	if reader != nil && reader.c != nil {
		_ = reader.c.Close()
	}
}

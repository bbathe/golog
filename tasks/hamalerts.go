package tasks

import (
	"bufio"
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
	quitHamAlert chan bool
	har          *hamAlertReader

	errEmptyMsg = fmt.Errorf("empty message from HamAlert server")
	badMsg      = "unexpected message from HamAlert server %s"
)

// our own type so we can make this "quitable"
type hamAlertReader struct {
	c net.Conn
	r *bufio.Reader
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

func (har *hamAlertReader) readString(delim byte) (*string, error) {
	msg, err := har.r.ReadString(delim)
	if err != nil {
		select {
		case <-quitHamAlert:
			return nil, nil
		default:
			if err == io.EOF {
				var empty string
				return &empty, nil
			}
			log.Printf("%+v", err)
			return nil, err
		}
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
			if strings.HasPrefix(*msg, "No Spots") {
				// pause
				time.Sleep(30 * time.Second)
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
	quitHamAlert = make(chan bool)

	go func() {
		// this gives us reconnect and ability to stop/start
		for {
			select {
			case <-quitHamAlert:
				return
			default:
				setTaskStatus(TaskHamAlert, TaskStatusOK)

				var err error
				har, err = NewHamAlertReader()
				if err != nil {
					log.Printf("%+v", err)
				} else {
					err = har.gatherHamAlerts()
					if err != nil {
						log.Printf("%+v", err)
					}
				}

				setTaskStatus(TaskHamAlert, TaskStatusFailed)

				if err != nil {
					// some connection problem, pause before retry
					time.Sleep(30 * time.Second)
				}
			}
		}
	}()
}

// StopHamAlerts shutdowns the collection of spots from HamAlert
func StopHamAlerts() {
	close(quitHamAlert)

	har.c.Close()
}

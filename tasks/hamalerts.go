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

	errEmptyMsg = fmt.Errorf("empty message from HamAlert server")
)

// our own type so we can make this "quitable"
type hamAlertReader struct {
	r *bufio.Reader
}

func NewHamAlertReader(rd io.Reader) *hamAlertReader {
	return &hamAlertReader{r: bufio.NewReader(rd)}
}

func (r *hamAlertReader) ReadString(delim byte) (*string, error) {
	msg, err := r.r.ReadString(delim)
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
				err := gatherHamAlerts()
				if err != nil {
					log.Printf("%+v", err)
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
}

// gatherHamAlerts does the real work of inetgrating with HamAlert to collect spot
// only returns if we couldn't connect to HamAlert or connection is closed
func gatherHamAlerts() error {
	var err error
	var msg *string
	badMsg := "unexpected message from HamAlert server %s"

	// connect
	var conHamAlert net.Conn
	conHamAlert, err = net.Dial("tcp", config.ClusterServices.HamAlert.HostPort)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	defer conHamAlert.Close()

	r := NewHamAlertReader(conHamAlert)

	// login
	msg, err = r.ReadString('\n')
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	if msg == nil {
		return errEmptyMsg
	}
	if !strings.Contains(*msg, "HamAlert") {
		err := fmt.Errorf(badMsg, *msg)
		log.Printf("%+v", err)
		return err
	}

	// hamalert username
	msg, err = r.ReadString(':')
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	if msg == nil {
		return errEmptyMsg
	}
	if !strings.HasSuffix(*msg, "login:") {
		err := fmt.Errorf(badMsg, *msg)
		log.Printf("%+v", err)
		return err
	}
	fmt.Fprintf(conHamAlert, config.ClusterServices.HamAlert.Username+"\n")

	// hamalert password
	msg, err = r.ReadString(':')
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	if msg == nil {
		return errEmptyMsg
	}
	if !strings.HasSuffix(*msg, "password:") {
		err := fmt.Errorf(badMsg, *msg)
		log.Printf("%+v", err)
		return err
	}
	fmt.Fprintf(conHamAlert, config.ClusterServices.HamAlert.Password+"\n")

	// verify greeting
	msg, err = r.ReadString('>')
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
	_, err = r.ReadString('\n')
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// prime with a little history
	fmt.Fprintf(conHamAlert, "show/dx 5\n")

	// wait for new lines to show up and add spot entries for them
	re := regexp.MustCompile(`^DX de\s+([^:]+):\s+([^\s]+)\s+([^\s]+)\s+(.*)([0-9]{4})Z`)
	for {
		msg, err = r.ReadString('\n')
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

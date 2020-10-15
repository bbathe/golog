package tasks

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/spot"
)

var (
	conHamAlert  net.Conn
	quitHamAlert chan bool
)

func StartHamAlerts() {
	quitHamAlert = make(chan bool)

	go func() {
		// this gives us reconnect and ability to stop/start
		for {
			select {
			case <-quitHamAlert:
				return
			default:
				GatherHamAlerts()
			}
		}
	}()
}

func StopHamAlerts() {
	close(quitHamAlert)

	conHamAlert.Close()
}

func GatherHamAlerts() {
	var err error
	var msg string

	// connect
	conHamAlert, err = net.Dial("tcp", config.ClusterServices.HamAlert.HostPort)
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	defer conHamAlert.Close()

	r := bufio.NewReader(conHamAlert)

	// login
	msg, err = r.ReadString('\n')
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	if !strings.Contains(msg, "HamAlert") {
		err := fmt.Errorf("doesn't appear to be HamAlert server %s", msg)
		log.Printf("%+v", err)
		return
	}

	// hamalert username
	msg, err = r.ReadString(':')
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	if strings.HasSuffix(msg, "login:") {
		fmt.Fprintf(conHamAlert, config.ClusterServices.HamAlert.Username+"\n")
	} else {
		log.Printf("%+v", err)
		return
	}

	// hamalert password
	msg, err = r.ReadString(':')
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	if strings.HasSuffix(msg, "password:") {
		fmt.Fprintf(conHamAlert, config.ClusterServices.HamAlert.Password+"\n")
	} else {
		log.Printf("%+v", err)
		return
	}

	// verify greeting
	msg, err = r.ReadString('>')
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	if !strings.Contains(msg, fmt.Sprintf("%s de HamAlert", strings.ToUpper(config.ClusterServices.HamAlert.Username))) {
		err := fmt.Errorf("authentication failure %s", msg)
		log.Printf("%+v", err)
		return
	}

	// throw out next line, part of greeting
	_, err = r.ReadString('\n')
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	// prime with a little history
	fmt.Fprintf(conHamAlert, "show/dx 5\n")

	re := regexp.MustCompile(`^[^:]+:\s+([^\s]+)\s+([^\s]+)\s+(.*)([0-9]{4})Z`)
	for {
		msg, err = r.ReadString('\n')
		if err != nil {
			select {
			case <-quitHamAlert:
				return
			default:
				log.Printf("%+v", err)
			}
			return
		}

		// parse for elements
		match := re.FindStringSubmatch(msg)
		if match == nil {
			err := fmt.Errorf("could not match on message '%s'", msg)
			log.Printf("%+v", err)
			return
		}

		err = spot.Add(match[4], match[2], match[1], match[3])
		if err != nil {
			log.Printf("%+v", err)
			return
		}
	}
}

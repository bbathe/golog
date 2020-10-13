package tasks

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/ui"
)

func GatherHamAlerts() {
	var err error
	var msg string

	// connect
	conn, err := net.Dial("tcp", config.ClusterServices.HamAlert.HostPort)
	if err != nil {
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}
	defer conn.Close()

	r := bufio.NewReader(conn)

	// login
	msg, err = r.ReadString('\n')
	if err != nil {
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}
	if !strings.Contains(msg, "HamAlert") {
		err := fmt.Errorf("doesn't appear to be HamAlert server %s", msg)
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}

	// hamalert username
	msg, err = r.ReadString(':')
	if err != nil {
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}
	if strings.HasSuffix(msg, "login:") {
		fmt.Fprintf(conn, config.ClusterServices.HamAlert.Username+"\n")
	} else {
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}

	// hamalert password
	msg, err = r.ReadString(':')
	if err != nil {
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}
	if strings.HasSuffix(msg, "password:") {
		fmt.Fprintf(conn, config.ClusterServices.HamAlert.Password+"\n")
	} else {
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}

	// verify greeting
	msg, err = r.ReadString('>')
	if err != nil {
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}
	if !strings.Contains(msg, fmt.Sprintf("%s de HamAlert", strings.ToUpper(config.ClusterServices.HamAlert.Username))) {
		err := fmt.Errorf("authentication failure %s", msg)
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}

	// throw out next line, part of greeting
	_, err = r.ReadString('\n')
	if err != nil {
		log.Printf("%+v", err)
		ui.MsgError(nil, err)
		return
	}

	// prime with a little history
	fmt.Fprintf(conn, "show/dx 5\n")

	re := regexp.MustCompile(`^[^:]+:\s+([^\s]+)\s+([^\s]+)\s+(.*)([0-9]{4})Z`)
	for {
		msg, err = r.ReadString('\n')
		if err != nil {
			log.Printf("%+v", err)
			ui.MsgError(nil, err)
			return
		}

		// parse for elements
		match := re.FindStringSubmatch(msg)
		if match == nil {
			err := fmt.Errorf("could not match on message '%s'", msg)
			log.Printf("%+v", err)
			ui.MsgError(nil, err)
			return
		}

		err = ui.AddSpot(match[4], match[2], match[1], match[3])
		if err != nil {
			log.Printf("%+v", err)
			ui.MsgError(nil, err)
			return
		}
	}
}

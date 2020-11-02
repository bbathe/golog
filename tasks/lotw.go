package tasks

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/qso"
)

var muxLotwUpload sync.Mutex

// QSLLotw uploads all QSOs to LoTW that are older than config.QSLDelay minutes old
func QSLLotw() error {
	muxLotwUpload.Lock()
	defer muxLotwUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLLotw, config.LogbookServices.QSLDelay)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	if len(qsos) > 0 {
		err = uploadQSOsToLoTW(qsos, false)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	return nil
}

// QSLLotwFinal uploads all QSOs to LoTW regardless of how old they are
func QSLLotwFinal() {
	muxLotwUpload.Lock()
	defer muxLotwUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLLotw, 0)
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	if len(qsos) > 0 {
		err = uploadQSOsToLoTW(qsos, true)
		if err != nil {
			log.Printf("%+v", err)
			return
		}
	}
}

// uploadQSOsToLoTW leverages tqsl to upload qsos to LoTW
func uploadQSOsToLoTW(qsos []qso.QSO, forceBulk bool) error {
	// only bulk so make sure there's a reasonable batch
	if len(qsos) > 16 || (forceBulk && len(qsos) > 0) {

		// form working file name
		fname := filepath.Join(config.WorkingDirectory, "LoTW-"+time.Now().UTC().Format("2006-Jan-02_15-04-05")+".adif")

		// write qsos as adif to file
		err := adif.WriteToFile(qsos, fname)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		// setup command execution, capturing stdout & stderr
		// #nosec G204
		cmd := exec.Command(
			config.LogbookServices.TQSL.ExeLocation,
			"--quiet",
			"--batch",
			"--nodate",
			"--upload",
			fmt.Sprintf("--location=%s", config.LogbookServices.TQSL.StationLocationName),
			fname,
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// doit!
		err = cmd.Run()
		if err != nil {
			log.Printf("error: %+v", err)
			log.Printf("stdout: %s", stdout.String())
			log.Printf("stderr: %s", stderr.String())
			return err
		}

		// set as sent in db
		err = qso.UpdateQSLsToSent(qsos, qso.QSLLotw)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	return nil
}

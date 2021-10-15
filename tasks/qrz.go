package tasks

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/util"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/qso"
)

var muxQrzUpload sync.Mutex

// QSLQrz uploads all QSOs to QRZ.com that are older than config.QSLDelay minutes old
func QSLQrz() error {
	muxQrzUpload.Lock()
	defer muxQrzUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLQrz, config.LogbookServices.QSLDelay)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	if len(qsos) > 0 {
		err = uploadQSOsToQRZ(qsos)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	return nil
}

// QSLQrzFinal uploads all QSOs to QRZ.com regardless of how old they are
func QSLQrzFinal() {
	muxQrzUpload.Lock()
	defer muxQrzUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLQrz, 0)
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	if len(qsos) > 0 {
		err = uploadQSOsToQRZ(qsos)
		if err != nil {
			log.Printf("%+v", err)
			return
		}
	}
}

// uploadQSOsToQRZ uploads qsos to QRZ.com
func uploadQSOsToQRZ(qsos []qso.QSO) error {
	// save all the qsos we are uploading to file
	fname := filepath.Join(config.WorkingDirectory, "QRZ-"+time.Now().UTC().Format("2006-Jan-02_15-04-05")+".adif")

	// write qsos as adif to file
	err := adif.WriteToFile(qsos, fname)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	i := 0
	for _, q := range qsos {
		if i > 0 {
			// pause between uploads
			time.Sleep(1 * time.Second)
		}
		i++

		// set the required form fields
		formData := url.Values{}
		s, err := adif.QSOToADIFRecord(q)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		formData.Set("ADIF", s)
		formData.Set("KEY", config.LogbookServices.QRZ.APIKey)
		formData.Set("ACTION", "INSERT")
		formData.Set("OPTION", "REPLACE")

		// setup request
		reqBody := formData.Encode()
		req, err := http.NewRequest("POST", "https://logbook.qrz.com/api", strings.NewReader(reqBody))
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(reqBody)))

		// do POST
		client := http.Client{
			Timeout: 5 * time.Minute,
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		defer resp.Body.Close()

		// get response body
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("returned bad statuscode")
			log.Printf("%+v", err)
			log.Printf("StatusCode: %d", resp.StatusCode)
			log.Printf("Header: %s", resp.Header)
			log.Printf("Body: %s", string(respBody))
			return err
		}

		// parse response to see if it was really a success
		m, err := url.ParseQuery(string(respBody))
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		if m.Get("RESULT") == "FAIL" {
			err := fmt.Errorf("api returned bad status")
			log.Printf("%+v", err)
			log.Printf("StatusCode: %d", resp.StatusCode)
			log.Printf("Header: %s", resp.Header)
			log.Printf("Return Values: %+v", m)
			log.Printf("Body: %s", string(respBody))
			return err
		}

		// set as sent in db
		err = qso.UpdateQSLsToSent([]qso.QSO{q}, qso.QSLQrz)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	// only keep the last 5 files of ours in the working directory
	err = util.DeleteHistoricalFiles(5, config.WorkingDirectory, "QRZ-", ".adif")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

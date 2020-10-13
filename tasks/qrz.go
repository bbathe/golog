package tasks

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/bbathe/golog/adif"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/qso"
	"github.com/bbathe/golog/ui"
)

var muxQrzUpload sync.Mutex

// QSLQrz uploads all QSOs to QRZ.com that are older than config.QSLDelay minutes old
func QSLQrz() {
	muxQrzUpload.Lock()
	defer muxQrzUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLQrz, config.LogbookServices.QSLDelay)
	if err != nil {
		ui.MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	if len(qsos) > 0 {
		err = uploadQSOsToQRZ(qsos)
		if err != nil {
			ui.MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}

		ui.RefreshQSOs()
	}
}

// QSLQrzFinal uploads all QSOs to QRZ.com regardless of how old they are
func QSLQrzFinal() {
	muxQrzUpload.Lock()
	defer muxQrzUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLQrz, 0)
	if err != nil {
		ui.MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	if len(qsos) > 0 {
		err = uploadQSOsToQRZ(qsos)
		if err != nil {
			ui.MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}

		ui.RefreshQSOs()
	}
}

// uploadQSOsToQRZ uploads qsos to QRZ.com
func uploadQSOsToQRZ(qsos []qso.QSO) error {
	for _, q := range qsos {
		s, err := adif.QSOToADIFRecord(q)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		reqBody := &bytes.Buffer{}
		w := multipart.NewWriter(reqBody)

		// add station_callsign to adif
		err = w.WriteField("ADIF", fmt.Sprintf("<station_callsign:%d>%s%s", len(config.Station.Callsign), config.Station.Callsign, s))
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		// set the other form fields required
		err = w.WriteField("KEY", config.LogbookServices.QRZ.APIKey)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		err = w.WriteField("ACTION", "INSERT")
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		// done forming request body
		err = w.Close()
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		// setup request
		req, err := http.NewRequest("POST", "https://logbook.qrz.com/api", reqBody)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

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
		if m.Get("RESULT") != "OK" {
			err := fmt.Errorf("api returned bad status")
			log.Printf("%+v", err)
			log.Printf("StatusCode: %d", resp.StatusCode)
			log.Printf("Header: %s", resp.Header)
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

	return nil
}

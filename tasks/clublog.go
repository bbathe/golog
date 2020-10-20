package tasks

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bbathe/golog/adif"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/qso"
)

var muxClublogUpload sync.Mutex

// QSLClublog uploads all QSOs to Club Log that are older than config.QSLDelay minutes old
func QSLClublog() error {
	muxClublogUpload.Lock()
	defer muxClublogUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLClublog, config.LogbookServices.QSLDelay)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	if len(qsos) > 0 {
		err = uploadQSOsToClublog(qsos, false)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	return nil
}

// QSLClublogFinal uploads all QSOs to Club Log regardless of how old they are
func QSLClublogFinal() {
	muxClublogUpload.Lock()
	defer muxClublogUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLClublog, 0)
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	if len(qsos) > 0 {
		err = uploadQSOsToClublog(qsos, true)
		if err != nil {
			log.Printf("%+v", err)
			return
		}
	}
}

// uploadQSOsToClublog uploads qsos to Club Log
func uploadQSOsToClublog(qsos []qso.QSO, forceBulk bool) error {
	reqBody := &bytes.Buffer{}
	w := multipart.NewWriter(reqBody)
	var url string

	// handle with bulk call or realtime call?
	if len(qsos) > 15 || (forceBulk && len(qsos) > 1) {
		// form working file name
		fname := filepath.Join(config.WorkingDirectory, "Clublog-"+time.Now().UTC().Format("2006-Jan-02_15-04-05")+".adif")

		// write qsos as adif to file
		err := adif.WriteToFile(qsos, fname)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		// open file with qso data
		f, err := os.Open(fname)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		defer f.Close()

		// create & populate form part with file data
		p, err := w.CreateFormFile("file", filepath.Base(fname))
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		_, err = io.Copy(p, f)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		url = "https://clublog.org/putlogs.php"
	} else {
		// single adif record upload
		qsos = qsos[0:1]

		s, err := adif.QSOToADIFRecord(qsos[0])
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		err = w.WriteField("adif", s)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		url = "https://clublog.org/realtime.php"
	}

	// set the other form fields required
	err := w.WriteField("email", config.LogbookServices.ClubLog.Email)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	err = w.WriteField("password", config.LogbookServices.ClubLog.Password)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	err = w.WriteField("callsign", config.LogbookServices.ClubLog.Callsign)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	err = w.WriteField("api", config.LogbookServices.ClubLog.APIKey)
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
	req, err := http.NewRequest("POST", url, reqBody)
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

	// set as sent in db
	err = qso.UpdateQSLsToSent(qsos, qso.QSLClublog)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

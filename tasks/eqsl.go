package tasks

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/qso"
	"github.com/bbathe/golog/util"
)

var muxEqslUpload sync.Mutex

// QSLEqsl uploads all QSOs to eQSL that are older than config.QSLDelay minutes old
func QSLEqsl() error {
	muxEqslUpload.Lock()
	defer muxEqslUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLEqsl, config.LogbookServices.QSLDelay)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	if len(qsos) > 0 {
		err = uploadQSOsToEqsl(qsos)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	return nil
}

// QSLEqslFinal uploads all QSOs to eQSL regardless of how old they are
func QSLEqslFinal() {
	muxEqslUpload.Lock()
	defer muxEqslUpload.Unlock()

	qsos, err := qso.FindQSLsToSend(qso.QSLEqsl, 0)
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	if len(qsos) > 0 {
		err = uploadQSOsToEqsl(qsos)
		if err != nil {
			log.Printf("%+v", err)
			return
		}
	}
}

// uploadQSOsToEqsl uploads qsos to eQSL
func uploadQSOsToEqsl(qsos []qso.QSO) error {
	// form working file name
	fname := filepath.Join(config.WorkingDirectory, "eQSL-"+time.Now().UTC().Format("2006-Jan-02_15-04-05")+".adif")

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
	reqBody := &bytes.Buffer{}
	w := multipart.NewWriter(reqBody)
	p, err := w.CreateFormFile("Filename", filepath.Base(fname))
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	_, err = io.Copy(p, f)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// set the other form fields required
	err = w.WriteField("eQSL_User", config.LogbookServices.EQSL.Username)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	err = w.WriteField("eQSL_Pswd", config.LogbookServices.EQSL.Password)
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
	req, err := http.NewRequest("POST", "https://www.eQSL.cc/qslcard/ImportADIF.cfm", reqBody)
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
	respBody, err := io.ReadAll(resp.Body)
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

	if strings.Contains(strings.ToLower(string(respBody)), "error") {
		err := fmt.Errorf("returned error")
		log.Printf("%+v", err)
		log.Printf("StatusCode: %d", resp.StatusCode)
		log.Printf("Header: %s", resp.Header)
		log.Printf("Body: %s", string(respBody))
		return err
	}

	// set as sent in db
	err = qso.UpdateQSL(qsos, qso.QSLEqsl, qso.Sent)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// only keep the last 5 files of ours in the working directory
	err = util.DeleteHistoricalFiles(5, config.WorkingDirectory, "eQSL-", ".adif")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

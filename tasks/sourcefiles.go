package tasks

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/qso"
	"github.com/bbathe/golog/ui"
)

var muxSourceFiles sync.Mutex

// SourceFiles is the task that aggregates all changes from the adifs being monitored and inserts them into the qso database
func SourceFiles() {
	muxSourceFiles.Lock()
	defer muxSourceFiles.Unlock()

	// form merged filename
	mergedFile := filepath.Join(config.WorkingDirectory, "QSOS-"+time.Now().UTC().Format("2006-Jan-02_15-04-05")+".adif")

	// create a file with changes from all the sourcefiles merged into one
	err := createMergedChanges(mergedFile)
	if err != nil {
		ui.MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	// see if there are any qsos to process
	n, err := getFileSize(mergedFile)
	if err != nil {
		ui.MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	// remove file if no qsos
	if n == 0 {
		err = os.Remove(mergedFile)
		if err != nil {
			ui.MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}

		// done
		return
	}

	// read in qsos
	qsos, err := adif.ReadFromFile(mergedFile, qso.NotSent, qso.NotSent, qso.NotSent, qso.NotSent)
	if err != nil {
		ui.MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	// put in db
	err = qso.BulkAdd(qsos)
	if err != nil {
		ui.MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	ui.RefreshQSOs()
}

// getFileSize returns the size of file fname
func getFileSize(fname string) (int64, error) {
	fi, err := os.Stat(fname)
	if err != nil {
		log.Printf("%+v %s", err, fname)
		return 0, err
	}

	// return size
	return fi.Size(), nil
}

// createMergedChanges cycles thru the sourcefiles in config and creates merged file with the new adif records since the last run
func createMergedChanges(mergedFile string) error {
	configChanged := false

	// open file to merge to
	fmerged, err := os.Create(mergedFile)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	defer fmerged.Close()

	// process all sourcefiles
	for i, wf := range config.SourceFiles {
		// open file to get adif records added since last execution
		f, err := os.Open(wf.Location)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		defer f.Close()

		// get to last location
		_, err = f.Seek(wf.Offset, 0)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		// be sure to write only whole adif records
		var nwrite int64
		var record string
		var lowerRecord string
		c := []byte{0}
		for {
			_, err := f.Read(c)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Printf("%+v", err)
					return err
				}
			} else {
				// accumulate characters to test for whole record
				record += string(c)
				lowerRecord += strings.ToLower(string(c))

				// got to the end of a record?
				if strings.HasSuffix(lowerRecord, "<eor>") {
					// write to merged file
					nw, err := fmerged.WriteString(record)
					if err != nil {
						log.Printf("%+v", err)
						return err
					}

					// accumlate bytes written
					nwrite += int64(nw)

					// reset record
					record = ""
					lowerRecord = ""
				}
			}
		}

		// anything processed?
		if nwrite == 0 {
			continue
		}

		// update offset
		config.SourceFiles[i].Offset += nwrite
		configChanged = true
	}

	if configChanged {
		err = config.Write()
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	return nil
}

package tasks

import (
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/config"
)

var muxSourceFiles sync.Mutex

// SourceFiles is the task that aggregates all changes from the adifs being monitored and inserts them into the qso database
func SourceFiles() error {
	muxSourceFiles.Lock()
	defer muxSourceFiles.Unlock()

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

		// be sure to process only whole adif records
		var nprocessed int64
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
					// parse record to QSO
					qso, err := adif.QSOFromADIFRecord(record)
					if err != nil {
						log.Printf("%+v", err)
						return err
					}

					// if not set, assume current station callsign
					if qso.StationCallsign == "" {
						qso.StationCallsign = config.Station.Callsign
					}

					// persist to database
					err = qso.Add()
					if err != nil {
						log.Printf("%+v", err)
						return err
					}

					// accumlate bytes processed
					nprocessed += int64(len(record))

					// reset record
					record = ""
					lowerRecord = ""
				}
			}
		}

		// anything processed?
		if nprocessed == 0 {
			continue
		}

		// update offset
		config.SourceFiles[i].Offset += nprocessed

		// write out changes individually
		err = config.Write()
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	return nil
}

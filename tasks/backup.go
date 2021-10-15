package tasks

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/qso"
)

var muxBackup sync.Mutex

func Backup() {
	muxBackup.Lock()
	defer muxBackup.Unlock()

	// form backup file name for QSOs
	fname := filepath.Join(config.BackupDirectory, "BackupQSOs-"+time.Now().UTC().Format("2006-Jan-02_15-04-05")+".adif")

	// get all qsos & write them out to file
	qs, err := qso.All()
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	err = adif.WriteToFile(qs, fname)
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	// only keep the last 5 backups
	files, err := ioutil.ReadDir(config.BackupDirectory)
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	cnt := 0
	for _, file := range files {
		// only deal with our files
		if !file.IsDir() && strings.HasPrefix(file.Name(), "BackupQSOs-") && strings.HasSuffix(file.Name(), ".adif") {
			cnt++

			if cnt > 5 {
				err := os.Remove(filepath.Join(config.BackupDirectory, file.Name()))
				if err != nil {
					log.Printf("%+v", err)
					return
				}
			}
		}
	}
}

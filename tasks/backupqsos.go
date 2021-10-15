package tasks

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/qso"
	"github.com/bbathe/golog/util"
)

var muxBackupQSOs sync.Mutex

func BackupQSOs() {
	muxBackupQSOs.Lock()
	defer muxBackupQSOs.Unlock()

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
	err = util.DeleteHistoricalFiles(5, config.BackupDirectory, "BackupQSOs-", ".adif")
	if err != nil {
		log.Printf("%+v", err)
		return
	}
}

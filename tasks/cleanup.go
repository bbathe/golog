package tasks

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bbathe/golog/config"
)

var muxCleanup sync.Mutex

func Cleanup() {
	muxCleanup.Lock()
	defer muxCleanup.Unlock()

	if config.WorkingDirectory == "" {
		// done
		return
	}

	// cleanup working directory
	f, err := os.Open(config.WorkingDirectory)
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		log.Printf("%+v", err)
		return
	}

	// only keep archive files for 3 days
	now := time.Now()
	then := now.Add(-(time.Duration(3*24) * time.Hour))
	for _, file := range files {
		if file.ModTime().Before(then) {
			err = os.Remove(filepath.Join(config.WorkingDirectory, file.Name()))
			if err != nil && !os.IsNotExist(err) {
				log.Printf("%+v", err)
				return
			}
		}
	}
}

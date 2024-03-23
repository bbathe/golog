package util

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DeleteHistoricalFiles only keeps the latest num files matching prefix & suffix in directory
func DeleteHistoricalFiles(num int, directory, prefix, suffix string) error {
	// get listing of directory
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	sort.Slice(files, func(i, j int) bool {
		ifi, err := files[i].Info()
		if err != nil {
			return true
		}
		jfi, err := files[j].Info()
		if err != nil {
			return true
		}

		return ifi.ModTime().After(jfi.ModTime())
	})

	cnt := 0
	for _, file := range files {
		// filter files based on prefix & suffix
		if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) && strings.HasSuffix(file.Name(), suffix) {
			cnt++

			// start deleting if we got more than num
			if cnt > num {
				err := os.Remove(filepath.Join(directory, file.Name()))
				if err != nil {
					log.Printf("%+v", err)
					return err
				}
			}
		}
	}

	return nil
}

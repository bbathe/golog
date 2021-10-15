package util

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DeleteHistoricalFiles only keeps the latest num files matching prefix & suffix in directory
func DeleteHistoricalFiles(num int, directory, prefix, suffix string) error {
	// get listing of directory
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
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

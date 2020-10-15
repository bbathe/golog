package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/db"
	"github.com/bbathe/golog/tasks"
	"github.com/bbathe/golog/ui"
)

func main() {
	// show file & location, date & time
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	// process command line
	var configFile string
	flg := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flg.StringVar(&configFile, "config", "", "Configuration file")
	err := flg.Parse(os.Args[1:])
	if err != nil {
		err := fmt.Errorf("%s\n\nUsage of %s\n  -config string\n    Configuration file", err.Error(), os.Args[0])
		ui.MsgError(nil, err)
		log.Fatalf("%+v", err)
	}

	// by default the log file is in the same directory as the executable with the same base name
	fn, err := os.Executable()
	if err != nil {
		ui.MsgError(nil, err)
		log.Fatalf("%+v", err)
	}
	basefn := strings.TrimSuffix(fn, path.Ext(fn))

	// log to file
	f, err := os.OpenFile(basefn+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		ui.MsgError(nil, err)
		log.Fatalf("%+v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// read config
	var cfn string
	if len(configFile) > 0 {
		// if user passed a filename, use that
		cfn = configFile
	} else {
		// default config file is in the same directory as the executable with the same base name
		cfn = basefn + ".yaml"
	}

	err = config.Read(cfn)
	if err != nil {
		// if there was an error, let them fix it
		err = ui.OptionsWindow(nil)
		if err != nil {
			ui.MsgError(nil, err)
			log.Printf("%+v", err)
		}

		ui.MsgInformation(nil, "restart for any changes to take effect")
		return
	}

	// make sure temp directory exists
	err = os.MkdirAll(config.WorkingDirectory, os.ModePerm)
	if err != nil {
		ui.MsgError(nil, err)
		log.Fatalf("%+v", err)
	}

	// open our databases
	err = db.OpenQSODb()
	if err != nil {
		ui.MsgError(nil, err)
		log.Fatalf("%+v", err)
	}

	err = db.NewSpotDb()
	if err != nil {
		ui.MsgError(nil, err)
		log.Fatalf("%+v", err)
	}

	// read in lookups
	err = config.ReadLookupsFromFile(filepath.Join(filepath.Dir(basefn), "lookups.yaml"))
	if err != nil {
		ui.MsgError(nil, err)
		log.Fatalf("%+v", err)
	}

	// start background tasks
	go func() {
		tasks.Start()
	}()

	// show app, doesn't come back until main window closed
	err = ui.GoLogWindow()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// stop background tasks
	tasks.Shutdown()

	// persist config
	err = config.Write()
	if err != nil {
		ui.MsgError(nil, err)
		log.Fatalf("%+v", err)
	}
}

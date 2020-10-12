package config

import (
	"io/ioutil"
	"log"

	"github.com/lxn/walk"
	"gopkg.in/yaml.v2"
)

type station struct {
	Callsign string
}

type qsodatabase struct {
	Location string
}

type mainwinrectangle struct {
	X             int `yaml:"topleftx"`
	Y             int `yaml:"toplefty"`
	Width, Height int
}

func (mwr *mainwinrectangle) FromBounds(bounds walk.Rectangle) {
	mwr.X = bounds.X
	mwr.Y = bounds.Y
	mwr.Width = bounds.Width
	mwr.Height = bounds.Height
}

func (mwr *mainwinrectangle) ToBounds() walk.Rectangle {
	return walk.Rectangle{
		X:      mwr.X,
		Y:      mwr.Y,
		Width:  mwr.Width,
		Height: mwr.Height,
	}
}

type ui struct {
	MainWinRectangle mainwinrectangle
}

type sourcefile struct {
	Location string
	Offset   int64
}

type clublog struct {
	Email    string
	Password string
	Callsign string
	APIKey   string
}

type eqsl struct {
	Username string
	Password string
}

type qrz struct {
	APIKey string
}

type tqsl struct {
	ExeLocation         string
	StationLocationName string
}

type logbookservices struct {
	TQSL    tqsl
	EQSL    eqsl
	ClubLog clublog
	QRZ     qrz
}

type hamalert struct {
	HostPort string
	Username string
	Password string
}

type clusterservices struct {
	HamAlert hamalert
}

// configuration holds the application configuration
type configuration struct {
	ArchiveDirectory string
	QSLDelay         int
	QSOHistory       int
	QSOLimit         int
	Station          station
	QSODatabase      qsodatabase
	UI               ui
	SourceFiles      []sourcefile
	LogbookServices  logbookservices
	Clusterservices  clusterservices
}

var (
	configFile string

	ArchiveDirectory string
	QSLDelay         int
	QSOHistory       int
	QSOLimit         int

	Station     station
	QSODatabase qsodatabase
	UI          ui
	SourceFiles []sourcefile
	TQSL        tqsl
	EQSL        eqsl
	ClubLog     clublog
	QRZ         qrz

	HamAlert hamalert
)

// Read reads application configuration from file fname
func Read(fname string) error {
	// save for write later
	configFile = fname

	// #nosec G304
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	var c configuration
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// unwrap
	Station = c.Station
	QSODatabase = c.QSODatabase
	UI = c.UI
	SourceFiles = c.SourceFiles
	ArchiveDirectory = c.ArchiveDirectory
	QSLDelay = c.QSLDelay
	QSOHistory = c.QSOHistory
	QSOLimit = c.QSOLimit
	TQSL = c.LogbookServices.TQSL
	EQSL = c.LogbookServices.EQSL
	ClubLog = c.LogbookServices.ClubLog
	QRZ = c.LogbookServices.QRZ
	HamAlert = c.Clusterservices.HamAlert

	return nil
}

// Write writes application configuration to the same file it was read from
func Write() error {
	// wrap
	c := configuration{
		Station:          Station,
		QSODatabase:      QSODatabase,
		UI:               UI,
		SourceFiles:      SourceFiles,
		ArchiveDirectory: ArchiveDirectory,
		QSLDelay:         QSLDelay,
		QSOHistory:       QSOHistory,
		QSOLimit:         QSOLimit,
		LogbookServices: logbookservices{
			TQSL:    TQSL,
			EQSL:    EQSL,
			ClubLog: ClubLog,
			QRZ:     QRZ,
		},
		Clusterservices: clusterservices{
			HamAlert: HamAlert,
		},
	}

	// create YAML to write from c
	b, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// write out to file
	err = ioutil.WriteFile(configFile, b, 0600)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

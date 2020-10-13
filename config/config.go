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

type qsotableview struct {
	History int
	Limit   int
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
	QSLDelay int
	TQSL     tqsl
	EQSL     eqsl
	ClubLog  clublog
	QRZ      qrz
}

type hamalert struct {
	HostPort string
	Username string
	Password string
}

type clusterservices struct {
	HamAlert hamalert
}

// Configuration is the application configuration that is serialized/deserialized to file
type Configuration struct {
	Station          station
	QSODatabase      qsodatabase
	QSOTableview     qsotableview
	UI               ui
	SourceFiles      []sourcefile
	LogbookServices  logbookservices
	ClusterServices  clusterservices
	WorkingDirectory string
}

func (c *Configuration) AddSourceFile(fname string) {
	if fname != "" {
		c.SourceFiles = append(c.SourceFiles, sourcefile{
			Location: fname,
			Offset:   0,
		})
	}
}

func (c *Configuration) RemoveSourceFile(idx int) {
	if idx >= 0 {
		c.SourceFiles = append(c.SourceFiles[:idx], c.SourceFiles[idx+1:]...)
	}
}

var (
	configFile string

	Station          station
	QSODatabase      qsodatabase
	QSOTableview     qsotableview
	UI               ui
	SourceFiles      []sourcefile
	LogbookServices  logbookservices
	ClusterServices  clusterservices
	WorkingDirectory string
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

	var c Configuration
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	Station = c.Station
	QSODatabase = c.QSODatabase
	QSOTableview = c.QSOTableview
	UI = c.UI
	SourceFiles = c.SourceFiles
	LogbookServices = c.LogbookServices
	ClusterServices = c.ClusterServices
	WorkingDirectory = c.WorkingDirectory

	return nil
}

// Write writes application configuration to the same file it was read from
func Write() error {
	c := Configuration{
		Station:          Station,
		QSODatabase:      QSODatabase,
		QSOTableview:     QSOTableview,
		UI:               UI,
		SourceFiles:      SourceFiles,
		LogbookServices:  LogbookServices,
		ClusterServices:  ClusterServices,
		WorkingDirectory: WorkingDirectory,
	}

	// create YAML to write from Options
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

// Copy provides the caller a copy of the current configuration
func Copy(c *Configuration) error {
	// make sure file is current config
	err := Write()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// reread from file
	// #nosec G304
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// into callers struct
	err = yaml.Unmarshal(bytes, c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// Reload current configuration from caller provided configuration
func Reload(c Configuration) error {
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

	// reread from file
	err = Read(configFile)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

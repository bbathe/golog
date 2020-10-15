package config

import (
	"fmt"
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

// Validate tests the required clublog fields
// doesn't log errors to cut down on log noise
func (c *clublog) Validate() error {
	missingField := "required configuration missing %s"

	if c.Email == "" {
		err := fmt.Errorf(missingField, "Username")
		return err
	}
	if c.Password == "" {
		err := fmt.Errorf(missingField, "Password")
		return err
	}
	if c.Callsign == "" {
		err := fmt.Errorf(missingField, "Callsign")
		return err
	}
	if c.APIKey == "" {
		err := fmt.Errorf(missingField, "APIKey")
		return err
	}

	return nil
}

type eqsl struct {
	Username string
	Password string
}

// Validate tests the required eqsl fields
// doesn't log errors to cut down on log noise
func (e *eqsl) Validate() error {
	missingField := "required configuration missing %s"

	if e.Username == "" {
		err := fmt.Errorf(missingField, "Username")
		return err
	}
	if e.Password == "" {
		err := fmt.Errorf(missingField, "Password")
		return err
	}

	return nil
}

type qrz struct {
	APIKey string
}

// Validate tests the required qrz fields
// doesn't log errors to cut down on log noise
func (q *qrz) Validate() error {
	missingField := "required configuration missing %s"

	if q.APIKey == "" {
		err := fmt.Errorf(missingField, "APIKey")
		return err
	}

	return nil
}

type tqsl struct {
	ExeLocation         string
	StationLocationName string
}

// Validate tests the required tqsl fields
// doesn't log errors to cut down on log noise
func (t *tqsl) Validate() error {
	missingField := "required configuration missing %s"

	if t.ExeLocation == "" {
		err := fmt.Errorf(missingField, "ExeLocation")
		return err
	}
	if t.StationLocationName == "" {
		err := fmt.Errorf(missingField, "StationLocationName")
		return err
	}

	return nil
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

// Validate tests the required hamalert fields
// doesn't log errors to cut down on log noise
func (h *hamalert) Validate() error {
	missingField := "required configuration missing %s"

	if h.HostPort == "" {
		err := fmt.Errorf(missingField, "HostPort")
		return err
	}
	if h.Username == "" {
		err := fmt.Errorf(missingField, "Username")
		return err
	}
	if h.Password == "" {
		err := fmt.Errorf(missingField, "Password")
		return err
	}

	return nil
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

// Validate tests the required Configuration fields
func (c *Configuration) Validate() error {
	missingField := "required configuration missing %s"

	if c.Station.Callsign == "" {
		err := fmt.Errorf(missingField, "Station Callsign")
		log.Printf("%+v", err)
		return err
	}
	if c.QSODatabase.Location == "" {
		err := fmt.Errorf(missingField, "Database")
		log.Printf("%+v", err)
		return err
	}
	if c.QSOTableview.History == 0 {
		err := fmt.Errorf(missingField, "QSO Tableview History")
		log.Printf("%+v", err)
		return err
	}
	if c.QSOTableview.Limit == 0 {
		err := fmt.Errorf(missingField, "QSO Tableview Limit")
		log.Printf("%+v", err)
		return err
	}
	if c.WorkingDirectory == "" {
		err := fmt.Errorf(missingField, "Working Directory")
		log.Printf("%+v", err)
		return err
	}

	return nil
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

	// make sure valid before proceeding
	err = c.Validate()
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

	// make sure valid before proceeding
	err := c.Validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
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
	if configFile == "" {
		// no current config
		return nil
	}

	// test if active configuration is valid for writing
	ac := &Configuration{
		Station:          Station,
		QSODatabase:      QSODatabase,
		QSOTableview:     QSOTableview,
		UI:               UI,
		SourceFiles:      SourceFiles,
		LogbookServices:  LogbookServices,
		ClusterServices:  ClusterServices,
		WorkingDirectory: WorkingDirectory,
	}
	err := ac.Validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// make sure file is current config
	err = Write()
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
	// make sure valid before proceeding
	err := c.Validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
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

	// reread from file
	err = Read(configFile)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

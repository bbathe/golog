package config

import (
	"errors"
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
// doesn't log errors because you don't have to use clublog
func (c *clublog) Validate() error {
	if c.Email == "" {
		err := fmt.Errorf(msgMissingField, "Club Log Username")
		return err
	}
	if c.Password == "" {
		err := fmt.Errorf(msgMissingField, "Club Log Password")
		return err
	}
	if c.Callsign == "" {
		err := fmt.Errorf(msgMissingField, "Club Log Callsign")
		return err
	}
	if c.APIKey == "" {
		err := fmt.Errorf(msgMissingField, "Club Log APIKey")
		return err
	}

	return nil
}

type eqsl struct {
	Username string
	Password string
}

// Validate tests the required eqsl fields
// doesn't log errors because you don't have to use eqsl
func (e *eqsl) Validate() error {
	if e.Username == "" {
		err := fmt.Errorf(msgMissingField, "eQSL Username")
		return err
	}
	if e.Password == "" {
		err := fmt.Errorf(msgMissingField, "eQSL Password")
		return err
	}

	return nil
}

type qrz struct {
	APIKey string
}

// Validate tests the required qrz fields
// doesn't log errors because you don't have to use qrz
func (q *qrz) Validate() error {
	if q.APIKey == "" {
		err := fmt.Errorf(msgMissingField, "QRZ API Key")
		return err
	}

	return nil
}

type tqsl struct {
	ExeLocation         string
	StationLocationName string
}

// Validate tests the required tqsl fields
// doesn't log errors because you don't have to use lotw
func (t *tqsl) Validate() error {
	if t.ExeLocation == "" {
		err := fmt.Errorf(msgMissingField, "LoTW TQSL Executable")
		return err
	}
	if t.StationLocationName == "" {
		err := fmt.Errorf(msgMissingField, "LoTW Station Location Name")
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
// doesn't log errors because you don't have to use hamalert
func (h *hamalert) Validate() error {
	if h.HostPort == "" {
		err := fmt.Errorf(msgMissingField, "HamAlert Host:Port")
		return err
	}
	if h.Username == "" {
		err := fmt.Errorf(msgMissingField, "HamAlert Username")
		return err
	}
	if h.Password == "" {
		err := fmt.Errorf(msgMissingField, "HamAlert Password")
		return err
	}

	return nil
}

type clusterservices struct {
	FlashWindowOnNewSpots bool
	HamAlert              hamalert
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
	BackupDirectory  string
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
	if c.Station.Callsign == "" {
		err := fmt.Errorf(msgMissingField, "Station Callsign")
		log.Printf("%+v", err)
		return err
	}
	if c.QSODatabase.Location == "" {
		err := fmt.Errorf(msgMissingField, "Database")
		log.Printf("%+v", err)
		return err
	}
	if c.QSOTableview.History == 0 {
		err := fmt.Errorf(msgMissingField, "QSO Tableview History")
		log.Printf("%+v", err)
		return err
	}
	if c.QSOTableview.Limit == 0 {
		err := fmt.Errorf(msgMissingField, "QSO Tableview Limit")
		log.Printf("%+v", err)
		return err
	}
	if c.WorkingDirectory == "" {
		err := fmt.Errorf(msgMissingField, "Working Directory")
		log.Printf("%+v", err)
		return err
	}
	if c.BackupDirectory == "" {
		err := fmt.Errorf(msgMissingField, "Backup Directory")
		log.Printf("%+v", err)
		return err
	}

	return nil
}

var (
	configFile      string
	errNoConfig     = errors.New("no current configuration file")
	msgMissingField = "required configuration missing %s"

	// unwrapped config values
	Station          station
	QSODatabase      qsodatabase
	QSOTableview     qsotableview
	UI               ui
	SourceFiles      []sourcefile
	LogbookServices  logbookservices
	ClusterServices  clusterservices
	WorkingDirectory string
	BackupDirectory  string
)

// Read loads application configuration from file fname
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

	// make sure valid before unwrapping
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
	BackupDirectory = c.BackupDirectory

	return nil
}

// Write writes application configuration to the same file it was read from
func Write() error {
	if configFile == "" {
		return errNoConfig
	}

	// wrap
	c := Configuration{
		Station:          Station,
		QSODatabase:      QSODatabase,
		QSOTableview:     QSOTableview,
		UI:               UI,
		SourceFiles:      SourceFiles,
		LogbookServices:  LogbookServices,
		ClusterServices:  ClusterServices,
		WorkingDirectory: WorkingDirectory,
		BackupDirectory:  BackupDirectory,
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
		return errNoConfig
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
		BackupDirectory:  BackupDirectory,
	}
	err := ac.Validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// make sure config file content matches current config
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
	if configFile == "" {
		return errNoConfig
	}

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

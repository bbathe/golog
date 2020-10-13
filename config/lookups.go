package config

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type adifband struct {
	XMLName  xml.Name `xml:"band"`
	AdifBand string   `xml:",chardata"`
	Low      string   `xml:"low,attr"`
	High     string   `xml:"high,attr"`
}

type adifbandmap struct {
	XMLName xml.Name   `xml:"bands"`
	Bands   []adifband `xml:"band"`
}

type adifmode struct {
	XMLName     xml.Name `xml:"adifmode"`
	AdifMode    string   `xml:"adif-mode,attr"`
	AdifSubmode string   `xml:"adif-submode,attr"`
}

type adifmodemap struct {
	XMLName   xml.Name   `xml:"adifmap"`
	Adifmodes []adifmode `xml:"adifmode"`
}

type tqslconfig struct {
	XMLName xml.Name    `xml:"tqslconfig"`
	Adifmap adifmodemap `xml:"adifmap"`
	Bandmap adifbandmap `xml:"bands"`
}

type band struct {
	Band     string
	FreqLow  int
	FreqHigh int
	Visible  bool
}

type mode struct {
	Mode    string
	Submode string
	Visible bool
}

type lookups struct {
	Bands []band
	Modes []mode
}

var (
	Bands []band
	Modes []mode
)

// LookupModeSubmode returns the mode & submode based on the mode name mode
func LookupModeSubmode(mode string) (string, string) {
	// try to match on submode first
	for _, m := range Modes {
		if m.Submode == mode {
			return m.Mode, m.Submode
		}
	}

	// see if it is really just mode
	for _, m := range Modes {
		if m.Mode == mode {
			return m.Mode, m.Submode
		}
	}

	// no match
	return "", ""
}

// LookupMode returns the mode name that matches the mode & submode
func LookupMode(mode, submode string) string {
	// for submode/mode we just keep submode if it available
	if submode != "" {
		return submode
	}
	return mode
}

// LookupBand returns the band name for teh frequency passed
func LookupBand(frequency int) string {
	for _, b := range Bands {
		if b.FreqHigh >= frequency && b.FreqLow <= frequency {
			return b.Band
		}
	}

	// no match
	return ""
}

// ListBandNames returns a list of the bands for displaying to the user
func ListBandNames() []string {
	bands := make([]string, 0, len(Bands))
	for _, b := range Bands {
		if b.Visible {
			bands = append(bands, b.Band)
		}
	}
	return bands
}

// ListModeNames returns a list of the modes for displaying to the user
func ListModeNames() []string {
	modes := make([]string, 0, len(Modes))
	for _, m := range Modes {
		n := m.Mode
		if m.Submode != "" {
			if m.Visible {
				n = m.Submode
			}
		}
		modes = append(modes, n)
	}
	return modes
}

// UpdateLookupsFromTQSL regenerates the lookuips based on the TQSL config.xml file
func UpdateLookupsFromTQSL(tqslconfigxml string) error {
	// #nosec G304
	bs, err := ioutil.ReadFile(tqslconfigxml)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// parse xml
	var tqslconf tqslconfig
	err = xml.Unmarshal(bs, &tqslconf)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// transform to our lookups
	bands := make([]band, 0, len(tqslconf.Bandmap.Bands))
	for _, b := range tqslconf.Bandmap.Bands {
		low, err := strconv.Atoi(b.Low)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		high, err := strconv.Atoi(b.High)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		bands = append(bands, band{
			Band:     strings.ToLower(b.AdifBand),
			FreqLow:  low,
			FreqHigh: high,
			Visible:  true,
		})
	}

	modes := make([]mode, 0, len(tqslconf.Adifmap.Adifmodes))
	for _, m := range tqslconf.Adifmap.Adifmodes {
		modes = append(modes, mode{
			Mode:    strings.ToUpper(m.AdifMode),
			Submode: strings.ToUpper(m.AdifSubmode),
			Visible: true,
		})
	}

	Bands = bands
	Modes = modes

	return nil
}

// ReadLookupsFromFile reads lookups from file fname
func ReadLookupsFromFile(fname string) error {
	// #nosec G304
	bs, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	var l lookups
	err = yaml.Unmarshal(bs, &l)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// unwrap
	Bands = l.Bands
	Modes = l.Modes

	return nil
}

// WriteLookupsToFile writes lookups to file fname
func WriteLookupsToFile(fname string) error {
	l := lookups{
		Bands: Bands,
		Modes: Modes,
	}

	// create YAML to write from lookups
	b, err := yaml.Marshal(l)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// write out to file
	err = ioutil.WriteFile(fname, b, 0600)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

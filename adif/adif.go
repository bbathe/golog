package adif

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/models/qso"
)

var (
	// regex's to extract data fields
	reStationCallsign = regexp.MustCompile(`(?i)^station_callsign:(\d+)>(.+)`)
	reBand            = regexp.MustCompile(`(?i)^band:(\d+)>(.+)`)
	reCall            = regexp.MustCompile(`(?i)^call:(\d+)>(.+)`)
	reMode            = regexp.MustCompile(`(?i)^mode:(\d+)>(.+)`)
	reSubmode         = regexp.MustCompile(`(?i)^submode:(\d+)>(.+)`)
	reDate            = regexp.MustCompile(`(?i)^qso_date:(\d+)>(.+)`)
	reTime            = regexp.MustCompile(`(?i)^time_on:(\d+)>(.+)`)
	reRSTRcvd         = regexp.MustCompile(`(?i)^rst_rcvd:(\d+)>(.+)`)
	reRSTSent         = regexp.MustCompile(`(?i)^rst_sent:(\d+)>(.+)`)
)

// extractValue does the work to pull the value out of the ADIF field
func extractValue(s string, r *regexp.Regexp) *string {
	m := r.FindStringSubmatch(s)
	if m != nil {
		// get field length
		i, err := strconv.Atoi(m[1])
		if err != nil {
			log.Printf("%+v", err)
			return nil
		}

		// get field value
		if i <= len(m[2]) {
			ss := m[2][:i]
			return &ss
		}
	}
	return nil
}

func QSOFromADIFRecord(record string) (*qso.QSO, error) {
	var err error
	var qso qso.QSO
	submode := ""

	// look at every field, picking out what we want
	fields := strings.Split(record, "<")
	for _, field := range fields {
		m := extractValue(field, reStationCallsign)
		if m != nil {
			qso.StationCallsign = strings.ToUpper(*m)
			continue
		}
		m = extractValue(field, reBand)
		if m != nil {
			qso.Band = strings.ToLower(*m)
			continue
		}
		m = extractValue(field, reCall)
		if m != nil {
			qso.Call = strings.ToUpper(*m)
			continue
		}
		m = extractValue(field, reMode)
		if m != nil {
			qso.Mode = strings.ToUpper(*m)
			continue
		}
		m = extractValue(field, reSubmode)
		if m != nil {
			submode = strings.ToUpper(*m)
			continue
		}
		m = extractValue(field, reDate)
		if m != nil {
			t, err := time.Parse("20060102", *m)
			if err != nil {
				log.Printf("%+v", err)
				continue
			}

			qso.Date = t.Format("2006-01-02")
			continue
		}
		m = extractValue(field, reTime)
		if m != nil {
			var t time.Time

			if len(*m) > 4 {
				t, err = time.Parse("150405", *m)
				if err != nil {
					log.Printf("%+v", err)
					continue
				}
			} else {
				t, err = time.Parse("1504", *m)
				if err != nil {
					log.Printf("%+v", err)
					continue
				}
			}

			// we don't keep seconds
			qso.Time = t.Format("15:04")
			continue
		}
		m = extractValue(field, reRSTRcvd)
		if m != nil {
			qso.RSTRcvd = strings.ToUpper(*m)
			continue
		}
		m = extractValue(field, reRSTSent)
		if m != nil {
			qso.RSTSent = strings.ToUpper(*m)
			continue
		}
	}

	// fixup mode/submode
	qso.Mode = config.LookupMode(qso.Mode, submode)

	return &qso, nil
}

// ReadFromFile reads QSOs from the ADIF file fname
func ReadFromFile(fname string, qsllotw, qsleqsl, qslqrz, qslclublog qso.QSLSent) ([]qso.QSO, error) {
	loadedAt := time.Now().Unix()

	file, err := os.Open(fname)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	defer file.Close()

	// we split on words, reconstruct the record and process record-by-record
	// this is to handle multiline records, etc.
	// for the fields we care about there are no embedded spaces so good enough for our purposes
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	var record string
	qsos := make([]qso.QSO, 0, 128)
	for scanner.Scan() {
		t := scanner.Text()
		record += t

		if strings.HasSuffix(strings.ToLower(t), "<eoh>") {
			record = ""
			continue
		}

		if strings.HasSuffix(strings.ToLower(t), "<eor>") {
			// get qso from record
			var qso *qso.QSO
			qso, err = QSOFromADIFRecord(record)
			if err != nil {
				log.Println(record)
				log.Printf("%+v", err)
				record = ""
				continue
			}

			// override/default some fields
			qso.LoadedAt = loadedAt
			qso.QSLLotw = qsllotw
			qso.QSLEqsl = qsleqsl
			qso.QSLQrz = qslqrz
			qso.QSLClublog = qslclublog
			if qso.StationCallsign == "" {
				qso.StationCallsign = config.Station.Callsign
			}

			// make sure all is good
			err = qso.Validate(false)
			if err != nil {
				log.Println(record)
				log.Printf("%+v", err)
			} else {
				qsos = append(qsos, *qso)
			}

			record = ""
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	return qsos, nil
}

// QSOToADIFRecord returns the adif record from the qso
func QSOToADIFRecord(qso qso.QSO) (string, error) {
	// format date
	t, err := time.Parse("2006-01-02", qso.Date)
	if err != nil {
		log.Printf("%+v", err)
		return "", err
	}
	qsodate := t.Format("20060102")

	// format time
	t, err = time.Parse("15:04", qso.Time)
	if err != nil {
		log.Printf("%+v", err)
		return "", err
	}
	qsotime := t.Format("1504")

	// handle optional fields
	var rstrcvd string
	if qso.RSTRcvd != "" {
		rstrcvd = fmt.Sprintf("<rst_rcvd:%d>%s", len(qso.RSTRcvd), qso.RSTRcvd)
	}

	var rstsent string
	if qso.RSTSent != "" {
		rstsent = fmt.Sprintf("<rst_sent:%d>%s", len(qso.RSTSent), qso.RSTSent)
	}

	var submode string
	mode, s := config.LookupModeSubmode(qso.Mode)
	if s != "" {
		submode = fmt.Sprintf("<submode:%d>%s", len(s), s)
	}

	return fmt.Sprintf(
		"<station_callsign:%d>%s<call:%d>%s<band:%d>%s<mode:%d>%s%s<qso_date:%d>%s<time_on:%d>%s%s%s<eor>\n",
		len(qso.StationCallsign),
		qso.StationCallsign,
		len(qso.Call),
		qso.Call,
		len(qso.Band),
		qso.Band,
		len(mode),
		mode,
		submode,
		len(qsodate),
		qsodate,
		len(qsotime),
		qsotime,
		rstrcvd,
		rstsent,
	), nil
}

// WriteToFile creates an ADIF file with all qsos
func WriteToFile(qsos []qso.QSO, fname string) error {
	// create file
	f, err := os.Create(fname)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	for _, qso := range qsos {
		// qso to adif record
		s, err := QSOToADIFRecord(qso)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		// write record
		_, err = w.WriteString(s)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	err = w.Flush()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

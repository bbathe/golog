package db

import (
	"log"
	"os"
	"path/filepath"

	"github.com/bbathe/golog/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

var QSODb *sqlx.DB

// OpenQSODb creates the connection to the qso database
func OpenQSODb() error {
	var err error

	// if already open, close it
	if QSODb != nil {
		err = CloseQSODb()
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	// if config not set, make something up
	if config.QSODatabase.Location == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		config.QSODatabase.Location = filepath.Join(cwd, "golog.db")
	}

	// test if db file exists
	_, err = os.Stat(config.QSODatabase.Location)
	if err != nil {
		if os.IsNotExist(err) {
			err = NewQSODb()
			if err != nil {
				log.Printf("%+v", err)
				return err
			}

			// NewQSODb leaves database open
			return nil
		}
	}

	// make connection
	QSODb, err = sqlx.Connect("sqlite3", config.QSODatabase.Location+"?_foreign_keys=true&_loc=UTC")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// CloseQSODb closes the connection to the qso database
func CloseQSODb() error {
	if QSODb != nil {
		err := QSODb.Close()
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		QSODb = nil
	}

	return nil
}

// NewQSODb initializes a new qso database
func NewQSODb() error {
	err := CloseQSODb()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// (re)create the database
	err = os.Remove(config.QSODatabase.Location)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("%+v", err)
		return err
	}
	d, err := os.Create(config.QSODatabase.Location)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}
	d.Close()

	// open new database
	err = OpenQSODb()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// create qso table
	_, err = QSODb.Exec(`
		create table qsos (
			id integer primary key asc not null,
			loaded_at integer not null,
			station_callsign text not null,
			band text not null,
			call text not null,
			mode text not null,
			qso_date text not null,
			qso_time text not null,
			rst_rcvd text null,
			rst_sent text null,
			qsl_lotw integer not null default 0 check (qsl_lotw in (0, 1)),
			qsl_eqsl integer not null default 0 check (qsl_eqsl in (0, 1)),
			qsl_qrz integer not null default 0 check (qsl_qrz in (0, 1)),
			qsl_clublog integer not null default 0 check (qsl_clublog in (0, 1))
		)
	`)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// index for what uniquely defines a qso
	_, err = QSODb.Exec(`
		create unique index unique_qso on qsos(station_callsign, band, call, mode, qso_date, qso_time)
	`)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// index for quering by loaded date/time
	_, err = QSODb.Exec(`
		create index loaded_at on qsos(loaded_at)
	`)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

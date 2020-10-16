package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

var SpotDb *sqlx.DB

// OpenSpotDb creates the connection to the spot database
func OpenSpotDb() error {
	var err error

	// if already open, close it
	if SpotDb != nil {
		err = CloseSpotDb()
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	// make connection
	SpotDb, err = sqlx.Connect("sqlite3", "file:spot.db?mode=memory&cache=shared&_foreign_keys=true&_loc=UTC")
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// CloseSpotDb closes the connection to the spot database
func CloseSpotDb() error {
	if SpotDb != nil {
		err := SpotDb.Close()
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
		SpotDb = nil
	}

	return nil
}

// NewSpotDb initializes a new spot database
func NewSpotDb() error {
	err := CloseSpotDb()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// open new database
	err = OpenSpotDb()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// create spots table
	_, err = SpotDb.Exec(`
		create table spots (
			id integer primary key asc not null,
			timestamp text null,
			call text null,
			band text null,
			frequency text null,
			comments text null
		)
	`)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// index for what uniquely defines a spot
	_, err = SpotDb.Exec(`
		create unique index unique_spots on spots(timestamp, call, band)
	`)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

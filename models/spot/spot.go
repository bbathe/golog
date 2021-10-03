package spot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/db"
	"github.com/bbathe/golog/util"
)

type Spot struct {
	ID        int64  `db:"id"`
	Timestamp string `db:"timestamp"`
	Call      string `db:"call"`
	Band      string `db:"band"`
	Frequency string `db:"frequency"`
	Comments  string `db:"comments"`
	Spotter   string `db:"spotter"`
}

const (
	stmtSpotInsert = `
		insert into spots (
			timestamp,
			call,
			band,
			frequency,
			comments,
			spotter
		) values (
			:timestamp,
			:call,
			:band,
			:frequency,
			:comments,
			:spotter
		)
		on conflict(timestamp, call, band, spotter) do nothing
	`

	stmtSpotSelectAllAfter = `
		select
			id,
			timestamp,
			call,
			band,
			frequency,
			comments,
			spotter
		from
			spots
		where
			id > :id
	`
)

var (
	errNoConnection = fmt.Errorf("no database connection")

	handlers []SpotChangeEventHandler
)

// allow callers to register to recieve event after any spot changes occur
type SpotChangeEventHandler func()

func Attach(handler SpotChangeEventHandler) int {
	handlers = append(handlers, handler)

	return len(handlers) - 1
}

func Detach(handle int) {
	handlers[handle] = nil
}

func publishSpotChange() {
	for _, h := range handlers {
		h()
	}
}

// Add inserts a single spot into the spot database
func Add(timestamp, call, frequency, comments, spotter string) error {
	var err error

	if db.SpotDb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return err
	}

	// make timestamps look how we want
	t, err := time.Parse("1504", timestamp)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// get frequency as a float
	freq, err := strconv.ParseFloat(frequency, 64)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// add spot
	spot := Spot{
		Timestamp: t.Format("15:04"),
		Call:      strings.ToUpper(call),
		Band:      config.LookupBand(int(freq)),
		Frequency: util.FormatFrequency(frequency),
		Comments:  strings.TrimSpace(comments),
		Spotter:   strings.ToUpper(spotter),
	}

	spotInsert, err := db.SpotDb.PrepareNamed(stmtSpotInsert)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	_, err = spotInsert.Exec(spot)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	publishSpotChange()
	return nil
}

// NewSpots returns all Spots with ID after lastID
func NewSpots(lastID int64) ([]Spot, error) {
	var err error

	if db.SpotDb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return []Spot{}, err
	}

	params := map[string]interface{}{
		"id": lastID,
	}

	q, err := db.SpotDb.PrepareNamed(stmtSpotSelectAllAfter)
	if err != nil {
		log.Printf("%+v", err)
		return []Spot{}, err
	}

	var spots []Spot
	err = q.Select(&spots, params)
	if err != nil {
		log.Printf("%+v", err)
		return []Spot{}, err
	}

	return spots, nil
}

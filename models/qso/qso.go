package qso

import (
	"fmt"
	"log"
	"time"

	"github.com/bbathe/golog/db"
)

type QSLSent int64

const (
	NotSent QSLSent = 0
	Sent    QSLSent = 1
)

type QSLService string

const (
	QSLLotw    = "qsl_lotw"
	QSLEqsl    = "qsl_eqsl"
	QSLQrz     = "qsl_qrz"
	QSLClublog = "qsl_clublog"
	QSLCard    = "qsl_card"
)

type QSO struct {
	ID              int64  `db:"id"`
	LoadedAt        int64  `db:"loaded_at"`
	StationCallsign string `db:"station_callsign"`

	Band    string `db:"band"`
	Call    string `db:"call"`
	Mode    string `db:"mode"`
	Date    string `db:"qso_date"`
	Time    string `db:"qso_time"`
	RSTRcvd string `db:"rst_rcvd"`
	RSTSent string `db:"rst_sent"`

	QSLLotw    QSLSent `db:"qsl_lotw"`
	QSLEqsl    QSLSent `db:"qsl_eqsl"`
	QSLQrz     QSLSent `db:"qsl_qrz"`
	QSLClublog QSLSent `db:"qsl_clublog"`
	QSLCard    QSLSent `db:"qsl_card"`
}

const (
	stmtQSOInsert = `
		insert into qsos (
			loaded_at,
			station_callsign,
			band,
			call,
			mode,
			qso_date,
			qso_time,
			rst_rcvd,
			rst_sent,
			qsl_lotw,
			qsl_eqsl,
			qsl_qrz,
			qsl_clublog,
			qsl_card
		) values (
			:loaded_at,
			:station_callsign,
			:band,
			:call,
			:mode,
			:qso_date,
			:qso_time,
			:rst_rcvd,
			:rst_sent,
			:qsl_lotw,
			:qsl_eqsl,
			:qsl_qrz,
			:qsl_clublog,
			:qsl_card
		)
		on conflict(station_callsign, band, call, mode, qso_date, qso_time) do nothing
	`

	stmtQSOOnlyInsert = `
		insert into qsos (
			loaded_at,
			station_callsign,
			band,
			call,
			mode,
			qso_date,
			qso_time,
			rst_rcvd,
			rst_sent
		) values (
			:loaded_at,
			:station_callsign,
			:band,
			:call,
			:mode,
			:qso_date,
			:qso_time,
			:rst_rcvd,
			:rst_sent
		)
		on conflict(station_callsign, band, call, mode, qso_date, qso_time) do nothing
	`

	stmtQSOSelectAll = `
		select
			id,
			loaded_at,
			station_callsign,
			band,
			call,
			mode,
			qso_date,
			qso_time,
			rst_rcvd,
			rst_sent,
			qsl_lotw,
			qsl_eqsl,
			qsl_qrz,
			qsl_clublog,
			qsl_card
		from
			qsos
	`

	stmtQSOSelectDupTest = `
		select
			id
		from
			qsos
		where
			station_callsign = :station_callsign
			and band = :band
			and call = :call
			and mode = :mode
			and qso_date = :qso_date
			and qso_time = :qso_time
		`

	stmtQSODelete = `
		delete from qsos where id = :id;	
	`

	stmtQSOOnlyUpdate = `
		update qsos set
			station_callsign = :station_callsign,
			band = :band,
			call = :call,
			mode = :mode,
			qso_date = :qso_date,
			qso_time = :qso_time,
			rst_rcvd = :rst_rcvd,
			rst_sent = :rst_sent
		where
			id = :id
	`
)

var (
	errNoConnection = fmt.Errorf("no database connection")

	handlers []QSOChangeEventHandler
)

// allow callers to register to recieve event after any qso changes occur
type QSOChangeEventHandler func()

func Attach(handler QSOChangeEventHandler) int {
	handlers = append(handlers, handler)

	return len(handlers) - 1
}

func Detach(handle int) {
	handlers[handle] = nil
}

func publishQSOChange() {
	for _, h := range handlers {
		h()
	}
}

// Validate tests the required QSO fields
func (qso *QSO) Validate(checkID bool) error {
	missingField := "required field missing %s"

	if checkID && qso.ID == 0 {
		err := fmt.Errorf(missingField, "ID")
		log.Printf("%+v", err)
		return err
	}
	if qso.Date == "" {
		err := fmt.Errorf(missingField, "Date")
		log.Printf("%+v", err)
		return err
	}
	if qso.Time == "" {
		err := fmt.Errorf(missingField, "Time")
		log.Printf("%+v", err)
		return err
	}
	if qso.Call == "" {
		err := fmt.Errorf(missingField, "Call")
		log.Printf("%+v", err)
		return err
	}
	if qso.Band == "" {
		err := fmt.Errorf(missingField, "Band")
		log.Printf("%+v", err)
		return err
	}
	if qso.Mode == "" {
		err := fmt.Errorf(missingField, "Mode")
		log.Printf("%+v", err)
		return err
	}
	if qso.LoadedAt == 0 {
		err := fmt.Errorf(missingField, "LoadedAt")
		log.Printf("%+v", err)
		return err
	}
	if qso.StationCallsign == "" {
		err := fmt.Errorf(missingField, "StationCallsign")
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// Add inserts a single QSO into the qso database
// sets qso.LoadedAt before insert
// does not insert QSL fields, they default to false
func (qso *QSO) Add() error {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return err
	}

	// set LoadedAt to now
	qso.LoadedAt = time.Now().Unix()

	err = qso.Validate(false)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	qsoInsert, err := db.QSODb.PrepareNamed(stmtQSOOnlyInsert)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	_, err = qsoInsert.Exec(qso)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	publishQSOChange()
	return nil
}

// BulkAdd inserts all QSOs into the qso database
// assumes qso.LoadedAt was already set
func BulkAdd(qsos []QSO) error {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return err
	}

	// in a transaction
	tx := db.QSODb.MustBegin()
	defer func() {
		// if we've had an error, rollback
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				log.Printf("%+v", err)
			}
		}
	}()

	qsoInsert, err := tx.PrepareNamed(stmtQSOInsert)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// insert all qsos
	for _, qso := range qsos {
		err := qso.Validate(false)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}

		// create qso record
		_, err = qsoInsert.Exec(qso)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	publishQSOChange()
	return nil
}

// UpdateOnlyQSO updates the QSO only fields in the qso database
// does not update QSL fields
func (qso *QSO) UpdateOnlyQSO() error {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return err
	}

	err = qso.Validate(true)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// in a transaction
	tx := db.QSODb.MustBegin()
	defer func() {
		// if we've had an error, rollback
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				log.Printf("%+v", err)
			}
		}
	}()

	// see if there is already a qso matching this update
	q, err := tx.PrepareNamed(stmtQSOSelectDupTest)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	var qsos []QSO
	err = q.Select(&qsos, qso)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// if what we found is not this qso then error out
	if len(qsos) > 0 && qsos[0].ID != qso.ID {
		err = fmt.Errorf("update would cause duplicate qso")
		log.Printf("%+v", err)
		return err
	}

	// good to go with update
	q, err = tx.PrepareNamed(stmtQSOOnlyUpdate)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	_, err = q.Exec(qso)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	publishQSOChange()
	return nil
}

// Delete removes a qso from the qso database
func (qso *QSO) Delete() error {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return err
	}

	err = qso.Validate(true)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	q, err := db.QSODb.PrepareNamed(stmtQSODelete)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	_, err = q.Exec(qso)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	publishQSOChange()
	return nil
}

// All returns all QSOs in the qso database
func All() ([]QSO, error) {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	var qsos []QSO
	err = db.QSODb.Select(&qsos, stmtQSOSelectAll)
	if err != nil {
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	return qsos, nil
}

// Get returns a single QSO identified by ID
func Get(ID int64) (*QSO, error) {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return nil, err
	}

	// start with All and add where clause to get single qso
	stmt := stmtQSOSelectAll
	stmt += " where id = :id"

	q, err := db.QSODb.PrepareNamed(stmt)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	params := map[string]interface{}{
		"id": ID,
	}

	var qso []QSO
	err = q.Select(&qso, params)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	return &qso[0], nil
}

// Search returns all QSOs matching the criteria limited by count limit
func Search(criteria QSO, limit int) ([]QSO, error) {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	// query parameters
	params := map[string]interface{}{
		"limit": limit,
	}

	// start with All and add where clause dynamically
	stmt := stmtQSOSelectAll

	needAnd := false
	appendFieldCriteria := func(s, f, v string) string {
		if v != "" {
			if needAnd {
				s += " and "
			} else {
				s += " where "
			}
			s += fmt.Sprintf(" %s like :%s ", f, f)
			params[f] = v
			needAnd = true
		}

		return s
	}

	stmt = appendFieldCriteria(stmt, "qso_date", criteria.Date)
	stmt = appendFieldCriteria(stmt, "qso_time", criteria.Time)
	stmt = appendFieldCriteria(stmt, "call", criteria.Call)
	stmt = appendFieldCriteria(stmt, "band", criteria.Band)
	stmt = appendFieldCriteria(stmt, "mode", criteria.Mode)
	stmt = appendFieldCriteria(stmt, "rst_rcvd", criteria.RSTRcvd)
	stmt = appendFieldCriteria(stmt, "rst_sent", criteria.RSTSent)

	// limit result set and order so we get the latest qsos
	stmt += " order by qso_date desc, qso_time desc limit :limit"

	q, err := db.QSODb.PrepareNamed(stmt)
	if err != nil {
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	var qsos []QSO
	err = q.Select(&qsos, params)
	if err != nil {
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	return qsos, nil
}

// History returns all QSO loaded after days ago limited by count limit
func History(days, limit int) ([]QSO, error) {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	// query parameters
	params := map[string]interface{}{}

	// start with All and add where clause
	stmt := stmtQSOSelectAll

	// how much history?
	if days > 0 {
		now := time.Now()
		then := now.Add(-(time.Duration(days*24) * time.Hour)).Unix()
		stmt += " where loaded_at >= :loadedat"
		params["loadedat"] = then
	}

	// order so we get the latest qsos
	stmt += " order by qso_date desc, qso_time desc"

	// limit result set
	if limit > 0 {
		stmt += "  limit :limit"
		params["limit"] = limit
	}

	q, err := db.QSODb.PrepareNamed(stmt)
	if err != nil {
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	var qsos []QSO
	err = q.Select(&qsos, params)
	if err != nil {
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	return qsos, nil
}

// FindQSLsToSend returns all QSOs that need QSLs for a specific service before delay minutes ago
func FindQSLsToSend(service QSLService, delay int) ([]QSO, error) {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	// query parameters
	params := map[string]interface{}{}

	// start with All and add where clause
	stmt := stmtQSOSelectAll
	stmt += fmt.Sprintf(" where %s = 0", service)

	// handle delay for sending QSL
	if delay > 0 {
		now := time.Now()
		then := now.Add(-(time.Duration(delay) * time.Minute)).Unix()
		stmt += " and loaded_at <= :loadedat"
		params["loadedat"] = then
	}

	// return results to caller with oldest first
	stmt += " order by qso_date asc, qso_time asc"

	q, err := db.QSODb.PrepareNamed(stmt)
	if err != nil {
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	var qsos []QSO
	err = q.Select(&qsos, params)
	if err != nil {
		log.Printf("%+v", err)
		return []QSO{}, err
	}

	return qsos, nil
}

// UpdateQSLsToSent updates the QSOs as being sent to QSL service
func UpdateQSLsToSent(qsos []QSO, service QSLService) error {
	var err error

	if db.QSODb == nil {
		err = errNoConnection
		log.Printf("%+v", err)
		return err
	}

	// in a transaction
	tx := db.QSODb.MustBegin()
	defer func() {
		// if we've had an error, rollback
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				log.Printf("%+v", err)
			}
		}
	}()

	q, err := tx.PrepareNamed(fmt.Sprintf("update qsos set %s = 1 where id = :id", service))
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	for _, qso := range qsos {
		_, err = q.Exec(qso)
		if err != nil {
			log.Printf("%+v", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	publishQSOChange()
	return nil
}

package ui

import (
	"log"
	"sort"
	"strings"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/config"

	"github.com/bbathe/golog/models/qso"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

// QSOModel is used to display QSOs in the main windows TableView
type QSOModel struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn     int
	sortOrder      walk.SortOrder
	searchCriteria *qso.QSO
	items          []*qso.QSO
}

func NewQSOModel() *QSOModel {
	m := new(QSOModel)
	m.sortColumn = 1
	m.sortOrder = walk.SortDescending
	m.ResetRows()

	// register event handler for any qso changes
	qso.Attach(func() {
		qsomodel.ResetRows()
	})

	return m
}

// RowCount is called by the TableView from SetModel and every time the model publishes a RowsReset event
func (m *QSOModel) RowCount() int {
	return len(m.items)
}

// Value is called by the TableView when it needs the text to display for a given cell
func (m *QSOModel) Value(row, col int) interface{} {
	item := m.items[row]

	switch col {
	case 0:
		return item.Date

	case 1:
		return item.Time

	case 2:
		return item.Call

	case 3:
		return item.Band

	case 4:
		return item.Mode

	case 5:
		return item.RSTRcvd

	case 6:
		return item.RSTSent

	case 7:
		if item.QSLLotw == qso.Sent {
			return "\u2713"
		} else {
			return ""
		}

	case 8:
		if item.QSLQrz == qso.Sent {
			return "\u2713"
		} else {
			return ""
		}

	case 9:
		if item.QSLClublog == qso.Sent {
			return "\u2713"
		} else {
			return ""
		}

	case 10:
		if item.QSLCard == qso.Sent {
			return "\u2713"
		} else {
			return ""
		}
	}

	return ""
}

// Sort is called by the TableView to sort the model
func (m *QSOModel) Sort(col int, order walk.SortOrder) error {
	m.sortColumn, m.sortOrder = col, order

	sort.SliceStable(m.items, func(i, j int) bool {
		a, b := m.items[i], m.items[j]

		c := func(ls bool) bool {
			if m.sortOrder == walk.SortAscending {
				return ls
			}
			return !ls
		}

		switch m.sortColumn {
		case 0:
			// sort by date then time
			if strings.Compare(a.Date, b.Date) == 0 {
				return c(strings.Compare(a.Time, b.Time) < 0)
			} else {
				return c(strings.Compare(a.Date, b.Date) < 0)
			}

		case 1:
			return c(strings.Compare(a.Time, b.Time) < 0)

		case 2:
			return c(strings.Compare(a.Call, b.Call) < 0)

		case 3:
			return c(strings.Compare(a.Band, b.Band) < 0)

		case 4:
			return c(strings.Compare(a.Mode, b.Mode) < 0)

		case 5:
			return c(strings.Compare(a.RSTRcvd, b.RSTRcvd) < 0)

		case 6:
			return c(strings.Compare(a.RSTSent, b.RSTSent) < 0)

		case 7:
			return c(a.QSLLotw > b.QSLLotw)

		case 8:
			return c(a.QSLQrz > b.QSLQrz)

		case 9:
			return c(a.QSLClublog > b.QSLClublog)

		case 10:
			return c(a.QSLCard > b.QSLCard)
		}

		return false
	})

	return m.SorterBase.Sort(col, order)
}

// ResetRows loads QSOs from the database
func (m *QSOModel) ResetRows() {
	var r []qso.QSO
	var err error

	// search or not based on if criteria is set in model
	if m.searchCriteria != nil {
		r, err = qso.Search(*m.searchCriteria, config.QSOTableview.Limit)
	} else {
		r, err = qso.History(config.QSOTableview.History, config.QSOTableview.Limit)
	}
	if err != nil {
		log.Printf("%+v", err)
		MsgError(nil, err)
		return
	}

	m.items = make([]*qso.QSO, len(r))
	for i := range m.items {
		m.items[i] = &r[i]
	}

	// notify TableView about the reset
	m.PublishRowsReset()

	// maintain same sorting
	err = m.Sort(m.sortColumn, m.sortOrder)
	if err != nil {
		log.Printf("%+v", err)
		MsgError(nil, err)
		return
	}
}

// Search establishes the selection criteria in the model
func (m *QSOModel) Search(date, time, call, band, mode, rstrcvd, rstsent string) {
	m.searchCriteria = &qso.QSO{
		Date:    strings.TrimSpace(date),
		Time:    strings.TrimSpace(time),
		Call:    strings.TrimSpace(call),
		Band:    strings.TrimSpace(band),
		Mode:    strings.TrimSpace(mode),
		RSTRcvd: strings.TrimSpace(rstrcvd),
		RSTSent: strings.TrimSpace(rstsent),
	}

	m.ResetRows()
}

// ClearSearch clears the selection criteria in the model
func (m *QSOModel) ClearSearch() {
	m.searchCriteria = nil

	m.ResetRows()
}

// Export generates an adif with the items in the model
func (m *QSOModel) Export() {
	// ask where to export to
	fname, err := SaveFilePicker(mainWin, "Select file to export QSOs", "ADIF Files (*.adi;*.adif)|*.adi;*.adif|All Files (*.*)|*.*")
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	if fname != nil {
		// get qsos ready for WriteToFile
		qs := make([]qso.QSO, 0, len(m.items))
		for _, item := range m.items {
			qs = append(qs, *item)
		}

		// write to file
		err := adif.WriteToFile(qs, *fname)
		if err != nil {
			log.Printf("%+v", err)
			MsgError(nil, err)
			return
		}
	}
}

// qsoTableView returns the QSO TableView to be included in the apps MainWindow
func qsoTableView() declarative.TableView {
	var tv *walk.TableView

	qsomodel = NewQSOModel()

	return declarative.TableView{
		AssignTo:            &tv,
		AlternatingRowBG:    true,
		CustomHeaderHeight:  30,
		LastColumnStretched: true,
		ColumnsOrderable:    false,
		ContextMenuItems: []declarative.MenuItem{
			declarative.Action{
				Text: "QRZ page",
				OnTriggered: func() {
					idx := tv.CurrentIndex()
					if idx >= 0 {
						err := launchQRZPage(qsomodel.items[idx].Call)
						if err != nil {
							MsgError(mainWin, err)
							log.Printf("%+v", err)
							return
						}
					}
				},
			},
			declarative.Action{
				Text: "QSOs with call",
				OnTriggered: func() {
					idx := tv.CurrentIndex()
					if idx >= 0 {
						// copy call to new qso
						*selectedQSO = qso.QSO{
							Call: qsomodel.items[idx].Call,
						}

						// refresh
						err := bndSelectedQSO.Reset()
						if err != nil {
							MsgError(mainWin, err)
							log.Printf("%+v", err)
							return
						}

						// search only on call
						qsomodel.Search(
							"",
							"",
							qsomodel.items[idx].Call,
							"",
							"",
							"",
							"",
						)
					}
				},
			},
			declarative.Action{
				Text: "QSL card",
				OnTriggered: func() {
					idx := tv.CurrentIndex()
					if idx >= 0 {
						sent := qso.Sent
						if qsomodel.items[idx].QSLCard == qso.Sent {
							sent = qso.NotSent
						}
						err := qso.UpdateQSL([]qso.QSO{*qsomodel.items[idx]}, qso.QSLCard, sent)
						if err != nil {
							MsgError(mainWin, err)
							log.Printf("%+v", err)
							return
						}
					}
				},
			},
			declarative.Action{
				Text: "Copy call",
				OnTriggered: func() {
					idx := tv.CurrentIndex()
					if idx >= 0 {
						err := copyToClipboard(qsomodel.items[idx].Call)
						if err != nil {
							MsgError(mainWin, err)
							log.Printf("%+v", err)
							return
						}
					}
				},
			},
		},
		Columns: []declarative.TableViewColumn{
			{Title: "Date"},
			{Title: "Time", Width: 85},
			{Title: "Callsign"},
			{Title: "Band", Width: 85},
			{Title: "Mode", Width: 85},
			{Title: "RST Rcvd", Width: 85},
			{Title: "RST Sent", Width: 85},
			{Title: "QSL LoTW", Width: 85},
			{Title: "QSL QRZ", Width: 85},
			{Title: "QSL Club Log", Width: 125},
			{Title: "QSL Card", Width: 85},
			{Title: ""},
		},
		Model: qsomodel,
		OnItemActivated: func() {
			idx := tv.CurrentIndex()
			if idx >= 0 {
				// copy activated qso to selected qso
				*selectedQSO = *qsomodel.items[idx]

				// refresh
				err := bndSelectedQSO.Reset()
				if err != nil {
					MsgError(mainWin, err)
					log.Printf("%+v", err)
					return
				}
			}
		},
		StyleCell: func(style *walk.CellStyle) {
			drawCellStyles(tv, style, qsomodel.SorterBase)
		},
	}
}

package ui

import (
	"log"
	"sort"

	"github.com/bbathe/golog/models/spot"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

var (
	dxclustermodel *DXClusterModel
)

// DXClusterModel is used to display cluster spots in the main windows TableView
type DXClusterModel struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn int
	sortOrder  walk.SortOrder
	items      []*spot.Spot
	lastID     int64
}

func NewDXClusterModel() *DXClusterModel {
	m := new(DXClusterModel)
	m.sortColumn = 99
	m.sortOrder = walk.SortAscending
	m.ResetRows()

	// register event handler for any spot changes
	spot.Attach(func() {
		dxclustermodel.ResetRows()
	})

	return m
}

// RowCount is called by the TableView from SetModel and every time the model publishes a RowsReset event
func (m *DXClusterModel) RowCount() int {
	return len(m.items)
}

// Value is called by the TableView when it needs the text to display for a given cell
func (m *DXClusterModel) Value(row, col int) interface{} {
	item := m.items[row]

	switch col {
	case 0:
		return item.ID

	case 1:
		return item.Timestamp

	case 2:
		return item.Call

	case 3:
		return item.Band

	case 4:
		return item.Frequency

	case 5:
		return item.Comments
	}

	return ""
}

// Sort is called by the TableView to sort the model
// always sorted by ID so latest spots are on top
func (m *DXClusterModel) Sort(col int, order walk.SortOrder) error {
	sort.SliceStable(m.items, func(i, j int) bool {
		a, b := m.items[i], m.items[j]

		return a.ID > b.ID
	})

	return m.SorterBase.Sort(m.sortColumn, m.sortOrder)
}

// ResetRows loads QSOs from the database
func (m *DXClusterModel) ResetRows() {
	var r []spot.Spot
	var err error

	// get new spots
	r, err = spot.NewSpots(m.lastID)
	if err != nil {
		log.Printf("%+v", err)
		MsgError(nil, err)
		return
	}

	// bail early if no updates
	if len(r) == 0 {
		return
	}

	// update models dataset
	for i := range r {
		m.items = append(m.items, &r[i])

		// keep track of max ID
		if m.lastID < r[i].ID {
			m.lastID = r[i].ID
		}
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

	flashWindow(mainWin, 3)
}

// dxClusterTableView returns the DX Cluster TableView to be included in the apps MainWindow
func dxClusterTableView() declarative.TableView {
	var tv *walk.TableView

	dxclustermodel = NewDXClusterModel()

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
						err := launchQRZPage(dxclustermodel.items[idx].Call)
						if err != nil {
							MsgError(mainWin, err)
							log.Printf("%+v", err)
							return
						}
					}
				},
			},
			declarative.Action{
				Text: "PSKreporter",
				OnTriggered: func() {
					idx := tv.CurrentIndex()
					if idx >= 0 {
						err := launchPSKreporter(dxclustermodel.items[idx].Call)
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
						err := copyToClipboard(dxclustermodel.items[idx].Call)
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
			{Title: "Spot #", Hidden: true},
			{Title: "Time"},
			{Title: "Callsign"},
			{Title: "Band"},
			{Title: "Frequency", Alignment: declarative.AlignFar},
			{Title: "Comments", Width: 150},
			{Title: ""},
		},
		Model: dxclustermodel,
		StyleCell: func(style *walk.CellStyle) {
			drawCellStyles(tv, style, dxclustermodel.SorterBase)
		},
	}
}

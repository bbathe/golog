package ui

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bbathe/golog/config"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

type DXClusterSpot struct {
	Timestamp string
	Call      string
	Band      string
	Frequency string
	Comments  string
}

// DXClusterModel is used to display cluster spots in the main windows TableView
type DXClusterModel struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn int
	sortOrder  walk.SortOrder
	items      []*DXClusterSpot
}

func NewDXClusterModel() *DXClusterModel {
	m := new(DXClusterModel)
	m.sortColumn = 1
	m.sortOrder = walk.SortDescending
	m.ResetRows()

	return m
}

// RefreshPosts is called by various places in the app to reload the HamAlert TableView
func RefreshPosts() {
	dxclustermodel.ResetRows()
}

func AddSpot(timestamp, call, frequency, comments string) error {
	// make timestamps look how we want
	t, err := time.Parse("1504", timestamp)
	if err != nil {
		log.Printf("%+v", err)
		MsgError(nil, err)
		return err
	}

	// get frequency as a float
	freq, err := strconv.ParseFloat(frequency, 64)
	if err != nil {
		log.Printf("%+v", err)
		MsgError(nil, err)
		return err
	}

	// add spot
	dxclustermodel.items = append(dxclustermodel.items, &DXClusterSpot{
		Timestamp: t.Format("15:04"),
		Call:      strings.ToUpper(call),
		Band:      config.LookupBand(int(freq)),
		Frequency: formatFrequency(frequency),
		Comments:  strings.TrimSpace(comments),
	})

	dxclustermodel.ResetRows()

	// notify user of new spot
	flashWindow(mainWin.Handle(), 3)

	return nil
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
		return item.Timestamp

	case 1:
		return item.Call

	case 2:
		return item.Band

	case 3:
		return item.Frequency

	case 4:
		return item.Comments
	}

	return ""
}

// Sort is called by the TableView to sort the model
func (m *DXClusterModel) Sort(col int, order walk.SortOrder) error {
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
			return c(strings.Compare(a.Timestamp, b.Timestamp) < 0)

		case 1:
			return c(strings.Compare(a.Call, b.Call) < 0)

		case 2:
			return c(strings.Compare(a.Band, b.Band) < 0)

		case 3:
			return c(a.Frequency > b.Frequency)

		case 4:
			return c(strings.Compare(a.Comments, b.Comments) < 0)
		}
		return false
	})

	return m.SorterBase.Sort(col, order)
}

// ResetRows loads QSOs from the database
func (m *DXClusterModel) ResetRows() {
	// notify TableView about the reset
	m.PublishRowsReset()

	// maintain same sorting
	err := m.Sort(m.sortColumn, m.sortOrder)
	if err != nil {
		log.Printf("%+v", err)
		MsgError(nil, err)
		return
	}
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
			{Title: "Time"},
			{Title: "Call"},
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

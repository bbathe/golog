package ui

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bbathe/golog/db"
	"github.com/bbathe/golog/tasks"
	"golang.org/x/sys/windows"

	"github.com/bbathe/golog/config"

	"github.com/bbathe/golog/models/qso"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

var (
	appName       = "Go Log"
	appIcon       *walk.Icon
	runDll32      string
	flashWindowEx *windows.Proc

	mainWin        *walk.MainWindow
	qsomodel       *QSOModel
	selectedQSO    = &qso.QSO{}
	bndSelectedQSO *walk.DataBinder
)

func init() {
	var err error

	// load app icon
	appIcon, err = walk.Resources.Icon("2")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// full path to rundll32 for launching web browser
	runDll32 = filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")

	winuserDll, err := windows.LoadDLL("User32.dll")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	flashWindowEx, err = winuserDll.FindProc("FlashWindowEx")
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

type BandListModel struct {
	walk.BindingValueProvider
	walk.ListModelBase
	items []string
}

func NewBandListModel() *BandListModel {
	m := &BandListModel{}
	m.RefreshItems()

	return m
}

func (m *BandListModel) RefreshItems() {
	m.items = config.ListBandNames()
	m.PublishItemsReset()
}

func (m *BandListModel) ItemCount() int {
	return len(m.items)
}

func (m *BandListModel) BindingValue(index int) interface{} {
	return m.items[index]
}

func (m *BandListModel) Value(index int) interface{} {
	return m.items[index]
}

type ModeListModel struct {
	walk.BindingValueProvider
	walk.ListModelBase
	items []string
}

func NewModeListModel() *ModeListModel {
	m := &ModeListModel{}
	m.RefreshItems()

	return m
}

func (m *ModeListModel) RefreshItems() {
	m.items = config.ListModeNames()
	m.PublishItemsReset()
}

func (m *ModeListModel) ItemCount() int {
	return len(m.items)
}

func (m *ModeListModel) BindingValue(index int) interface{} {
	return m.items[index]
}

func (m *ModeListModel) Value(index int) interface{} {
	return m.items[index]
}

// gologWindow creates the main window and begins processing of user input
func GoLogWindow() error {
	var err error

	var leDate *walk.LineEdit
	var leTime *walk.LineEdit
	var leCall *walk.LineEdit
	var cbBand *walk.ComboBox
	var cbMode *walk.ComboBox
	var leRSTRcvd *walk.LineEdit
	var leRSTSent *walk.LineEdit

	var pbQRZ *walk.PushButton
	var pbCurrentTime *walk.PushButton
	var pbCurrentDate *walk.PushButton

	qsomodel = NewQSOModel()
	bands := NewBandListModel()
	modes := NewModeListModel()

	// golog main window
	err = declarative.MainWindow{
		AssignTo: &mainWin,
		Title:    appName,
		Icon:     appIcon,
		Visible:  false,
		Font: declarative.Font{
			Family:    "MS Shell Dlg 2",
			PointSize: 10,
		},
		Layout: declarative.VBox{},
		DataBinder: declarative.DataBinder{
			Name:           "qso",
			AssignTo:       &bndSelectedQSO,
			DataSource:     selectedQSO,
			ErrorPresenter: declarative.ToolTipErrorPresenter{},
		},
		MenuItems: []declarative.MenuItem{
			declarative.Menu{
				Text: "&File",
				Items: []declarative.MenuItem{
					declarative.Action{
						Text: "&Options...",
						OnTriggered: func() {
							updateConfig()

							// in case lookups were changed
							bands.RefreshItems()
							modes.RefreshItems()
						},
					},
					declarative.Separator{},
					declarative.Action{
						Text: "E&xit",
						OnTriggered: func() {
							mainWin.Close()
						},
					},
				},
			},
			declarative.Menu{
				Text: "&ADIF",
				Items: []declarative.MenuItem{
					declarative.Action{
						Text: "&Import...",
						OnTriggered: func() {
							err := importADIF(mainWin)
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							qsomodel.ResetRows()
						},
					},
					declarative.Action{
						Text: "&Export...",
						OnTriggered: func() {
							qsomodel.Export()
						},
					},
					declarative.Action{
						Text: "&Export All...",
						OnTriggered: func() {
							exportADIF(mainWin)
						},
					},
				},
			},
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.HBox{MarginsZero: true},
				Children: []declarative.Widget{
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Date",
							},
							declarative.Composite{
								Layout: declarative.HBox{MarginsZero: true},
								Children: []declarative.Widget{
									declarative.LineEdit{
										Text:     declarative.Bind("Date"),
										AssignTo: &leDate,
										OnTextChanged: func() {
											selectedQSO.Date = leDate.Text()
										},
									},
									declarative.PushButton{
										AssignTo:    &pbCurrentDate,
										Text:        "\U0001F4C5",
										ToolTipText: "set to current date",
										MaxSize: declarative.Size{
											Width: 30,
										},
										MinSize: declarative.Size{
											Width: 30,
										},
										Font: declarative.Font{
											Family:    "MS Shell Dlg 2",
											PointSize: 9,
										},
										OnClicked: func() {
											n := time.Now().UTC()
											selectedQSO.Date = n.Format("2006-01-02")

											// refresh
											err := bndSelectedQSO.Reset()
											if err != nil {
												MsgError(mainWin, err)
												log.Printf("%+v", err)
												return
											}
										},
									},
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Time",
							},
							declarative.Composite{
								Layout: declarative.HBox{MarginsZero: true},
								Children: []declarative.Widget{
									declarative.LineEdit{
										Text:     declarative.Bind("Time"),
										AssignTo: &leTime,
										OnTextChanged: func() {
											selectedQSO.Time = leTime.Text()
										},
									},
									declarative.PushButton{
										AssignTo:    &pbCurrentTime,
										Text:        "\U0001F551",
										ToolTipText: "set to current time",
										MaxSize: declarative.Size{
											Width: 30,
										},
										MinSize: declarative.Size{
											Width: 30,
										},
										Font: declarative.Font{
											Family:    "MS Shell Dlg 2",
											PointSize: 9,
										},
										OnClicked: func() {
											n := time.Now().UTC()
											selectedQSO.Time = n.Format("15:04")

											// refresh
											err := bndSelectedQSO.Reset()
											if err != nil {
												MsgError(mainWin, err)
												log.Printf("%+v", err)
												return
											}
										},
									},
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Callsign",
							},
							declarative.Composite{
								Layout: declarative.HBox{MarginsZero: true},
								Children: []declarative.Widget{
									declarative.LineEdit{
										Text:     declarative.Bind("Call"),
										CaseMode: declarative.CaseModeUpper,
										AssignTo: &leCall,
										OnTextChanged: func() {
											selectedQSO.Call = leCall.Text()
										},
									},
									declarative.PushButton{
										AssignTo:    &pbQRZ,
										Text:        "\U0001F310",
										ToolTipText: "visit QRZ.com page",
										MaxSize: declarative.Size{
											Width: 30,
										},
										MinSize: declarative.Size{
											Width: 30,
										},
										Font: declarative.Font{
											Family:    "MS Shell Dlg 2",
											PointSize: 9,
										},
										OnClicked: func() {
											err := launchQRZPage(selectedQSO.Call)
											if err != nil {
												MsgError(mainWin, err)
												log.Printf("%+v", err)
												return
											}
										},
									},
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Band",
							},
							declarative.ComboBox{
								Value:    declarative.Bind("Band"),
								Model:    bands,
								AssignTo: &cbBand,
								Editable: false,
								OnCurrentIndexChanged: func() {
									idx := cbBand.CurrentIndex()
									if idx >= 0 {
										selectedQSO.Band = bands.Value(idx).(string)
									}
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Mode",
							},
							declarative.ComboBox{
								Value:    declarative.Bind("Mode"),
								Model:    modes,
								AssignTo: &cbMode,
								Editable: false,
								OnCurrentIndexChanged: func() {
									idx := cbMode.CurrentIndex()
									if idx >= 0 {
										selectedQSO.Mode = modes.Value(idx).(string)
									}
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "RST Rcvd",
							},
							declarative.LineEdit{
								Text:     declarative.Bind("RSTRcvd"),
								CaseMode: declarative.CaseModeUpper,
								AssignTo: &leRSTRcvd,
								OnTextChanged: func() {
									selectedQSO.RSTRcvd = leRSTRcvd.Text()
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "RST Sent",
							},
							declarative.LineEdit{
								Text:     declarative.Bind("RSTSent"),
								CaseMode: declarative.CaseModeUpper,
								AssignTo: &leRSTSent,
								OnTextChanged: func() {
									selectedQSO.RSTSent = leRSTSent.Text()
								},
							},
						},
					},
				},
			},
			declarative.Composite{
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.PushButton{
						Text:        "New",
						ToolTipText: "start new QSO",
						MaxSize: declarative.Size{
							Width: 50,
						},
						MinSize: declarative.Size{
							Width: 50,
						},
						OnClicked: func() {
							n := time.Now().UTC()
							*selectedQSO = qso.QSO{
								Date: n.Format("2006-01-02"),
								Time: n.Format("15:04"),
							}

							// refresh
							err := bndSelectedQSO.Reset()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}
						},
					},
					declarative.PushButton{
						Text:        "Add",
						ToolTipText: "add QSO to log",
						MaxSize: declarative.Size{
							Width: 50,
						},
						MinSize: declarative.Size{
							Width: 50,
						},
						OnClicked: func() {
							// log under current station callsign
							selectedQSO.StationCallsign = config.Station.Callsign

							err := selectedQSO.Add()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							// refresh
							*selectedQSO = qso.QSO{}
							err = bndSelectedQSO.Reset()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							qsomodel.ResetRows()
						},
					},
					declarative.PushButton{
						Text:        "Update",
						ToolTipText: "update QSO in log",
						MaxSize: declarative.Size{
							Width: 50,
						},
						MinSize: declarative.Size{
							Width: 50,
						},
						OnClicked: func() {
							err := selectedQSO.UpdateOnlyQSO()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							// refresh
							*selectedQSO = qso.QSO{}
							err = bndSelectedQSO.Reset()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							qsomodel.ResetRows()
						},
					},
					declarative.PushButton{
						Text:        "Delete",
						ToolTipText: "delete QSO from log",
						MaxSize: declarative.Size{
							Width: 50,
						},
						MinSize: declarative.Size{
							Width: 50,
						},
						OnClicked: func() {
							err := selectedQSO.Delete()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							// refresh
							*selectedQSO = qso.QSO{}
							err = bndSelectedQSO.Reset()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							qsomodel.ResetRows()
						},
					},
					declarative.PushButton{
						Text:        "Search",
						ToolTipText: "search for QSOs from log",
						MaxSize: declarative.Size{
							Width: 50,
						},
						MinSize: declarative.Size{
							Width: 50,
						},
						OnClicked: func() {
							qsomodel.Search(
								leDate.Text(),
								leTime.Text(),
								leCall.Text(),
								cbBand.Text(),
								cbMode.Text(),
								leRSTRcvd.Text(),
								leRSTSent.Text(),
							)
						},
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							if button != walk.RightButton {
								return
							}

							qsomodel.Search(
								leDate.Text()+"%",
								leTime.Text()+"%",
								leCall.Text()+"%",
								cbBand.Text()+"%",
								cbMode.Text()+"%",
								leRSTRcvd.Text()+"%",
								leRSTSent.Text()+"%",
							)
						},
					},
					declarative.PushButton{
						Text:        "Cancel",
						ToolTipText: "cancel QSO add/changes and clear search",
						MaxSize: declarative.Size{
							Width: 50,
						},
						MinSize: declarative.Size{
							Width: 50,
						},
						OnClicked: func() {
							// refresh
							*selectedQSO = qso.QSO{}
							err := bndSelectedQSO.Reset()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							qsomodel.ClearSearch()
						},
					},
				},
			},
			qsoTableView(),
			dxClusterTableView(),
			mainStatusBar(),
		},
	}.Create()
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return err
	}

	// set window position based on config
	err = mainWin.SetBounds(config.UI.MainWinRectangle.ToBounds())
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return err
	}

	// save windows position in config during window close
	mainWin.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		config.UI.MainWinRectangle.FromBounds(mainWin.Bounds())
	})

	// sort to latest qsos on top
	err = qsomodel.Sort(0, walk.SortDescending)
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return err
	}

	// sort to latest cluster posts on top
	err = dxclustermodel.Sort(0, walk.SortAscending)
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return err
	}

	// remove buttons from tabstop
	win.SetWindowLong(pbCurrentDate.Handle(), win.GWL_STYLE,
		win.GetWindowLong(pbCurrentDate.Handle(), win.GWL_STYLE) & ^win.WS_TABSTOP)
	win.SetWindowLong(pbCurrentTime.Handle(), win.GWL_STYLE,
		win.GetWindowLong(pbCurrentTime.Handle(), win.GWL_STYLE) & ^win.WS_TABSTOP)
	win.SetWindowLong(pbQRZ.Handle(), win.GWL_STYLE,
		win.GetWindowLong(pbQRZ.Handle(), win.GWL_STYLE) & ^win.WS_TABSTOP)

	// make visible
	mainWin.SetVisible(true)

	// start message loop
	mainWin.Run()

	return nil
}

func updateConfig() {
	// stop background tasks
	tasks.Pause()

	err := OptionsWindow(mainWin)
	if err != nil {
		MsgError(mainWin, err)
		log.Printf("%+v", err)
		return
	}

	// reset spot history
	err = db.NewSpotDb()
	if err != nil {
		MsgError(mainWin, err)
		log.Printf("%+v", err)
		return
	}

	// restart background tasks
	go func() {
		tasks.Start()
	}()

	// reload so new changes are in effect
	qsomodel.ResetRows()

	err = bndSelectedQSO.Reset()
	if err != nil {
		MsgError(mainWin, err)
		log.Printf("%+v", err)
		return
	}
}

package ui

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/db"

	"github.com/bbathe/golog/models/qso"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

var (
	appName  = "Go Log"
	appIcon  *walk.Icon
	runDll32 string

	mainWin        *walk.MainWindow
	qsomodel       *QSOModel
	selectedQSO    = &qso.QSO{}
	bndSelectedQSO *walk.DataBinder
)

// initVariables initializes shared ui variables
func initVariables() {
	// load app icon
	ico, err := walk.Resources.Icon("3")
	if err != nil {
		MsgError(nil, err)
		log.Fatalf("%+v", err)
	}
	appIcon = ico

	// full path to rundll32 for launching web browser
	runDll32 = filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")
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

	initVariables()

	qsomodel = NewQSOModel()
	bands := []string{""}
	bands = append(bands, config.ListBandNames()...)
	modes := []string{""}
	modes = append(modes, config.ListModeNames()...)

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
						Text: "&New",
						OnTriggered: func() {
							err := db.NewQSODb()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							qsomodel.ResetRows()
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
						Text: "&Import",
						OnTriggered: func() {
							importADIF(qsomodel)
						},
					},
					declarative.Action{
						Text: "&Export",
						OnTriggered: func() {
							exportADIF()
						},
					},
				},
			},
		},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.HBox{},
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
								Text: "Call",
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
										selectedQSO.Band = bands[idx]
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
										selectedQSO.Mode = modes[idx]
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
							err := selectedQSO.Add()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							*selectedQSO = qso.QSO{}

							// refresh
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

							*selectedQSO = qso.QSO{}

							// refresh
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

							*selectedQSO = qso.QSO{}

							// refresh
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

							qsomodel.ResetRows()
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
							*selectedQSO = qso.QSO{}
							qsomodel.ClearSearch()

							// refresh
							err := bndSelectedQSO.Reset()
							if err != nil {
								MsgError(mainWin, err)
								log.Printf("%+v", err)
								return
							}

							qsomodel.ResetRows()
						},
					},
				},
			},
			qsoTableView(),
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

// importADIF drives the user thru doing an ADIF import
func importADIF(model *QSOModel) {
	file, err := OpenFilePicker(
		[]FileFilter{
			{Description: "ADIF Files", Wildcard: "*.adif;*.adi"},
			{Description: "All Files", Wildcard: "*.*"},
		},
	)
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	if file != nil {
		qs, err := adif.ReadFromFile(*file, qso.Sent)
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}

		err = qso.BulkAdd(qs)
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}
		model.ResetRows()
	}
}

// exportADIF drives the user thru doing an ADIF export
func exportADIF() {
	file, err := SaveFilePicker(
		[]FileFilter{
			{Description: "ADIF Files", Wildcard: "*.adif;*.adi"},
			{Description: "All Files", Wildcard: "*.*"},
		},
	)
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	if file != nil {
		qs, err := qso.All()
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}

		err = adif.WriteToFile(qs, *file)
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}
	}
}

// // importTQSLconfig drives the user thru doing an import of the lookup data from the TQSL config.xml
// func importTQSLconfig() {
// 	// change working directory to where tqsl.exe is, thats where config.xml is also
// 	err := os.Chdir(filepath.Dir(config.TQSL.ExeLocation))
// 	if err != nil {
// 		MsgError(nil, err)
// 		log.Printf("%+v", err)
// 		return
// 	}

// 	// prompt user for file
// 	f, err := OpenFilePicker(
// 		[]FileFilter{
// 			{Description: "XML Files", Wildcard: "*.xml"},
// 			{Description: "All Files", Wildcard: "*.*"},
// 		},
// 	)
// 	if err != nil {
// 		MsgError(nil, err)
// 		log.Printf("%+v", err)
// 		return
// 	}

// 	// update lookups
// 	err = config.UpdateLookupsFromTQSL(*f)
// 	if err != nil {
// 		MsgError(nil, err)
// 		log.Printf("%+v", err)
// 		return
// 	}

// 	basefn, err := os.Executable()
// 	if err != nil {
// 		MsgError(nil, err)
// 		log.Printf("%+v", err)
// 		return
// 	}

// 	err = config.WriteLookupsToFile(filepath.Join(filepath.Dir(basefn), "lookups.yaml"))
// 	if err != nil {
// 		MsgError(nil, err)
// 		log.Printf("%+v", err)
// 		return
// 	}
// }

// launchQRZPage opens the users default web browser to the qso partners QRZ.com page
func launchQRZPage(call string) error {
	u := "https://www.qrz.com"
	if call != "" {
		u += "/db/" + strings.Replace(call, "%", "", -1)
	}

	err := exec.Command(runDll32, "url.dll,FileProtocolHandler", u).Start()
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return err
	}

	return nil
}

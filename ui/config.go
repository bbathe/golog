package ui

import (
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/db"
	"github.com/lxn/walk"

	"github.com/lxn/walk/declarative"
)

var (
	newConfig config.Configuration

	modelBands *BandLookupModel
	modelModes *ModeLookupModel

	configForm walk.Form
)

func persistConfigChanges() error {
	// persist config
	err := config.Reload(newConfig)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// persist lookups
	err = config.ReloadLookups(config.Lookups{
		Bands: modelBands.GetBands(),
		Modes: modelModes.GetModes(),
	})
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}

func OptionsWindow(parent *walk.MainWindow) error {
	// make working copy of current config
	err := config.Copy(&newConfig)
	if err != nil {
		log.Printf("%+v", err)
	}

	if parent == nil {
		var configWin *walk.MainWindow

		// if no parent, make it a window
		err := declarative.MainWindow{
			AssignTo: &configWin,
			Title:    appName + " Options",
			Icon:     appIcon,
			Visible:  false,
			MinSize:  declarative.Size{Width: 800},
			Font: declarative.Font{
				Family:    "MS Shell Dlg 2",
				PointSize: 10,
			},
			Layout: declarative.VBox{},
			Children: []declarative.Widget{
				declarative.TabWidget{
					Alignment: declarative.AlignHNearVNear,
					Pages: []declarative.TabPage{
						tabConfigGeneral(),
						tabConfigSourceFiles(),
						tabConfigLogbookServices(),
						tabConfigDXClusters(),
					},
				},
				declarative.Composite{
					Layout: declarative.HBox{},
					Children: []declarative.Widget{
						declarative.HSpacer{},
						declarative.PushButton{
							Text: "OK",
							OnClicked: func() {
								err := persistConfigChanges()
								if err != nil {
									MsgError(configForm, err)
									log.Printf("%+v", err)
									return
								}

								configWin.Close()
							},
						},
						declarative.PushButton{
							Text: "Cancel",
							OnClicked: func() {
								configWin.Close()
							},
						},
					},
				},
			},
		}.Create()
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return err
		}

		// set common parent for all pop-ups, etc.
		configForm = configWin.Form()

		// make visible
		configWin.SetVisible(true)

		// start message loop
		configWin.Run()
	} else {
		var configDlg *walk.Dialog

		err = declarative.Dialog{
			AssignTo:  &configDlg,
			Title:     appName + " Options",
			Icon:      appIcon,
			FixedSize: true,
			MinSize:   declarative.Size{Width: 800},
			Font: declarative.Font{
				Family:    "MS Shell Dlg 2",
				PointSize: 10,
			},
			Layout: declarative.VBox{},
			Children: []declarative.Widget{
				declarative.TabWidget{
					Alignment: declarative.AlignHNearVNear,
					Pages: []declarative.TabPage{
						tabConfigGeneral(),
						tabConfigSourceFiles(),
						tabConfigLogbookServices(),
						tabConfigDXClusters(),
						tabConfigLookups(),
					},
				},
				declarative.Composite{
					Layout: declarative.HBox{},
					Children: []declarative.Widget{
						declarative.HSpacer{},
						declarative.PushButton{
							Text: "OK",
							OnClicked: func() {
								err := persistConfigChanges()
								if err != nil {
									MsgError(configForm, err)
									log.Printf("%+v", err)
									return
								}

								// reopen database
								err = db.OpenQSODb()
								if err != nil {
									MsgError(configForm, err)
									log.Printf("%+v", err)
									return
								}

								configDlg.Accept()
							},
						},
						declarative.PushButton{
							Text: "Cancel",
							OnClicked: func() {
								configDlg.Cancel()
							},
						},
					},
				},
			},
		}.Create(parent)
		if err != nil {
			MsgError(parent, err)
			log.Printf("%+v", err)
			return err
		}

		// set common parent for all pop-ups, etc.
		configForm = configDlg.Form()

		// start message loop
		configDlg.Run()
	}

	return nil
}

func tabConfigGeneral() declarative.TabPage {
	var leCallsign *walk.LineEdit
	var leQSODatabase *walk.LineEdit
	var neQSOHistory *walk.NumberEdit
	var neQSOLimit *walk.NumberEdit
	var leWorkingDirectory *walk.LineEdit

	return declarative.TabPage{
		Title:  "General",
		Layout: declarative.VBox{Alignment: declarative.AlignHNearVNear},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.VBox{},
				DataBinder: declarative.DataBinder{
					DataSource:     &newConfig.Station,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Children: []declarative.Widget{
					declarative.Label{
						Text: "Station Callsign",
					},
					declarative.LineEdit{
						AssignTo: &leCallsign,
						Text:     declarative.Bind("Callsign"),
						CaseMode: declarative.CaseModeUpper,
						OnTextChanged: func() {
							newConfig.Station.Callsign = leCallsign.Text()
						},
					},
				},
			},
			declarative.Composite{
				Layout: declarative.VBox{},
				Children: []declarative.Widget{
					declarative.Label{
						Text: "Database",
					},
					declarative.Composite{
						Layout: declarative.HBox{MarginsZero: true},
						DataBinder: declarative.DataBinder{
							DataSource:     &newConfig.QSODatabase,
							ErrorPresenter: declarative.ToolTipErrorPresenter{},
						},
						Children: []declarative.Widget{
							declarative.LineEdit{
								AssignTo: &leQSODatabase,
								Text:     declarative.Bind("Location"),
								ReadOnly: true,
								OnTextChanged: func() {
									newConfig.QSODatabase.Location = leQSODatabase.Text()
								},
							},
							declarative.PushButton{
								Text:    "\u2026",
								MaxSize: declarative.Size{Width: 30},
								MinSize: declarative.Size{Width: 30},
								Font: declarative.Font{
									Family:    "MS Shell Dlg 2",
									PointSize: 9,
								},
								OnClicked: func() {
									fname, err := SaveFilePicker(configForm, "Select QSO database file", "DB Files (*.db)|*.db|All Files (*.*)|*.*")
									if err != nil {
										MsgError(configForm, err)
										log.Printf("%+v", err)
										return
									}

									if fname != nil {
										err = leQSODatabase.SetText(*fname)
										if err != nil {
											MsgError(configForm, err)
											log.Printf("%+v", err)
											return
										}
									}
								},
							},
						},
					},
				},
			},
			declarative.RadioButtonGroupBox{
				Title:  "QSO Tableview",
				Layout: declarative.HBox{MarginsZero: true},
				DataBinder: declarative.DataBinder{
					DataSource:     &newConfig.QSOTableview,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Children: []declarative.Widget{
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text:        "History",
								ToolTipText: "how far back (in days) of QSOs to show in the QSO tableview",
							},
							declarative.NumberEdit{
								AssignTo: &neQSOHistory,
								Value:    declarative.Bind("History"),
								Decimals: 0,
								OnValueChanged: func() {
									newConfig.QSOTableview.History = int(neQSOHistory.Value())
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text:        "Limit",
								ToolTipText: "the maximum number of QSOs to show in the QSO tableview",
							},
							declarative.NumberEdit{
								AssignTo: &neQSOLimit,
								Value:    declarative.Bind("Limit"),
								Decimals: 0,
								OnValueChanged: func() {
									newConfig.QSOTableview.Limit = int(neQSOLimit.Value())
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
						Text: "Working Directory",
					},
					declarative.Composite{
						Layout: declarative.HBox{MarginsZero: true},
						DataBinder: declarative.DataBinder{
							DataSource:     &newConfig,
							ErrorPresenter: declarative.ToolTipErrorPresenter{},
						},
						Children: []declarative.Widget{
							declarative.LineEdit{
								AssignTo: &leWorkingDirectory,
								Text:     declarative.Bind("WorkingDirectory"),
								ReadOnly: true,
								OnTextChanged: func() {
									newConfig.WorkingDirectory = leWorkingDirectory.Text()
								},
							},
							declarative.PushButton{
								Text:    "\u2026",
								MaxSize: declarative.Size{Width: 30},
								MinSize: declarative.Size{Width: 30},
								Font: declarative.Font{
									Family:    "MS Shell Dlg 2",
									PointSize: 9,
								},
								OnClicked: func() {
									dname, err := OpenFolderPicker(configForm, "Choose folder for temporary working files")
									if err != nil {
										MsgError(configForm, err)
										log.Printf("%+v", err)
										return
									}

									if dname != nil {
										err = leWorkingDirectory.SetText(*dname)
										if err != nil {
											MsgError(configForm, err)
											log.Printf("%+v", err)
											return
										}
									}
								},
							},
						},
					},
				},
			},
			declarative.HSpacer{},
		},
	}
}

// sourceFileModel is back by config, no data items here
type sourceFileModel struct {
	walk.ListModelBase
}

func NewSourceFileModel() *sourceFileModel {
	return &sourceFileModel{}
}

func (m *sourceFileModel) ItemCount() int {
	return len(newConfig.SourceFiles)
}

func (m *sourceFileModel) Value(index int) interface{} {
	return newConfig.SourceFiles[index].Location
}

func (m *sourceFileModel) ResetItems() {
	m.PublishItemsReset()
}

func tabConfigSourceFiles() declarative.TabPage {
	var leSourceFileLocation *walk.LineEdit
	var lbSourceFiles *walk.ListBox

	model := NewSourceFileModel()

	return declarative.TabPage{
		Title:  "Source Files",
		Layout: declarative.VBox{Alignment: declarative.AlignHNearVNear},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.VBox{},
				Children: []declarative.Widget{
					declarative.Label{
						Text: "ADIF file",
					},
					declarative.Composite{
						Layout: declarative.HBox{MarginsZero: true},
						DataBinder: declarative.DataBinder{
							DataSource:     &newConfig.SourceFiles,
							ErrorPresenter: declarative.ToolTipErrorPresenter{},
						},
						Children: []declarative.Widget{
							declarative.LineEdit{
								AssignTo: &leSourceFileLocation,
								Text:     declarative.Bind("ExeLocation"),
								ReadOnly: true,
							},
							declarative.PushButton{
								Text:    "\u2026",
								MaxSize: declarative.Size{Width: 30},
								MinSize: declarative.Size{Width: 30},
								Font: declarative.Font{
									Family:    "MS Shell Dlg 2",
									PointSize: 9,
								},
								OnClicked: func() {
									// prompt user for file
									fname, err := OpenFilePicker(configForm, "Select source file", "ADIF Files (*.adi;*.adif)|*.adi;*.adif|All Files (*.*)|*.*")
									if err != nil {
										MsgError(configForm, err)
										log.Printf("%+v", err)
										return
									}

									if fname != nil {
										err = leSourceFileLocation.SetText(*fname)
										if err != nil {
											MsgError(configForm, err)
											log.Printf("%+v", err)
											return
										}
									}
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.HBox{},
						Children: []declarative.Widget{
							declarative.PushButton{
								Text:    "\u2795",
								MaxSize: declarative.Size{Width: 30},
								MinSize: declarative.Size{Width: 30},
								OnClicked: func() {
									if leSourceFileLocation.Text() != "" {
										newConfig.AddSourceFile(leSourceFileLocation.Text())

										err := leSourceFileLocation.SetText("")
										if err != nil {
											MsgError(configForm, err)
											log.Printf("%+v", err)
											return
										}

										model.ResetItems()
									}
								},
							},
							declarative.PushButton{
								Text:    "\u2796",
								MaxSize: declarative.Size{Width: 30},
								MinSize: declarative.Size{Width: 30},
								OnClicked: func() {
									i := lbSourceFiles.CurrentIndex()
									if i >= 0 {
										newConfig.RemoveSourceFile(i)

										model.ResetItems()
									}
								},
							},
						},
					},
					declarative.ListBox{
						AssignTo: &lbSourceFiles,
						Model:    model,
					},
				},
			},
			declarative.HSpacer{},
		},
	}
}

func tabConfigLogbookServices() declarative.TabPage {
	var neQSLDelay *walk.NumberEdit

	var leCLEmail *walk.LineEdit
	var leCLPassword *walk.LineEdit
	var leCLCallsign *walk.LineEdit
	var leCLAPIKey *walk.LineEdit

	var leEQSLUsername *walk.LineEdit
	var leEQSLPassword *walk.LineEdit

	var leTQSLExeLocation *walk.LineEdit
	var leTQSLStationLocationName *walk.LineEdit

	var leQRZAPIKey *walk.LineEdit

	return declarative.TabPage{
		Title:  "Logbook Services",
		Layout: declarative.VBox{Alignment: declarative.AlignHNearVNear},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.VBox{},
				DataBinder: declarative.DataBinder{
					DataSource:     &newConfig.LogbookServices,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Children: []declarative.Widget{
					declarative.Label{
						Text:        "QSL Delay",
						ToolTipText: "the delay (in minutes) from when a QSO is logged vs sent to the logbook services",
					},
					declarative.NumberEdit{
						AssignTo: &neQSLDelay,
						Value:    declarative.Bind("QSLDelay"),
						Decimals: 0,
						OnValueChanged: func() {
							newConfig.LogbookServices.QSLDelay = int(neQSLDelay.Value())
						},
					},
				},
			},
			declarative.RadioButtonGroupBox{
				Title:  "LoTW",
				Layout: declarative.HBox{MarginsZero: true},
				DataBinder: declarative.DataBinder{
					DataSource:     &newConfig.LogbookServices.TQSL,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Children: []declarative.Widget{
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "TQSL Executable",
							},
							declarative.Composite{
								Layout: declarative.HBox{MarginsZero: true},
								DataBinder: declarative.DataBinder{
									DataSource:     &newConfig.LogbookServices.TQSL,
									ErrorPresenter: declarative.ToolTipErrorPresenter{},
								},
								Children: []declarative.Widget{
									declarative.LineEdit{
										AssignTo: &leTQSLExeLocation,
										Text:     declarative.Bind("ExeLocation"),
										ReadOnly: true,
										OnTextChanged: func() {
											newConfig.LogbookServices.TQSL.ExeLocation = leTQSLExeLocation.Text()
										},
									},
									declarative.PushButton{
										Text:    "\u2026",
										MaxSize: declarative.Size{Width: 30},
										MinSize: declarative.Size{Width: 30},
										Font: declarative.Font{
											Family:    "MS Shell Dlg 2",
											PointSize: 9,
										},
										OnClicked: func() {
											// prompt user for file
											fname, err := OpenFilePicker(configForm, "Select TQSL executable", "EXE Files (*.exe)|*.exe|All Files (*.*)|*.*")
											if err != nil {
												MsgError(configForm, err)
												log.Printf("%+v", err)
												return
											}

											if fname != nil {
												err = leTQSLExeLocation.SetText(*fname)
												if err != nil {
													MsgError(configForm, err)
													log.Printf("%+v", err)
													return
												}
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
								Text: "Station Location Name",
							},
							declarative.LineEdit{
								AssignTo: &leTQSLStationLocationName,
								Text:     declarative.Bind("StationLocationName"),
								OnTextChanged: func() {
									newConfig.LogbookServices.TQSL.StationLocationName = leTQSLStationLocationName.Text()
								},
							},
						},
					},
				},
			},
			declarative.RadioButtonGroupBox{
				Title:  "eQSL",
				Layout: declarative.HBox{MarginsZero: true},
				DataBinder: declarative.DataBinder{
					DataSource:     &newConfig.LogbookServices.EQSL,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Children: []declarative.Widget{
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Username",
							},
							declarative.LineEdit{
								AssignTo: &leEQSLUsername,
								Text:     declarative.Bind("Username"),
								OnTextChanged: func() {
									newConfig.LogbookServices.EQSL.Username = leEQSLUsername.Text()
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Password",
							},
							declarative.LineEdit{
								AssignTo: &leEQSLPassword,
								Text:     declarative.Bind("Password"),
								OnTextChanged: func() {
									newConfig.LogbookServices.EQSL.Password = leEQSLPassword.Text()
								},
							},
						},
					},
				},
			},
			declarative.RadioButtonGroupBox{
				Title:  "QRZ",
				Layout: declarative.HBox{MarginsZero: true},
				DataBinder: declarative.DataBinder{
					DataSource:     &newConfig.LogbookServices.QRZ,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Children: []declarative.Widget{
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "API Key",
							},
							declarative.LineEdit{
								AssignTo: &leQRZAPIKey,
								Text:     declarative.Bind("APIKey"),
								OnTextChanged: func() {
									newConfig.LogbookServices.QRZ.APIKey = leQRZAPIKey.Text()
								},
							},
						},
					},
				},
			},
			declarative.RadioButtonGroupBox{
				Title:  "Club Log",
				Layout: declarative.HBox{MarginsZero: true},
				DataBinder: declarative.DataBinder{
					DataSource:     &newConfig.LogbookServices.ClubLog,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Children: []declarative.Widget{
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Composite{
								Layout: declarative.VBox{},
								Children: []declarative.Widget{
									declarative.Label{
										Text: "Email",
									},
									declarative.LineEdit{
										AssignTo: &leCLEmail,
										Text:     declarative.Bind("Email"),
										OnTextChanged: func() {
											newConfig.LogbookServices.ClubLog.Email = leCLEmail.Text()
										},
									},
								},
							},
							declarative.Composite{
								Layout: declarative.VBox{},
								Children: []declarative.Widget{
									declarative.Label{
										Text: "Password",
									},
									declarative.LineEdit{
										AssignTo: &leCLPassword,
										Text:     declarative.Bind("Password"),
										OnTextChanged: func() {
											newConfig.LogbookServices.ClubLog.Password = leCLPassword.Text()
										},
									},
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Composite{
								Layout: declarative.VBox{},
								Children: []declarative.Widget{
									declarative.Label{
										Text: "Callsign",
									},
									declarative.LineEdit{
										AssignTo: &leCLCallsign,
										Text:     declarative.Bind("Callsign"),
										OnTextChanged: func() {
											newConfig.LogbookServices.ClubLog.Callsign = leCLCallsign.Text()
										},
									},
								},
							},
							declarative.Composite{
								Layout: declarative.VBox{},
								Children: []declarative.Widget{
									declarative.Label{
										Text: "API Key",
									},
									declarative.LineEdit{
										AssignTo: &leCLAPIKey,
										Text:     declarative.Bind("APIKey"),
										OnTextChanged: func() {
											newConfig.LogbookServices.ClubLog.APIKey = leCLAPIKey.Text()
										},
									},
								},
							},
						},
					},
				},
			},
			declarative.HSpacer{},
		},
	}
}

func tabConfigDXClusters() declarative.TabPage {
	var cbFlashWindowOnNewSpots *walk.CheckBox
	var leHAHostPort *walk.LineEdit
	var leHAUsername *walk.LineEdit
	var leHAPassword *walk.LineEdit

	return declarative.TabPage{
		Title:  "DX Clusters",
		Layout: declarative.VBox{Alignment: declarative.AlignHNearVNear},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.VBox{},
				DataBinder: declarative.DataBinder{
					DataSource:     &newConfig.ClusterServices,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Children: []declarative.Widget{
					declarative.CheckBox{
						AssignTo: &cbFlashWindowOnNewSpots,
						Text:     "Flash window on new spots",
						Checked:  declarative.Bind("FlashWindowOnNewSpots"),
						OnCheckStateChanged: func() {
							newConfig.ClusterServices.FlashWindowOnNewSpots = cbFlashWindowOnNewSpots.Checked()
						},
					},
				},
			},
			declarative.RadioButtonGroupBox{
				Title:  "HamAlert",
				Layout: declarative.HBox{MarginsZero: true},
				DataBinder: declarative.DataBinder{
					DataSource:     &newConfig.ClusterServices.HamAlert,
					ErrorPresenter: declarative.ToolTipErrorPresenter{},
				},
				Children: []declarative.Widget{
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Host:Port",
							},
							declarative.LineEdit{
								AssignTo: &leHAHostPort,
								Text:     declarative.Bind("HostPort"),
								OnTextChanged: func() {
									newConfig.ClusterServices.HamAlert.HostPort = leHAHostPort.Text()
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Username",
							},
							declarative.LineEdit{
								AssignTo: &leHAUsername,
								Text:     declarative.Bind("Username"),
								OnTextChanged: func() {
									newConfig.ClusterServices.HamAlert.Username = leHAUsername.Text()
								},
							},
						},
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Password",
							},
							declarative.LineEdit{
								AssignTo: &leHAPassword,
								Text:     declarative.Bind("Password"),
								OnTextChanged: func() {
									newConfig.ClusterServices.HamAlert.Password = leHAPassword.Text()
								},
							},
						},
					},
				},
			},
			declarative.HSpacer{},
		},
	}
}

type BandLookupModel struct {
	walk.ReflectTableModelBase
	walk.ItemChecker
	walk.SorterBase
	bands []config.Band
}

func NewBandLookupModel() *BandLookupModel {
	m := new(BandLookupModel)

	// make copy
	m.bands = make([]config.Band, len(config.Bands))
	for i := range config.Bands {
		m.bands[i] = config.Band{
			Band:     config.Bands[i].Band,
			FreqLow:  config.Bands[i].FreqLow,
			FreqHigh: config.Bands[i].FreqHigh,
			Visible:  config.Bands[i].Visible,
		}
	}

	return m
}

func (m *BandLookupModel) GetBands() []config.Band {
	return m.bands
}

func (m *BandLookupModel) Merge(bands []config.Band) {
	// driven by new bands
	var i int
	for i = 0; i < len(bands); i++ {
		// search thru old bands
		var j int
		for j = 0; j < len(m.bands); j++ {
			// match on band name
			if m.bands[j].Band == bands[i].Band {
				// update details, leave Visible alone
				m.bands[j].FreqLow = bands[i].FreqLow
				m.bands[j].FreqHigh = bands[i].FreqHigh
				break
			}
		}

		// if we didn't have this in the old bands, add it
		if j >= len(m.bands) {
			m.bands = append(m.bands, config.Band{
				Band:     bands[i].Band,
				FreqLow:  bands[i].FreqLow,
				FreqHigh: bands[i].FreqHigh,
				Visible:  bands[i].Visible,
			})
		}
	}

	// notify TableView about the reset
	m.PublishRowsReset()
}

func (m *BandLookupModel) RowCount() int {
	return len(m.bands)
}

func (m *BandLookupModel) Value(row, col int) interface{} {
	return m.bands[row].Band
}

func (m *BandLookupModel) Sort(col int, order walk.SortOrder) error {
	sort.SliceStable(m.bands, func(i, j int) bool {
		a, b := m.bands[i], m.bands[j]

		return strings.Compare(a.Band, b.Band) < 0
	})

	return m.SorterBase.Sort(99, walk.SortDescending)
}

func (m *BandLookupModel) Checked(index int) bool {
	return m.bands[index].Visible
}

func (m *BandLookupModel) SetChecked(index int, checked bool) error {
	m.bands[index].Visible = checked
	return nil
}

type ModeLookupModel struct {
	walk.ReflectTableModelBase
	walk.ItemChecker
	walk.SorterBase
	modes []config.Mode
}

func NewModeLookupModel() *ModeLookupModel {
	m := new(ModeLookupModel)

	// make copy
	m.modes = make([]config.Mode, len(config.Modes))
	for i := range config.Modes {
		m.modes[i] = config.Mode{
			Mode:    config.Modes[i].Mode,
			Submode: config.Modes[i].Submode,
			Visible: config.Modes[i].Visible,
		}
	}

	return m
}

func (m *ModeLookupModel) GetModes() []config.Mode {
	return m.modes
}

func (m *ModeLookupModel) Merge(modes []config.Mode) {
	// driven by new modes
	var i int
	for i = 0; i < len(modes); i++ {
		// search thru old modes
		var j int
		for j = 0; j < len(m.modes); j++ {
			// match on mode & submode names
			if m.modes[j].Mode == modes[i].Mode && m.modes[j].Submode == modes[i].Submode {
				// no more details, but we have a match
				break
			}
		}

		// if we didn't have this in the old modes
		if j >= len(m.modes) {
			// then add it
			m.modes = append(m.modes, config.Mode{
				Mode:    modes[i].Mode,
				Submode: modes[i].Submode,
				Visible: modes[i].Visible,
			})
		}
	}

	// notify TableView about the reset
	m.PublishRowsReset()
}

func (m *ModeLookupModel) RowCount() int {
	return len(m.modes)
}

func (m *ModeLookupModel) Value(row, col int) interface{} {
	item := m.modes[row]

	n := item.Mode
	if item.Submode != "" {
		n = item.Submode
	}

	return n
}

func (m *ModeLookupModel) Sort(col int, order walk.SortOrder) error {
	sort.SliceStable(m.modes, func(i, j int) bool {
		a, b := m.modes[i], m.modes[j]

		na := a.Mode
		if a.Submode != "" {
			na = a.Submode
		}

		nb := b.Mode
		if b.Submode != "" {
			nb = b.Submode
		}

		return strings.Compare(na, nb) < 0
	})

	return m.SorterBase.Sort(99, walk.SortDescending)
}

func (m *ModeLookupModel) Checked(index int) bool {
	return m.modes[index].Visible
}

func (m *ModeLookupModel) SetChecked(index int, checked bool) error {
	m.modes[index].Visible = checked
	return nil
}

func tabConfigLookups() declarative.TabPage {
	var leTQSLConfigXMLLocation *walk.LineEdit
	var tvBands *walk.TableView
	var tvModes *walk.TableView

	modelBands = NewBandLookupModel()
	modelModes = NewModeLookupModel()

	return declarative.TabPage{
		Title:  "Lookups",
		Layout: declarative.VBox{Alignment: declarative.AlignHNearVNear},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.VBox{},
				Children: []declarative.Widget{
					declarative.Label{
						Text: "Refresh from TQSL config.xml",
					},
					declarative.Composite{
						Layout: declarative.HBox{MarginsZero: true},
						Children: []declarative.Widget{
							declarative.LineEdit{
								AssignTo: &leTQSLConfigXMLLocation,
								ReadOnly: true,
							},
							declarative.PushButton{
								Text:    "\u2026",
								MaxSize: declarative.Size{Width: 30},
								MinSize: declarative.Size{Width: 30},
								Font: declarative.Font{
									Family:    "MS Shell Dlg 2",
									PointSize: 9,
								},
								OnClicked: func() {
									// prompt user for file
									fname, err := OpenFilePickerWithInitialDir(configForm, "Select TQSL config.xml", "XML Files (*.xml)|*.xml|All Files (*.*)|*.*", filepath.Dir(config.LogbookServices.TQSL.ExeLocation))
									if err != nil {
										MsgError(configForm, err)
										log.Printf("%+v", err)
										return
									}

									if fname != nil {
										err = leTQSLConfigXMLLocation.SetText(*fname)
										if err != nil {
											MsgError(configForm, err)
											log.Printf("%+v", err)
											return
										}
									}

									l, err := config.ReadLookupsFromTQSL(*fname)
									if err != nil {
										MsgError(configForm, err)
										log.Printf("%+v", err)
										return
									}

									modelBands.Merge(l.Bands)
									modelModes.Merge(l.Modes)
								},
							},
						},
					},
				},
			},
			declarative.HSpacer{},
			declarative.Label{
				Text:          "Check the bands/modes you want available when logging & editing a QSO",
				TextAlignment: declarative.AlignCenter,
			},
			declarative.Composite{
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.VSpacer{},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Bands",
							},
							declarative.TableView{
								AssignTo:            &tvBands,
								AlternatingRowBG:    true,
								CheckBoxes:          true,
								LastColumnStretched: true,
								HeaderHidden:        true,
								Columns: []declarative.TableViewColumn{
									{Title: "Band"},
								},
								Model: modelBands,
							},
						},
					},
					declarative.VSpacer{},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "Modes",
							},
							declarative.TableView{
								AssignTo:            &tvModes,
								AlternatingRowBG:    true,
								CheckBoxes:          true,
								LastColumnStretched: true,
								HeaderHidden:        true,
								Columns: []declarative.TableViewColumn{
									{Title: "Mode"},
								},
								Model: modelModes,
							},
						},
					},
					declarative.VSpacer{},
				},
			},
			declarative.HSpacer{},
		},
	}
}

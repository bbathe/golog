package ui

import (
	"log"
	"os"

	"github.com/bbathe/golog/config"
	"github.com/bbathe/golog/db"
	"github.com/lxn/walk"

	"github.com/lxn/walk/declarative"
)

var (
	configDlg *walk.Dialog

	newConfig config.Configuration
)

func optionsWindows(parent *walk.MainWindow) error {
	// make working copy
	err := config.Copy(&newConfig)
	if err != nil {
		MsgError(parent, err)
		log.Printf("%+v", err)
		return err
	}

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
				},
			},
			declarative.Composite{
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.HSpacer{},
					declarative.PushButton{
						Text: "OK",
						OnClicked: func() {
							// persist config
							err := config.Reload(newConfig)
							if err != nil {
								MsgError(configDlg, err)
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

	// start message loop
	configDlg.Run()

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
									fname, err := OpenFilePicker(configDlg, "Select QSO database file", "DB Files (*.db)|*.db|All Files (*.*)|*.*")
									if err != nil {
										MsgError(configDlg, err)
										log.Printf("%+v", err)
										return
									}

									if fname != nil {
										err = leQSODatabase.SetText(*fname)
										if err != nil {
											MsgError(configDlg, err)
											log.Printf("%+v", err)
											return
										}

										// if file doesn't exist, create new db
										_, err = os.Stat(*fname)
										if err != nil {
											if os.IsNotExist(err) {
												err = db.NewQSODb()
											} else {
												err = db.OpenQSODb()
											}
										}
										if err != nil {
											MsgError(configDlg, err)
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
									dname, err := OpenFolderPicker(configDlg, "Choose folder for temporary working files")
									if err != nil {
										MsgError(configDlg, err)
										log.Printf("%+v", err)
										return
									}

									if dname != nil {
										err = leWorkingDirectory.SetText(*dname)
										if err != nil {
											MsgError(configDlg, err)
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
									fname, err := OpenFilePicker(configDlg, "Select source file", "ADIF Files (*.adi;*.adif)|*.adi;*.adif|All Files (*.*)|*.*")
									if err != nil {
										MsgError(configDlg, err)
										log.Printf("%+v", err)
										return
									}

									if fname != nil {
										err = leSourceFileLocation.SetText(*fname)
										if err != nil {
											MsgError(configDlg, err)
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
											MsgError(configDlg, err)
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
											fname, err := OpenFilePicker(configDlg, "Select TQSL executable", "EXE Files (*.exe)|*.exe|All Files (*.*)|*.*")
											if err != nil {
												MsgError(configDlg, err)
												log.Printf("%+v", err)
												return
											}

											if fname != nil {
												err = leTQSLExeLocation.SetText(*fname)
												if err != nil {
													MsgError(configDlg, err)
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
	var leHAHostPort *walk.LineEdit
	var leHAUsername *walk.LineEdit
	var leHAPassword *walk.LineEdit

	return declarative.TabPage{
		Title:  "DX Clusters",
		Layout: declarative.VBox{Alignment: declarative.AlignHNearVNear},
		Children: []declarative.Widget{
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

package ui

import (
	"log"

	"github.com/bbathe/golog/adif"
	"github.com/bbathe/golog/models/qso"
	"github.com/lxn/walk"

	"github.com/lxn/walk/declarative"
)

var (
	adifDlg *walk.Dialog
)

// importADIF drives the user thru importing QSOs from an ADIF file
func importADIF(parent walk.Form) error {
	var leADIFFile *walk.LineEdit

	var cbClublog *walk.CheckBox
	var cbEQSL *walk.CheckBox
	var cbLoTW *walk.CheckBox
	var cbQRZ *walk.CheckBox

	err := declarative.Dialog{
		AssignTo:  &adifDlg,
		Title:     appName + " Import ADIF",
		Icon:      appIcon,
		FixedSize: true,
		MinSize:   declarative.Size{Width: 600},
		Font: declarative.Font{
			Family:    "MS Shell Dlg 2",
			PointSize: 10,
		},
		Layout: declarative.VBox{Alignment: declarative.AlignHNearVNear},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.VBox{},
				Children: []declarative.Widget{
					declarative.Label{
						Text: "ADIF File",
					},
					declarative.Composite{
						Layout: declarative.HBox{MarginsZero: true},
						Children: []declarative.Widget{
							declarative.LineEdit{
								AssignTo: &leADIFFile,
								ReadOnly: true,
								OnTextChanged: func() {
									newConfig.QSODatabase.Location = leADIFFile.Text()
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
									fname, err := OpenFilePicker(parent, "Select file to import", "ADIF Files (*.adi;*.adif)|*.adi;*.adif|All Files (*.*)|*.*")
									if err != nil {
										MsgError(adifDlg, err)
										log.Printf("%+v", err)
										return
									}

									if fname != nil {
										err = leADIFFile.SetText(*fname)
										if err != nil {
											MsgError(adifDlg, err)
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
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.RadioButtonGroupBox{
						Title:  "QSL to Logbook Services",
						Layout: declarative.HBox{},
						Children: []declarative.Widget{
							declarative.CheckBox{
								Text:     "LoTW",
								AssignTo: &cbLoTW,
							},
							declarative.CheckBox{
								Text:     "eQSL",
								AssignTo: &cbEQSL,
							},
							declarative.CheckBox{
								Text:     "QRZ",
								AssignTo: &cbQRZ,
							},
							declarative.CheckBox{
								Text:     "Club Log",
								AssignTo: &cbClublog,
							},
						},
					},
				},
			},
			declarative.Composite{
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.HSpacer{},
					declarative.PushButton{
						Text: "OK",
						OnClicked: func() {
							fname := leADIFFile.Text()
							if fname != "" {
								qsllotw := qso.Sent
								if cbLoTW.Checked() {
									qsllotw = qso.NotSent
								}

								qsleqsl := qso.Sent
								if cbEQSL.Checked() {
									qsleqsl = qso.NotSent
								}

								qslqrz := qso.Sent
								if cbQRZ.Checked() {
									qslqrz = qso.NotSent
								}

								qslclublog := qso.Sent
								if cbClublog.Checked() {
									qslclublog = qso.NotSent
								}

								qs, err := adif.ReadFromFile(fname, qsllotw, qsleqsl, qslqrz, qslclublog)
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
							}

							adifDlg.Accept()
						},
					},
					declarative.PushButton{
						Text: "Cancel",
						OnClicked: func() {
							adifDlg.Cancel()
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
	adifDlg.Run()

	return nil
}

// exportADIF drives the user thru exporting QSOs to an ADIF file
func exportADIF(parent walk.Form) {
	fname, err := SaveFilePicker(parent, "Select file to export QSOs", "ADIF Files (*.adi;*.adif)|*.adi;*.adif|All Files (*.*)|*.*")
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return
	}

	if fname != nil {
		qs, err := qso.All()
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}

		err = adif.WriteToFile(qs, *fname)
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return
		}
	}
}

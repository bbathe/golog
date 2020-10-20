package ui

import (
	"image/color"
	"log"

	"github.com/bbathe/golog/tasks"
	"github.com/bbathe/golog/util"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

var (
	icSourceFiles *walk.ImageView
	icLoTW        *walk.ImageView
	icEQSL        *walk.ImageView
	icQRZ         *walk.ImageView
	icClubLog     *walk.ImageView
	icHamAlert    *walk.ImageView

	imgOK     walk.Image
	imgFailed walk.Image
)

func statusImage(s tasks.GoLogTaskStatus) walk.Image {
	if s == tasks.TaskStatusOK {
		return imgOK
	}
	return imgFailed
}

func updateStatuses(statuses []tasks.GoLogTaskStatus) {
	err := icSourceFiles.SetImage(statusImage(statuses[tasks.TaskSourceFiles]))
	if err != nil {
		MsgError(mainWin, err)
		log.Printf("%+v", err)
		return
	}

	err = icLoTW.SetImage(statusImage(statuses[tasks.TaskQSLLoTW]))
	if err != nil {
		MsgError(mainWin, err)
		log.Printf("%+v", err)
		return
	}

	err = icEQSL.SetImage(statusImage(statuses[tasks.TaskQSLEQSL]))
	if err != nil {
		MsgError(mainWin, err)
		log.Printf("%+v", err)
		return
	}

	err = icQRZ.SetImage(statusImage(statuses[tasks.TaskQSLQRZ]))
	if err != nil {
		MsgError(mainWin, err)
		log.Printf("%+v", err)
		return
	}

	err = icClubLog.SetImage(statusImage(statuses[tasks.TaskQSLClubLog]))
	if err != nil {
		MsgError(mainWin, err)
		log.Printf("%+v", err)
		return
	}

	err = icHamAlert.SetImage(statusImage(statuses[tasks.TaskHamAlert]))
	if err != nil {
		MsgError(mainWin, err)
		log.Printf("%+v", err)
		return
	}
}

func mainStatusBar() declarative.Composite {
	var err error

	imgOK, err = walk.NewIconFromImageForDPI(util.GenerateStatusImage(color.RGBA{R: 34, G: 139, B: 34, A: 255}), 96)
	if err != nil {
		log.Printf("%+v", err)
	}

	imgFailed, err = walk.NewIconFromImageForDPI(util.GenerateStatusImage(color.RGBA{R: 237, G: 28, B: 36, A: 255}), 96)
	if err != nil {
		log.Printf("%+v", err)
	}

	c := declarative.Composite{
		Layout: declarative.HBox{MarginsZero: true},
		Children: []declarative.Widget{
			declarative.ImageView{
				Image:       imgOK,
				AssignTo:    &icSourceFiles,
				ToolTipText: "Source Files",
			},
			declarative.ImageView{
				Image:       imgOK,
				AssignTo:    &icLoTW,
				ToolTipText: "LoTW",
			},
			declarative.ImageView{
				Image:       imgOK,
				AssignTo:    &icEQSL,
				ToolTipText: "eQSL",
			},
			declarative.ImageView{
				Image:       imgOK,
				AssignTo:    &icQRZ,
				ToolTipText: "QRZ",
			},
			declarative.ImageView{
				Image:       imgOK,
				AssignTo:    &icClubLog,
				ToolTipText: "Club Log",
			},
			declarative.ImageView{
				Image:       imgOK,
				AssignTo:    &icHamAlert,
				ToolTipText: "HamAlert",
			},
		},
	}

	tasks.Attach(updateStatuses)

	return c
}

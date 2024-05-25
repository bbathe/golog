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
	icQRZ         *walk.ImageView
	icClubLog     *walk.ImageView
	icHamAlert    *walk.ImageView

	imgOK         walk.Image
	imgFailed     walk.Image
	imgNotRunning walk.Image
)

func statusImage(s tasks.GoLogTaskStatus) walk.Image {
	switch s {
	case tasks.TaskStatusOK:
		return imgOK
	case tasks.TaskStatusNotRunning:
		return imgNotRunning
	}

	return imgFailed
}

func updateStatuses(statuses []tasks.GoLogTaskStatus) {
	if icSourceFiles != nil {
		err := icSourceFiles.SetImage(statusImage(statuses[tasks.TaskSourceFiles]))
		if err != nil {
			log.Printf("%+v", err)
			return
		}
	}

	if icLoTW != nil {
		err := icLoTW.SetImage(statusImage(statuses[tasks.TaskQSLLoTW]))
		if err != nil {
			log.Printf("%+v", err)
			return
		}
	}

	if icQRZ != nil {
		err := icQRZ.SetImage(statusImage(statuses[tasks.TaskQSLQRZ]))
		if err != nil {
			log.Printf("%+v", err)
			return
		}
	}

	if icClubLog != nil {
		err := icClubLog.SetImage(statusImage(statuses[tasks.TaskQSLClubLog]))
		if err != nil {
			log.Printf("%+v", err)
			return
		}
	}

	if icHamAlert != nil {
		err := icHamAlert.SetImage(statusImage(statuses[tasks.TaskHamAlert]))
		if err != nil {
			log.Printf("%+v", err)
			return
		}
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

	imgNotRunning, err = walk.NewIconFromImageForDPI(util.GenerateStatusImage(color.RGBA{R: 128, G: 128, B: 128, A: 255}), 96)
	if err != nil {
		log.Printf("%+v", err)
	}

	c := declarative.Composite{
		Layout: declarative.HBox{MarginsZero: true},
		Children: []declarative.Widget{
			declarative.ImageView{
				Image:       imgNotRunning,
				AssignTo:    &icSourceFiles,
				ToolTipText: "Source Files",
			},
			declarative.ImageView{
				Image:       imgNotRunning,
				AssignTo:    &icLoTW,
				ToolTipText: "LoTW",
			},
			declarative.ImageView{
				Image:       imgNotRunning,
				AssignTo:    &icQRZ,
				ToolTipText: "QRZ",
			},
			declarative.ImageView{
				Image:       imgNotRunning,
				AssignTo:    &icClubLog,
				ToolTipText: "Club Log",
			},
			declarative.ImageView{
				Image:       imgNotRunning,
				AssignTo:    &icHamAlert,
				ToolTipText: "HamAlert",
			},
		},
	}

	tasks.Attach(updateStatuses)

	return c
}

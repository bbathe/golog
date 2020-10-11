package ui

import (
	"github.com/lxn/walk"
)

// MsgError displays dialog to user with error details
func MsgError(p *walk.MainWindow, err error) {
	if p == nil {
		walk.MsgBox(nil, appName, err.Error(), walk.MsgBoxIconError|walk.MsgBoxServiceNotification)
	} else {
		walk.MsgBox(p, appName, err.Error(), walk.MsgBoxIconError)
	}
}

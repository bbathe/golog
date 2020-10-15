package ui

import (
	"github.com/lxn/walk"
)

// MsgError displays dialog to user with error details
func MsgError(p walk.Form, err error) {
	if p == nil {
		walk.MsgBox(nil, appName, err.Error(), walk.MsgBoxIconError|walk.MsgBoxServiceNotification)
	} else {
		walk.MsgBox(p, appName, err.Error(), walk.MsgBoxIconError)
	}
}

// MsgInformation displays dialog to user with non-error information
func MsgInformation(p walk.Form, info string) {
	if p == nil {
		walk.MsgBox(nil, appName, info, walk.MsgBoxIconInformation|walk.MsgBoxServiceNotification)
	} else {
		walk.MsgBox(p, appName, info, walk.MsgBoxIconInformation)
	}
}

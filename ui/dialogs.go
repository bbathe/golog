package ui

import (
	"log"
	"os"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

type FileFilter struct {
	Description string
	Wildcard    string
}

// setupOPENFILENAME is a helper function to initialize the OPENFILENAME structure
func setupOPENFILENAME(filters []FileFilter, file []uint16, fileLen int) (*win.OPENFILENAME, error) {
	// start at current working directory
	p, err := os.Getwd()
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return nil, err
	}

	// convert it to UTF16 string
	pwdLen := 1024
	pwd := make([]uint16, pwdLen+1)

	pwdU16, err := syscall.UTF16FromString(p)
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return nil, err
	}
	copy(pwd, pwdU16)

	// convert filters to UTF16 string
	filter := make([]uint16, 1024)
	var i int
	for _, f := range filters {
		fu16, err := syscall.UTF16FromString(f.Description)
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return nil, err
		}

		// copy over bytes
		for _, u16 := range fu16 {
			filter[i] = u16
			i++
		}

		fu16, err = syscall.UTF16FromString(f.Wildcard)
		if err != nil {
			MsgError(nil, err)
			log.Printf("%+v", err)
			return nil, err
		}

		// copy over bytes
		for _, u16 := range fu16 {
			filter[i] = u16
			i++
		}
	}

	ofn := win.OPENFILENAME{
		HwndOwner:       win.HWND(0),
		LpstrFile:       (*uint16)(unsafe.Pointer(&file[0])),
		NMaxFile:        uint32(fileLen),
		LpstrFilter:     (*uint16)(unsafe.Pointer(&filter[0])),
		NFilterIndex:    1,
		LpstrInitialDir: (*uint16)(unsafe.Pointer(&pwd[0])),
		Flags:           win.OFN_PATHMUSTEXIST | win.OFN_FILEMUSTEXIST,
	}
	ofn.LStructSize = uint32(unsafe.Sizeof(ofn))

	return &ofn, nil
}

// OpenFilePicker presents the open file dialog for the user to choose a file
func OpenFilePicker(filters []FileFilter) (*string, error) {
	// buffer for chosen filename to be put
	fileLen := 1024
	file := make([]uint16, fileLen+1)

	pofn, err := setupOPENFILENAME(filters, file, fileLen)
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return nil, err
	}

	if win.GetOpenFileName(pofn) {
		// convert to regular string for caller
		s := syscall.UTF16ToString(file)

		return &s, nil
	}

	// nothing chosen by user
	return nil, nil
}

// SaveFilePicker presents the save file dialog for the user to choose a file
func SaveFilePicker(filters []FileFilter) (*string, error) {
	// buffer for chosen filename to be put
	fileLen := 1024
	file := make([]uint16, fileLen+1)

	pofn, err := setupOPENFILENAME(filters, file, fileLen)
	if err != nil {
		MsgError(nil, err)
		log.Printf("%+v", err)
		return nil, err
	}

	if win.GetSaveFileName(pofn) {
		// convert to regular string for caller
		s := syscall.UTF16ToString(file)

		return &s, nil
	}

	// nothing chosen by user
	return nil, nil
}

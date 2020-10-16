package ui

import (
	"github.com/lxn/walk"
)

// OpenFilePicker presents the open file dialog for the user to choose a file
func OpenFilePicker(parent walk.Form, title, filters string) (*string, error) {
	dlg := new(walk.FileDialog)
	dlg.Filter = filters
	dlg.Title = title

	ok, err := dlg.ShowOpen(parent)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	return &dlg.FilePath, nil
}

// OpenFilePicker presents the open file dialog for the user to choose a file
func OpenFilePickerWithInitialDir(parent walk.Form, title, filters string, initialdirpath string) (*string, error) {
	dlg := new(walk.FileDialog)
	dlg.Filter = filters
	dlg.Title = title
	dlg.InitialDirPath = initialdirpath

	ok, err := dlg.ShowOpen(parent)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	return &dlg.FilePath, nil
}

// SaveFilePicker presents the save file dialog for the user to choose a file
func SaveFilePicker(parent walk.Form, title, filters string) (*string, error) {
	dlg := new(walk.FileDialog)
	dlg.Filter = filters
	dlg.Title = title

	ok, err := dlg.ShowSave(parent)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	return &dlg.FilePath, nil
}

// OpenFolderPicker presents the browse for folder dialog for the user to choose a folder
func OpenFolderPicker(parent walk.Form, title string) (*string, error) {
	dlg := new(walk.FileDialog)
	dlg.Title = title

	ok, err := dlg.ShowBrowseFolder(parent)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	return &dlg.FilePath, nil
}

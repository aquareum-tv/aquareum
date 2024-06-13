package app

import (
	"embed"
	"io/fs"
)

//go:embed all:dist/**
var files embed.FS

// fetch a static snapshot of the current Aquareum web app
func Files() (fs.FS, error) {
	rootFiles, err := fs.Sub(files, "dist")
	if err != nil {
		return nil, err
	}
	return rootFiles, nil
}

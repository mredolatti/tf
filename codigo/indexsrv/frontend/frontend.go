package frontend

import (
	"embed"
)

//go:embed assets/*
var assets embed.FS

// Index returns the contents (in bytes) of index.html
func Index() []byte {
	data, err := assets.ReadFile("assets/index.html")
	if err != nil {
		panic("Error fetching static asset index.html: " + err.Error())
	}
	return data
}

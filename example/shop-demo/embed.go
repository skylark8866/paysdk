package main

import (
	"embed"
	"html/template"
)

//go:embed templates/*
var templatesFS embed.FS

func loadTemplates() (*template.Template, error) {
	return template.ParseFS(templatesFS, "templates/*")
}

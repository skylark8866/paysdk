package main

import (
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*
var templatesFS embed.FS

func yuan(cents int64) string {
	return fmt.Sprintf("%.2f", float64(cents)/100)
}

func loadTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"yuan": yuan,
	}
	return template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*")
}

package main

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates/*.html
var embeddedTemplates embed.FS

// EmbeddedTemplate serves the templates compiled into the binary, parsed once at
// startup, so every deploy carries exactly the templates it was built with.
type EmbeddedTemplate struct {
	tmpl *template.Template
}

func NewEmbeddedTemplate() (*EmbeddedTemplate, error) {
	sub, err := fs.Sub(embeddedTemplates, "templates")
	if err != nil {
		return nil, err
	}
	tmpl, err := parseTemplates(sub)
	if err != nil {
		return nil, err
	}
	return &EmbeddedTemplate{tmpl: tmpl}, nil
}

func (e *EmbeddedTemplate) GetTemplate() *template.Template {
	return e.tmpl
}

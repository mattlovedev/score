package main

import (
	"html/template"
	"log"
	"os"
)

// LocalTemplate re-reads ./templates on every request, so template edits show up
// without restarting the server.
type LocalTemplate struct{}

func (t *LocalTemplate) GetTemplate() *template.Template {
	tmpl, err := parseTemplates(os.DirFS("templates"))
	if err != nil {
		log.Fatal(err)
	}
	return tmpl
}

package main

import (
	"html/template"
	"io/fs"
)

type TemplateGetter interface {
	GetTemplate() *template.Template
}

// parseTemplates parses every .html file in fsys. Each file is available by its
// filename (e.g. "homepage.html") and by any {{define}} names inside it.
func parseTemplates(fsys fs.FS) (*template.Template, error) {
	return template.ParseFS(fsys, "*.html")
}

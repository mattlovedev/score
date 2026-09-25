package main

import "html/template"

type TemplateGetter interface {
	GetTemplate() *template.Template
}

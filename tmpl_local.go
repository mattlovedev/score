package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
)

type LocalTemplate struct {
	//tmpl *template.Template
}

func (t *LocalTemplate) GetTemplate() *template.Template {
	/*if t.tmpl == nil {
		t.tmpl = template.Must(template.ParseGlob("../templates/*"))
	}
	return t.tmpl*/
	// since this is for testing not gonna cache
	ents, err := os.ReadDir("templates")
	if err != nil {
		log.Fatal(err)
	}
	tmpl := template.New("tmpl")
	for _, ent := range ents {
		data, err := os.ReadFile(filepath.Join("templates", ent.Name()))
		if err != nil {
			log.Fatal(err)
		}
		tmpl, err = tmpl.Parse(string(data))
		if err != nil {
			log.Fatalf("Parse: %s", err)
		}
	}
	return tmpl
	//return template.Must(template.ParseGlob("../templates/*"))
}

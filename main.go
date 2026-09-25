package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	_ "time/tzdata" // types.go loads America/New_York; don't depend on the image having tzdata
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	local := flag.Bool("local", false, "use ./templates (re-read per request) and ./storage instead of the embedded templates and Firestore")
	flag.Parse()

	var s Storage
	var t TemplateGetter
	if *local {
		s, t = &LocalStorage{}, &LocalTemplate{}
		log.Print("using local templates and storage")
	} else {
		project := getenv("SCORE_PROJECT", "mattlovedev-apps")
		database := getenv("SCORE_DATABASE", "score")

		fs, err := NewFirestoreStorage(context.Background(), project, database)
		if err != nil {
			log.Fatalf("firestore %s/%s: %v", project, database, err)
		}
		et, err := NewEmbeddedTemplate()
		if err != nil {
			log.Fatalf("templates: %v", err)
		}
		s, t = fs, et
		log.Printf("using firestore %s/%s and embedded templates", project, database)
	}

	port := getenv("PORT", "8080")
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, newRouter(s, t)); err != nil {
		log.Fatal(err)
	}
}

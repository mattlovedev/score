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
	local := flag.Bool("local", false, "use ./templates and ./storage instead of GCS and Firestore")
	flag.Parse()

	var s Storage
	var t TemplateGetter
	if *local {
		s, t = &LocalStorage{}, &LocalTemplate{}
		log.Print("using local templates and storage")
	} else {
		ctx := context.Background()
		project := getenv("SCORE_PROJECT", "mattlovedev-apps")
		database := getenv("SCORE_DATABASE", "score")
		bucket := getenv("SCORE_TEMPLATES_BUCKET", "mattlovedev-apps-score-templates")

		fs, err := NewFirestoreStorage(ctx, project, database)
		if err != nil {
			log.Fatalf("firestore %s/%s: %v", project, database, err)
		}
		gt, err := NewGcsTemplate(ctx, bucket)
		if err != nil {
			log.Fatalf("templates gs://%s: %v", bucket, err)
		}
		s, t = fs, gt
		log.Printf("using firestore %s/%s and templates gs://%s", project, database, bucket)
	}

	port := getenv("PORT", "8080")
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, newRouter(s, t)); err != nil {
		log.Fatal(err)
	}
}

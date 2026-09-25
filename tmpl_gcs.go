package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type GcsTemplate struct {
	tmpl *template.Template
}

// NewGcsTemplate parses every object in the bucket once, at startup.
func NewGcsTemplate(ctx context.Context, bucketName string) (*GcsTemplate, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage client: %w", err)
	}
	defer func() { _ = client.Close() }()
	it := client.Bucket(bucketName).Objects(ctx, nil)
	tmpl := template.New("tmpl")
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterate objects: %w", err)
		}
		rc, err := client.Bucket(bucketName).Object(attrs.Name).NewReader(ctx)
		if err != nil {
			return nil, fmt.Errorf("NewReader %s: %w", attrs.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, fmt.Errorf("ReadAll %s: %w", attrs.Name, err)
		}

		tmpl, err = tmpl.Parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("Parse %s: %w", attrs.Name, err)
		}
	}
	return &GcsTemplate{tmpl: tmpl}, nil
}

func (g *GcsTemplate) GetTemplate() *template.Template {
	return g.tmpl
}

package main

import (
	"context"
	"errors"
	"reflect"

	"cloud.google.com/go/firestore"
)

type FirestoreStorage struct {
	firestoreClient *firestore.Client
}

func NewFirestoreStorage(ctx context.Context, projectID string, database string) (*FirestoreStorage, error) {
	client, err := firestore.NewClientWithDatabase(ctx, projectID, database)
	if err != nil {
		return nil, err
	}
	return &FirestoreStorage{firestoreClient: client}, nil
}

func (f *FirestoreStorage) getClient() *firestore.Client {
	return f.firestoreClient
}

func (f *FirestoreStorage) Add(collection Collection, id string, value any) error {
	_, err := f.getClient().Collection(string(collection)).Doc(id).Set(context.Background(), value)
	return err
}

func (f *FirestoreStorage) Delete(collection Collection, id string) error {
	_ /*writeResult*/, err := f.getClient().Collection(string(collection)).Doc(id).Delete(context.Background())
	return err
}

func (f *FirestoreStorage) Get(collection Collection, id string, val any) error {
	resultValue := reflect.ValueOf(val)
	if resultValue.Kind() != reflect.Ptr {
		return errors.New("result must be a pointer to a struct")
	}

	resultType := resultValue.Elem().Type()
	resultPointer := reflect.New(resultType)

	doc, err := f.getClient().Collection(string(collection)).Doc(id).Get(context.Background())
	if err != nil {
		return err
	}

	if err = doc.DataTo(resultPointer.Interface()); err != nil {
		return err
	}

	resultValue.Elem().Set(resultPointer.Elem())
	return nil
}

func (f *FirestoreStorage) Query(collection Collection, query Query, vals any) error {
	resultValue, resultType, err := getResultValAndType(vals)
	if err != nil {
		return err
	}
	q := f.getClient().Collection(string(collection)).Query
	for _, w := range query.wheres {
		q = q.Where(w.Path, w.Op, w.Val)
	}
	if query.order.Path != "" {
		q = q.OrderBy(query.order.Path, firestore.Direction(query.order.Direction))
	}
	if query.limit != 0 {
		q = q.Limit(query.limit)
	}
	docs, err := q.Documents(context.Background()).GetAll()
	if err != nil {
		return err
	}
	resultSlice := reflect.MakeSlice(resultType, len(docs), len(docs))
	for i, doc := range docs {
		resultSliceIndexedPointerInterface := resultSlice.Index(i).Addr().Interface()
		if err = doc.DataTo(resultSliceIndexedPointerInterface); err != nil {
			return err
		}
	}
	resultValue.Elem().Set(resultSlice)
	return nil
}

func (f *FirestoreStorage) Set(collection Collection, id string, value any) error {
	_, err := f.getClient().Collection(string(collection)).Doc(id).Set(context.Background(), value)
	return err
}

func (f *FirestoreStorage) Count(collection Collection, query Query) (int, error) {
	return 0, nil
}

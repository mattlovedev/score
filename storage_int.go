package main

import (
	"errors"
	"io/fs"
	"math/rand"
	"reflect"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomString(length int) string {
	seededRand := rand.New(rand.NewSource(timeNow().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func getResultValAndType(p interface{}) (reflect.Value, reflect.Type, error) {
	resultValue := reflect.ValueOf(p)
	if resultValue.Kind() != reflect.Ptr {
		return reflect.Value{}, nil, errors.New("result must be a pointer of a slice not a " + resultValue.Kind().String())
	}
	resultType := resultValue.Elem().Type()
	if resultType.Kind() != reflect.Slice {
		return reflect.Value{}, nil, errors.New("result must be a pointer of a slice not a " + resultType.Kind().String())
	}
	return resultValue, resultType, nil
}

type Where struct {
	Path string
	Op   string
	Val  any
}

type Direction firestore.Direction

const (
	Asc  Direction = Direction(firestore.Asc)
	Desc Direction = Direction(firestore.Desc)
)

type Collection string

const (
	activeGamesCollection   Collection = "activeGames"
	finishedGamesCollection Collection = "finishedGames"
)

type Order struct {
	Path      string
	Direction Direction
}

type Query struct {
	wheres []Where
	limit  int
	order  Order
}

func (q Query) Where(path string, op string, val any) Query {
	q.wheres = append(q.wheres, Where{Path: path, Op: op, Val: val})
	return q
}

func (q Query) Limit(limit int) Query {
	q.limit = limit
	return q
}

func (q Query) Order(path string, direction Direction) Query {
	q.order = Order{Path: path, Direction: direction}
	return q
}

func NewQuery() Query {
	return Query{}
}

type Storage interface {
	Add(collection Collection, id string, value any) error
	Delete(collection Collection, id string) error
	Get(collection Collection, id string, val any) error
	Query(collection Collection, query Query, vals any) error
	Set(collection Collection, id string, value any) error
	Count(collection Collection, query Query) (int, error)
	// RunTransaction runs fn atomically. fn may be retried, so it must not keep
	// state between calls. All Gets must happen before any Set or Delete.
	RunTransaction(fn func(tx Tx) error) error
}

// isNotFound reports whether err means the document doesn't exist, for either
// storage backend.
func isNotFound(err error) bool {
	return status.Code(err) == codes.NotFound || errors.Is(err, fs.ErrNotExist)
}

// Tx is the subset of Storage available inside a transaction.
type Tx interface {
	Get(collection Collection, id string, val any) error
	Set(collection Collection, id string, value any) error
	Delete(collection Collection, id string) error
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"time"
)

type LocalStorage struct{}

func (l LocalStorage) Add(collection Collection, id string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(fmt.Sprintf("storage/%s/%s", collection, id), b, 0644); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err = os.MkdirAll(fmt.Sprintf("storage/%s", collection), 0755); err != nil {
				return err
			}
			return l.Add(collection, id, value)
		}
		return err
	}
	return nil
}

func (l LocalStorage) Delete(collection Collection, id string) error {
	return os.Remove(fmt.Sprintf("storage/%s/%s", collection, id))
}

func (l LocalStorage) Get(collection Collection, id string, val any) error {
	resultValue := reflect.ValueOf(val)
	if resultValue.Kind() != reflect.Ptr {
		return errors.New("result must be a pointer to a struct")
	}

	fileData, err := os.ReadFile(fmt.Sprintf("storage/%s/%s", collection, id))
	if err != nil {
		return err
	}

	resultType := resultValue.Elem().Type()
	resultPointer := reflect.New(resultType)

	if err = json.Unmarshal(fileData, resultPointer.Interface()); err != nil {
		return err
	}

	resultValue.Elem().Set(resultPointer.Elem())
	return nil
}

func (l LocalStorage) Query(collection Collection, query Query, vals any) error {

	ents, err := os.ReadDir(fmt.Sprintf("storage/%s", collection))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	resultValue, resultType, err := getResultValAndType(vals)
	if err != nil {
		return err
	}
	resultSlice := reflect.MakeSlice(resultType, 0, len(ents))

	peType := resultType.Elem()
	peZero := reflect.Zero(peType)
	pe := reflect.New(peType)

	for _, ent := range ents {
		pe.Elem().Set(peZero)
		if b, err := os.ReadFile(fmt.Sprintf("storage/%s/%s", collection, ent.Name())); err != nil {
			return err
		} else if err = json.Unmarshal(b, pe.Interface()); err != nil {
			return err
		}
		adding := true
		for _, w := range query.wheres {
			switch w.Op {
			case "==":
				if pe.Elem().FieldByName(w.Path).Interface() != w.Val {
					adding = false
					break
				}
			case "!=":
				if pe.Elem().FieldByName(w.Path).Interface() == w.Val {
					adding = false
					break
				}
			}
		}
		if adding {
			resultSlice = reflect.Append(resultSlice, pe.Elem())
		}
	}
	if query.order.Path != "" && resultSlice.Len() > 1 {
		fieldType := reflect.TypeOf(resultSlice.Index(0).FieldByName(query.order.Path).Interface()).String()
		switch fieldType {
		case "time.Time":
			if query.order.Direction == Asc {
				sort.Slice(resultSlice.Interface(), func(i int, j int) bool {
					return resultSlice.Index(i).FieldByName(query.order.Path).Interface().(time.Time).Before(
						resultSlice.Index(j).FieldByName(query.order.Path).Interface().(time.Time))
				})
			} else {
				sort.Slice(resultSlice.Interface(), func(i int, j int) bool {
					return resultSlice.Index(j).FieldByName(query.order.Path).Interface().(time.Time).Before(
						resultSlice.Index(i).FieldByName(query.order.Path).Interface().(time.Time))
				})
			}
		default:
			return errors.New("incompatible field type: " + fieldType)
		}
	}

	if query.limit != 0 && resultSlice.Len() > query.limit {
		limitedResultSlice := reflect.MakeSlice(resultType, query.limit, query.limit)
		for i := 0; i < query.limit; i++ {
			limitedResultSlice.Index(i).Set(resultSlice.Index(i))
		}
		resultValue.Elem().Set(limitedResultSlice)
	} else {
		resultValue.Elem().Set(resultSlice)
	}
	return nil
}

func (l LocalStorage) Set(collection Collection, id string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(fmt.Sprintf("storage/%s/%s", collection, id), b, 0644); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err = os.MkdirAll(fmt.Sprintf("storage/%s", collection), 0755); err != nil {
				return err
			}
			return l.Add(collection, id, value)
		}
		return err
	}
	return nil
}

func (l LocalStorage) Count(collection Collection, query Query) (int, error) {
	return 0, nil
}

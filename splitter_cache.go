package sqlg

import (
	"errors"
	"reflect"
)

type splitterChacheRecord struct {
	Keys   []string
	Values []interface{}
}

type splitterCache map[interface{}]splitterChacheRecord

func (o splitterCache) Set(ptr interface{}, keys []string, vals []interface{}) error {
	if reflect.ValueOf(ptr).Kind() != reflect.Pointer {
		return errors.New("Ptr must be pointer type.")
	}

	o[ptr] = splitterChacheRecord{keys, vals}

	return nil
}

func (o splitterCache) Get(ptr interface{}) ([]string, []interface{}) {
	if scr, ok := o[ptr]; ok {
		return scr.Keys, scr.Values
	}

	return nil, nil
}

func newSplitterCache() splitterCache {
	return make(splitterCache)
}

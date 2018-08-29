package sqlg

import (
	"errors"
	"reflect"
)

type StructSplitter struct {
	Exclude []string
	Tag     string
}

func (ss *StructSplitter) Split(i interface{}) (resKeys []string, resVals []interface{}, err error) {
	val := reflect.ValueOf(i)

	switch val.Kind() {
	case reflect.Map:
		keys := val.MapKeys()

	MapOuterLoop:
		for _, kv := range keys {
			if kv.Kind() != reflect.String {
				err = errors.New("Map component must have a string keys!")
				return
			}

			key := kv.String()

			// filter out excluded fields
			if len(ss.Exclude) > 0 {
				for i := 0; i < len(ss.Exclude); i++ {
					if ss.Exclude[i] == key {
						continue MapOuterLoop
					}
				}
			}

			resKeys = append(resKeys, key)
			resVals = append(resVals, val.MapIndex(kv).Interface())
		}

	case reflect.Struct:
		numFields := val.NumField()

		if numFields == 0 {
			err = errors.New("Struct has no fields!")
			return
		}

		tpOf := reflect.TypeOf(i)

	StructOuterLoop:
		for i := 0; i < numFields; i++ {
			f := val.Field(i)
			tf := tpOf.Field(i)

			// skip private fields
			if tf.PkgPath != "" {
				continue
			}

			key := tf.Name

			// check if we have tag
			if ss.Tag != "" {
				if tag := tf.Tag.Get(ss.Tag); tag != "" {
					if tag == "-" {
						continue StructOuterLoop
					}

					key = tag
				}
			}

			// filter out excluded fields
			if len(ss.Exclude) > 0 {
				for i := 0; i < len(ss.Exclude); i++ {
					if ss.Exclude[i] == key {
						continue StructOuterLoop
					}
				}
			}

			resKeys = append(resKeys, key)
			resVals = append(resVals, f.Interface())
		}

	default:
		err = errors.New("Component must be map[string]... or struct type")
	}

	return
}

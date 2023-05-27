package sqlg

import (
	"errors"
	"reflect"
)

var splitter *Splitter

func init() {
	splitter = NewSplitter().KeyModifier(ToSnake)
}

type Splitter struct {
	keyModifier func(string) string
	tag         string
}

func (ss *Splitter) Split(i interface{}, exclude []string) (resKeys []string, resVals []interface{}, err error) {
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

			if ss.keyModifier != nil {
				key = ss.keyModifier(key)
			}

			// filter out excluded fields
			if len(exclude) > 0 {
				for i := 0; i < len(exclude); i++ {
					if exclude[i] == key {
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

			// if field is embeded struct
			if f.Kind() == reflect.Struct && tf.Anonymous {
				if esKeys, esVals, esErr := ss.Split(f.Interface(), exclude); esErr != nil {
					err = esErr
					return
				} else {
					resKeys = append(resKeys, esKeys...)
					resVals = append(resVals, esVals...)
				}

				continue
			}

			var key string

			// check if we have tag
			if tag := tf.Tag.Get(ss.tag); tag != "" {
				if tag == "-" {
					continue StructOuterLoop
				}

				key = tag
			} else if ss.keyModifier != nil {
				key = ss.keyModifier(tf.Name)
			} else {
				key = tf.Name
			}

			// filter out excluded fields
			if len(exclude) > 0 {
				for i := 0; i < len(exclude); i++ {
					if exclude[i] == key {
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

func (ss *Splitter) KeyModifier(km func(string) string) *Splitter {
	ss.keyModifier = km
	return ss
}

func (ss *Splitter) Tag(t string) *Splitter {
	ss.tag = t
	return ss
}

func NewSplitter() *Splitter {
	return &Splitter{tag: "sgsp"}
}

func GetSplitter() *Splitter {
	return splitter
}

func Split(i interface{}, exclude []string) (resKeys []string, resVals []interface{}, err error) {
	return splitter.Split(i, exclude)
}

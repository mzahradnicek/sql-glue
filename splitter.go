package sqlg

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

var (
	ErrMapNoStringKeys = errors.New("Map component must have a string keys!")
	ErrStructNoFields  = errors.New("Struct has no fields!")
	ErrBadType         = errors.New("Component must be map[string]... or struct type")
	ErrNilInput        = errors.New("Input is nil")
)

// Splitter is a configuration tool for splitting a structure or map into a list of keys and values.
type Splitter struct {
	keyModifier func(string) string
	tag         string
}

// Split splits the input value (structure or map) into keys and values.
// The `exclude` field is used to exclude specific fields.
func (ss *Splitter) Split(i interface{}, exclude ...string) (resKeys []string, resVals []interface{}, err error) {
	if i == nil {
		return nil, nil, ErrNilInput
	}

	val := reflect.ValueOf(i)

	if val.Kind() == reflect.Pointer && !val.IsNil() {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		return ss.splitMap(val, exclude)
	case reflect.Struct:
		return ss.splitStruct(val, exclude)
	default:
		return nil, nil, fmt.Errorf("%w: got %s", ErrBadType, val.Kind())
	}
}

// splitMap process maps
func (ss *Splitter) splitMap(val reflect.Value, exclude []string) (resKeys []string, resVals []interface{}, err error) {
	keys := val.MapKeys()

	for _, kv := range keys {
		if kv.Kind() != reflect.String {
			return nil, nil, ErrMapNoStringKeys
		}

		key := kv.String()

		// check if field is excluded
		if slices.Contains(exclude, key) {
			continue
		}

		if ss.keyModifier != nil {
			key = ss.keyModifier(key)
		}

		resKeys = append(resKeys, key)
		resVals = append(resVals, val.MapIndex(kv).Interface())
	}

	return resKeys, resVals, nil
}

// splitStruct process structs
func (ss *Splitter) splitStruct(val reflect.Value, exclude []string) (resKeys []string, resVals []interface{}, err error) {
	numFields := val.NumField()
	if numFields == 0 {
		return nil, nil, ErrStructNoFields
	}

	tpOf := val.Type()

StructOuterLoop:
	for i := 0; i < numFields; i++ {
		f := val.Field(i)
		tf := tpOf.Field(i)

		// skip private fields
		if tf.PkgPath != "" || slices.Contains(exclude, tf.Name) {
			continue
		}

		var key string

		// process tag
		if tag, _ := tf.Tag.Lookup(ss.tag); tag != "" {
			if tag == "-" {
				continue
			}

			fkey, opts, _ := strings.Cut(tag, ",")

			// check exclude groups in tags - little ugly, but maybe more performant way than use slices.Split...
			if len(opts) > 0 {
				s := string(opts)
				for s != "" {
					var opt string
					opt, s, _ = strings.Cut(s, ",")
					if slices.Contains(exclude, opt) {
						continue StructOuterLoop
					}
				}
			}

			if fkey != "" {
				key = fkey
			}
		}

		if key == "" {
			if ss.keyModifier != nil {
				key = ss.keyModifier(tf.Name)
			} else {
				key = tf.Name
			}
		}

		// process nested struct
		if f.Kind() == reflect.Struct {
			if !reflect.PointerTo(f.Type()).Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
				nestedKeys, nestedVals, nestedErr := ss.Split(f.Interface(), exclude...)
				if nestedErr != nil {
					return nil, nil, nestedErr
				}

				// add struct name prefix
				for j, l := 0, len(nestedKeys); j < l; j++ {
					nestedKeys[j] = key + "." + nestedKeys[j]
				}

				resKeys = append(resKeys, nestedKeys...)
				resVals = append(resVals, nestedVals...)
				continue
			}
		}

		resKeys = append(resKeys, key)
		resVals = append(resVals, f.Interface())

	}

	return resKeys, resVals, nil
}

// KeyModifier sets a function that modifies the key name.
func (ss *Splitter) KeyModifier(km func(string) string) *Splitter {
	ss.keyModifier = km
	return ss
}

// Tag sets the name of the tag in the structure to be used for naming keys.
func (ss *Splitter) Tag(t string) *Splitter {
	ss.tag = t
	return ss
}

// NewSplitter returns a new Splitter with the default tag "sqlg".
func NewSplitter() *Splitter {
	return &Splitter{tag: "sqlg"}
}

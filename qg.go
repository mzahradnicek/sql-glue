package sqlg

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
)

type qgConfig struct {
	IdentifierEscape func(string) string
	KeyModifier      func(string) string
	Placeholder      func(buf *bytes.Buffer)
	Tag              string
}

type Qg []interface{}

func (qg *Qg) Append(chunks ...interface{}) {
	*qg = append(*qg, chunks...)
}

func (qg *Qg) ToSql(cfg *Config) (string, []interface{}, error) {
	res, args, err := qg.Compile(&qgConfig{
		IdentifierEscape: cfg.IdentifierEscape,
		KeyModifier:      cfg.KeyModifier,
		Placeholder:      cfg.PlaceholderInit(),
		Tag:              cfg.Tag,
	})

	return strings.Join(res, " "), args, err
}

func (qg *Qg) Compile(cfg *qgConfig) ([]string, []interface{}, error) {
	var ressql []string
	var resargs []interface{}

	for qi, qend := 0, len(*qg); qi < qend; {
		chunk := &bytes.Buffer{}

		switch qval := (*qg)[qi].(type) {
		case Qg:
			sql, args, err := qval.Compile(cfg)
			if err != nil {
				return []string{}, []interface{}{}, err
			}

			resargs = append(resargs, args...)

			chunk.WriteString(strings.Join(sql, " "))
		case string:
			for i, end := 0, len(qval); i < end; {
				lasti := i
				for i < end && qval[i] != '%' {
					i++
				}

				if i > lasti {
					chunk.WriteString(qval[lasti:i])
				}

				if i >= end {
					break
				}

				i++

				var flagBuf bytes.Buffer

				for i < end && (qval[i] == '%' || (qval[i] >= 'a' && qval[i] <= 'z')) {
					flagBuf.WriteByte(qval[i])
					i++
				}

				flag := flagBuf.String()

				switch flag {
				case "v": // escaped value
					qi++
					// make placeholder
					cfg.Placeholder(chunk)

					// add value to args
					resargs = append(resargs, (*qg)[qi])

				case "and", "or":
					qi++
					switch aoval := (*qg)[qi].(type) {
					case Qg:
						sql, args, err := aoval.Compile(cfg)
						if err != nil {
							return []string{}, []interface{}{}, err
						}

						resargs = append(resargs, args...)

						chunk.WriteString("(" + strings.Join(sql, " "+strings.ToUpper(flag)+" ") + ")")
					default:
						return []string{}, []interface{}{}, errors.New("Component must be Qg type")
					}

				case "sp":
					qi++

					val := reflect.ValueOf((*qg)[qi])
					var placeholderNum = 0

					switch val.Kind() {
					case reflect.Map, reflect.Struct:
						ss := NewSplitter().KeyModifier(cfg.KeyModifier).Tag(cfg.Tag)
						_, vals, err := ss.Split((*qg)[qi], nil)

						if err != nil {
							return []string{}, []interface{}{}, err
						}

						resargs = append(resargs, vals...)
						placeholderNum = len(vals)
					case reflect.Array, reflect.Slice:
						if val.Type().String() == "sqlg.Qg" {
							return []string{}, []interface{}{}, errors.New("Component cant be sqlg.Qg")
						}

						for i := 0; i < val.Len(); i++ {
							resargs = append(resargs, val.Index(i).Interface())
						}

						placeholderNum = val.Len()
					default:
						return []string{}, []interface{}{}, errors.New("Component must be map, struct or array type")
					}

					for i := 0; i < placeholderNum; i++ {
						if i > 0 {
							chunk.WriteString(", ")
						}

						cfg.Placeholder(chunk)
					}
				case "set":
					qi++

					val := reflect.ValueOf((*qg)[qi])

					switch val.Kind() {
					case reflect.Map, reflect.Struct:
						ss := NewSplitter().KeyModifier(cfg.KeyModifier).Tag(cfg.Tag)
						keys, vals, err := ss.Split((*qg)[qi], nil)

						if err != nil {
							return []string{}, []interface{}{}, err
						}

						for i := 0; i < len(vals); i++ {
							if i > 0 {
								chunk.WriteString(", ")
							}

							// write key
							if cfg.IdentifierEscape != nil {
								chunk.WriteString(cfg.IdentifierEscape(keys[i]))
							} else {
								chunk.WriteString(keys[i])
							}

							chunk.WriteString(" = ")

							// write value
							cfg.Placeholder(chunk)
						}

						resargs = append(resargs, vals...)

					default:
						return []string{}, []interface{}{}, errors.New("Component must be map or struct type")
					}
				case "%": // escaped %%
					chunk.WriteByte('%')
				}
			}

		default:
			return []string{}, []interface{}{}, errors.New("Component must be string or another Qg")
		}
		qi++
		ressql = append(ressql, chunk.String())
	}

	return ressql, resargs, nil
}

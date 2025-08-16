package sqlg

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type Qg []interface{}

// Excluded field list
type Qe []string

func (qg *Qg) Append(chunks ...interface{}) {
	*qg = append(*qg, chunks...)
}

func (qg Qg) Process(cfg *Config) (string, []interface{}, error) {
	placeholderWriter := cfg.PlaceholderInit()
	resSql, resArgs, err := qg.compile(cfg, placeholderWriter)

	return strings.Join(resSql, " "), resArgs, err
}

func (qg Qg) compile(cfg *Config, placeholderWriter func(buf *bytes.Buffer)) (resSql []string, resArgs []interface{}, err error) {
	splitterCache := newSplitterCache()
	for qi, qlen := 0, len(qg); qi < qlen; {
		chunk := &bytes.Buffer{}

		getNextEl := func() interface{} {
			if qi+1 >= qlen {
				return nil
			}

			qi++

			return qg[qi]
		}

		// if we have pointer, replace it with type
		qElem, _ := resolveElemType(qg[qi], reflect.Struct, reflect.String)
		if qElem != nil {
			qg[qi] = qElem
		}

		switch qval := qg[qi].(type) {
		case Qg:
			sql, args, err := qval.compile(cfg, placeholderWriter)
			if err != nil {
				return nil, nil, err
			}

			resArgs = append(resArgs, args...)

			chunk.WriteString(strings.Join(sql, " "))
		case string: // template processing
			for sPos, sLen := 0, len(qval); sPos < sLen; {
				chunkStartPos := sPos
				for sPos < sLen && qval[sPos] != '%' {
					sPos++
				}

				if sPos > chunkStartPos {
					chunk.WriteString(qval[chunkStartPos:sPos])
				}

				if sPos >= sLen {
					break
				}

				// move pointer to first character of expression
				sPos++

				// read expression
				var expBuf bytes.Buffer

				for sPos < sLen && (qval[sPos] == '%' || (qval[sPos] >= 'a' && qval[sPos] <= 'z')) {
					expBuf.WriteByte(qval[sPos])
					sPos++
				}

				exp := expBuf.String()

				switch exp {
				case "v": // escaped value
					nextEl := getNextEl()
					if nextEl == nil {
						return nil, nil, fmt.Errorf(`Missing field or nil value - #%d '%s'`, qi-1, errorChunkPreview(qval, sPos+len(exp)))
					}

					// make placeholder
					placeholderWriter(chunk)

					// add value to args
					resArgs = append(resArgs, nextEl)
				case "and", "or":
					nextEl := getNextEl()
					if nextEl == nil {
						return nil, nil, fmt.Errorf(`Missing field or nil value - #%d '%s'`, qi-1, errorChunkPreview(qval, sPos+len(exp)))
					}

					switch aoVal := nextEl.(type) {
					case Qg:
						sql, args, err := aoVal.compile(cfg, placeholderWriter)
						if err != nil {
							return nil, nil, err
						}

						resArgs = append(resArgs, args...)

						chunk.WriteString("(" + strings.Join(sql, " "+strings.ToUpper(exp)+" ") + ")")
					default:
						return nil, nil, fmt.Errorf(`Component must be Qg type - #%d '%s' it is %T`, qi, errorChunkPreview(qval, sPos+len(exp)), nextEl)
					}
				case "keys": // process keys of struct or map
					nextEl := getNextEl()
					if nextEl == nil {
						return nil, nil, fmt.Errorf(`Missing field or nil value - #%d '%s'`, qi-1, errorChunkPreview(qval, sPos+len(exp)))
					}

					el, ptr := resolveElemType(nextEl, reflect.Struct, reflect.Map)
					if el == nil {
						return nil, nil, fmt.Errorf(`Component must be Map or Struct type - #%d '%s' it is %T`, qi, errorChunkPreview(qval, sPos+len(exp)), nextEl)
					}

					// check if next element is Qe
					var qExclude Qe = nil
					if qi+1 < qlen {
						if v, ok := qg[qi+1].(Qe); ok {
							qExclude = v
							qi++
						}
					}

					// check cache
					var keys []string
					keys, _ = splitterCache.Get(ptr)

					if keys == nil || len(qExclude) > 0 {
						if len(qExclude) == 0 {
							qExclude = append(qExclude, "kv")
						}

						var vals []interface{}
						ss := NewSplitter().KeyModifier(cfg.KeyModifier).Tag(cfg.Tag)
						keys, vals, err = ss.Split(nextEl, qExclude...)

						if err != nil {
							return nil, nil, fmt.Errorf("Splitter error near #%d '%s': %w", qi, errorChunkPreview(qval, sPos+len(exp)), err)
						}

						// cache if we have a pointer
						if ptr != nil {
							splitterCache.Set(ptr, keys, vals)
						}
					}

					for ki, kv := range keys {
						if ki > 0 {
							chunk.WriteString(", ")
						}
						chunk.WriteString(cfg.IdentifierEscape(kv))
					}
				case "vals":
					nextEl := getNextEl()
					if nextEl == nil {
						return nil, nil, fmt.Errorf(`Missing field or nil value - #%d '%s'`, qi-1, errorChunkPreview(qval, sPos+len(exp)))
					}

					var placeholderCnt = 0

					// process Struct or Map
					el, ptr := resolveElemType(nextEl, reflect.Struct, reflect.Map)
					if el != nil {
						// check if next element is Qe
						var qExclude Qe = nil
						if qi+1 < qlen {
							if v, ok := qg[qi+1].(Qe); ok {
								qExclude = v
								qi++
							}
						}

						// check cache
						var vals []interface{}
						_, vals = splitterCache.Get(ptr)

						if vals == nil || len(qExclude) > 0 {
							if len(qExclude) == 0 {
								qExclude = append(qExclude, "kv")
							}

							var keys []string
							ss := NewSplitter().KeyModifier(cfg.KeyModifier).Tag(cfg.Tag)
							keys, vals, err = ss.Split(nextEl, qExclude...)

							if err != nil {
								return nil, nil, fmt.Errorf("Splitter error near #%d '%s': %w", qi, errorChunkPreview(qval, sPos+len(exp)), err)
							}

							// cache if we have a pointer
							if ptr != nil {
								splitterCache.Set(ptr, keys, vals)
							}
						}

						resArgs = append(resArgs, vals...)
						placeholderCnt = len(vals)
					} else if el, _ = resolveElemType(nextEl, reflect.Slice, reflect.Array); el != nil {
						val := reflect.ValueOf(el)
						if val.Type().String() == "sqlg.Qg" {
							return nil, nil, fmt.Errorf(`Component can't be Qg type - #%d '%s'`, qi, errorChunkPreview(qval, sPos+len(exp)))
						}

						for i := 0; i < val.Len(); i++ {
							resArgs = append(resArgs, val.Index(i).Interface())
						}

						placeholderCnt = val.Len()
					} else { // return error
						return nil, nil, fmt.Errorf(`Component must be Map, Struct, Slice or Array type - #%d '%s' it is %T`, qi, errorChunkPreview(qval, sPos+len(exp)), nextEl)
					}

					for i := 0; i < placeholderCnt; i++ {
						if i > 0 {
							chunk.WriteString(", ")
						}

						placeholderWriter(chunk)
					}
				case "set":
					nextEl := getNextEl()
					if nextEl == nil {
						return nil, nil, fmt.Errorf(`Missing field or nil value - #%d '%s'`, qi-1, errorChunkPreview(qval, sPos+len(exp)))
					}

					el, _ := resolveElemType(nextEl, reflect.Struct, reflect.Map)
					if el == nil {
						return nil, nil, fmt.Errorf(`Component must be Map or Struct type - #%d '%s' it is %T`, qi, errorChunkPreview(qval, sPos+len(exp)), nextEl)
					}

					// check if next element is Qe
					qExclude := Qe{"set"}
					if qi+1 < qlen {
						if v, ok := qg[qi+1].(Qe); ok {
							qExclude = v
							qi++
						}
					}

					ss := NewSplitter().KeyModifier(cfg.KeyModifier).Tag(cfg.Tag)
					keys, vals, err := ss.Split(nextEl, qExclude...)

					if err != nil {
						return nil, nil, fmt.Errorf("Splitter error near #%d '%s': %w", qi, errorChunkPreview(qval, sPos+len(exp)), err)
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
						placeholderWriter(chunk)
					}

					resArgs = append(resArgs, vals...)

				case "%": // escaped %%
					chunk.WriteByte('%')
				}
			}

		default:
			return nil, nil, fmt.Errorf("Component at #%d must be String or Qg type it is %T", qi, qval)
		}
		qi++
		resSql = append(resSql, chunk.String())
	}

	return resSql, resArgs, nil
}

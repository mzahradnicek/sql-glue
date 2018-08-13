package sqlg

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	IdentifierEscape func(string) string
	Placeholder      PlaceholderFormat
}

type qgConfig struct {
	IdentifierEscape func(string) string
	Placeholder      func(buf *bytes.Buffer)
}

func (c *Config) Glue(q Qg) (string, []interface{}, error) {
	return q.ToSql(&qgConfig{
		IdentifierEscape: c.IdentifierEscape,
		Placeholder:      c.Placeholder.GeneratePlaceholderFunc(),
	})
}

type Qg []interface{}

func (qg Qg) ToSql(cfg *qgConfig) (string, []interface{}, error) {
	res, args, err := qg.Compile(cfg)

	return strings.Join(res, " "), args, err
}

func (qg Qg) Compile(cfg *qgConfig) ([]string, []interface{}, error) {
	var ressql []string
	var resargs []interface{}

	for qi, qend := 0, len(qg); qi < qend; {
		chunk := &bytes.Buffer{}

		switch qval := qg[qi].(type) {
		case Qg:
			sql, args, err := qval.ToSql(cfg)
			if err != nil {
				return []string{}, []interface{}{}, err
			}

			resargs = append(resargs, args...)

			chunk.WriteString(sql)
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

				fmt.Printf("%d: %v\n", i, qval[i])
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
					resargs = append(resargs, qg[qi])

				case "and", "or":
					qi++
					switch aoval := qg[qi].(type) {
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

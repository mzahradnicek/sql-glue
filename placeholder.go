package sqlg

import (
	"bytes"
	"fmt"
)

// Numbering placeholder generator - PostgreSQL
func NrPlaceholderInit() func(buf *bytes.Buffer) {
	var cnt = 0

	return func(buf *bytes.Buffer) {
		cnt++
		fmt.Fprintf(buf, "$%d", cnt)
	}
}

// Question mark placehodler generator - SQLite
func QmPlaceholderInit() func(buf *bytes.Buffer) {
	return func(buf *bytes.Buffer) {
		buf.WriteString("?")
	}
}

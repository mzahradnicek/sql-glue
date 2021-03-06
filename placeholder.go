package sqlg

import (
	"bytes"
	"fmt"
)

// PostgreSQL placeholder generator
func PqPlaceholder() func(buf *bytes.Buffer) {
	var cnt = 0

	return func(buf *bytes.Buffer) {
		cnt++
		fmt.Fprintf(buf, "$%d", cnt)
	}
}

// Question mark placehodler generator
func QmPlaceholder() func(buf *bytes.Buffer) {
	return func(buf *bytes.Buffer) {
		buf.WriteString("?")
	}
}

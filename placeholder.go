package sqlg

import (
	"bytes"
	"fmt"
)

type PlaceholderFormat interface {
	GeneratePlaceholderFunc() func(buf *bytes.Buffer)
}

type QmPlaceholder struct{}

func (_ QmPlaceholder) GeneratePlaceholderFunc() func(buf *bytes.Buffer) {
	return func(buf *bytes.Buffer) {
		buf.WriteString("?")
	}
}

type PqPlaceholder struct{}

func (_ PqPlaceholder) GeneratePlaceholderFunc() func(buf *bytes.Buffer) {
	var cnt = 0

	return func(buf *bytes.Buffer) {
		cnt++
		fmt.Fprintf(buf, "$%d", cnt)
	}
}

package sqlg

import "bytes"

type Config struct {
	IdentifierEscape func(string) string
	KeyModifier      func(string) string
	PlaceholderInit  func() func(buf *bytes.Buffer)
	Tag              string
}

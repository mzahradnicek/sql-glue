package sqlg

import (
	"testing"
)

func TestPqPlaceholder(t *testing.T) {
	b := NewBuilder(Config{
		IdentifierEscape: func(s string) string { return s },
		KeyModifier:      func(s string) string { return s },
		PlaceholderInit:  PqPlaceholder,
	})

	q, _, _ := b.Glue(&Qg{"SELECT * FROM test WHERE id = %v AND bar = %v", 10, 20})

	if q != "SELECT * FROM test WHERE id = $1 AND bar = $2" {
		t.Error("Error generating pq placeholders")
	}
}

func TestQmPlaceholder(t *testing.T) {
	b := NewBuilder(Config{
		IdentifierEscape: func(s string) string { return s },
		KeyModifier:      func(s string) string { return s },
		PlaceholderInit:  QmPlaceholder,
	})

	q, _, _ := b.Glue(&Qg{"SELECT * FROM test WHERE id = %v AND bar = %v", 10, 20})

	if q != "SELECT * FROM test WHERE id = ? AND bar = ?" {
		t.Error("Error generating qm placeholders")
	}
}

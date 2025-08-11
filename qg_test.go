package sqlg

import (
	"strings"
	"testing"
)

type person struct {
	FirstName string
	LastName  string `sqlg:"last_name" sgsp:"last_name"`
	Password  string `sqlg:"-" sgsp:"-"`
	Age       int

	private bool
}

var pp *person = &person{"John", "Smith", "Secret", 24, true}

var tests = []struct {
	q    *Qg
	want string
}{
	// %and, %or
	{&Qg{"SELECT * FROM test WHERE %and", Qg{"x = 10", "y = %v", 5}}, "SELECT * FROM test WHERE (x = 10 AND y = $1)"},
	{&Qg{"SELECT * FROM test WHERE %or", Qg{"x = 10", "y = %v", 5}}, "SELECT * FROM test WHERE (x = 10 OR y = $1)"},

	// %keys
	{&Qg{"(%keys)", person{FirstName: "John", LastName: "Smith", Password: "NBUSR123", Age: 24}}, `("firstname", "last_name", "age")`},
	{&Qg{"(%keys)", map[string]string{"something": "sausage", "second": "big sausage"}}, `("something", "second")`},

	// %vals
	{&Qg{"(%vals)", map[string]string{"something": "sausage", "second": "big sausage"}}, "($1, $2)"},
	{&Qg{"(%vals)", []string{"sausage", "big sausage"}}, "($1, $2)"},

	// check cache
	{&Qg{"(%keys), (%vals)", pp, Qe{"Age"}, pp}, `("firstname", "last_name"), ($1, $2)`},

	// %set
	// {&Qg{"(%set)", map[string]string{"something": "sausage", "second": "big sausage"}}, `("something" = $1, "second" = $2)`},
	{&Qg{"(%set)", person{FirstName: "John", LastName: "Smith", Password: "NBUSR123", Age: 24}}, `("firstname" = $1, "last_name" = $2, "age" = $3)`},
}

func initBuilder() *Builder {
	return NewBuilder(&Config{
		IdentifierEscape: func(s string) string { return `"` + s + `"` },
		KeyModifier:      strings.ToLower,
		PlaceholderInit:  NrPlaceholderInit,
	})
}

func TestQg(t *testing.T) {
	b := initBuilder()

	for _, tt := range tests {
		res, _, err := b.Glue(tt.q)
		if res != tt.want {
			t.Errorf("got %v, want %v", res, tt.want)
		}

		if err != nil {
			t.Error(err)
		}
	}
}

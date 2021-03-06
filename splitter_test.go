package sqlg

import (
	"strings"
	"testing"
)

// TODO: test for map
var splitterTests = []struct {
	input   interface{}
	exclude []string
	keys    []string
	vals    []interface{}
	err     error
}{
	{person{FirstName: "John", LastName: "Smith", Password: "Secret", Age: 33}, nil, []string{"firstname", "last_name", "age"}, []interface{}{"John", "Smith", 33}, nil},
	{map[string]interface{}{"FirstName": "John", "LastName": "Smith", "Password": "Secret", "Age": 33}, []string{"password"}, []string{"firstname", "lastname", "age"}, []interface{}{"John", "Smith", 33}, nil},
}

func initSplitter() *Splitter {
	return NewSplitter().KeyModifier(strings.ToLower)
}

func TestSplitter(t *testing.T) {
	b := initSplitter()

	for _, tt := range splitterTests {
		keys, vals, err := b.Split(tt.input, tt.exclude)
		t.Logf("Keys: %v Vals: %v Err: %v\n", keys, vals, err)
		if err != tt.err {
			t.Errorf("Error got \"%v\", want \"%v\"", err, tt.err)
		}

		for i, v := range keys {
			if tt.keys[i] != v {
				t.Errorf("Keys got \"%v\", want \"%v\"", v, tt.keys[i])
			}
		}

		for i, v := range vals {
			switch vt := v.(type) {
			case nil:
				if tt.vals[i] != nil {
					t.Errorf("Values got \"%v\", want \"%v\"", vt, tt.vals[i])
				}
			case int:
				if tt.vals[i].(int) != vt {
					t.Errorf("values got \"%v\", want \"%v\"", vt, tt.vals[i])
				}
			case string:
				if tt.vals[i].(string) != vt {
					t.Errorf("values got \"%v\", want \"%v\"", vt, tt.vals[i])
				}
			default:
				t.Errorf("Unknown type for:  \"%v\"", v)
			}
		}
	}
}

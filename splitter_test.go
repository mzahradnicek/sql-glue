package sqlg

import (
	"slices"
	"strings"
	"testing"
)

var testKeyModifier = strings.ToLower

type testPerson1 struct {
	FirstName string
	LastName  string `sqlg:"last_name"`
	Password  string `sqlg:"-"`
	Age       int

	private bool
}

type testPerson2 struct {
	FirstName string `sqlg:",ins"`
	LastName  string `sqlg:"last_name"`
	Password  string `sqlg:"-"`
	Age       int

	private bool
}

type withNested struct {
	Address string `sqlg:"user_address"`

	TestUser testPerson2
}

var splitterStructTests = []struct {
	input   interface{}
	exclude []string
	keys    []string
	vals    []interface{}
	err     error
}{
	{testPerson1{FirstName: "John", LastName: "Smith", Password: "Secret", Age: 33}, nil, []string{"firstname", "last_name", "age"}, []interface{}{"John", "Smith", 33}, nil},
	{&testPerson1{FirstName: "John", LastName: "Smith", Password: "Secret", Age: 33}, nil, []string{"firstname", "last_name", "age"}, []interface{}{"John", "Smith", 33}, nil},
	{testPerson2{FirstName: "John", LastName: "Smith", Password: "Secret", Age: 33}, nil, []string{"firstname", "last_name", "age"}, []interface{}{"John", "Smith", 33}, nil},
	{testPerson2{FirstName: "John", LastName: "Smith", Password: "Secret", Age: 33}, []string{"ins"}, []string{"last_name", "age"}, []interface{}{"Smith", 33}, nil},
	{testPerson2{FirstName: "John", LastName: "Smith", Password: "Secret", Age: 33}, []string{"Age"}, []string{"firstname", "last_name"}, []interface{}{"John", "Smith"}, nil},
	{withNested{TestUser: testPerson2{FirstName: "John", LastName: "Smith", Password: "Secret", Age: 33}, Address: "Fort Knocks 13"}, []string{"Age"}, []string{"user_address", "testuser.firstname", "testuser.last_name"}, []interface{}{"Fort Knocks 13", "John", "Smith"}, nil},
}

var splitterMapTests = []struct {
	input   map[string]interface{}
	exclude []string
	keys    []string
	err     error
}{
	{map[string]interface{}{"FirstName": "John", "LastName": "Smith", "Password": "Secret", "Age": 33}, []string{"Password"}, []string{"firstname", "lastname", "age"}, nil},
}

func initSplitter() *Splitter {
	return NewSplitter().KeyModifier(testKeyModifier)
}

func TestSplitterStruct(t *testing.T) {
	b := initSplitter()

	for ti, tt := range splitterStructTests {
		keys, vals, err := b.Split(tt.input, tt.exclude...)
		t.Logf("#%v Keys: %v Vals: %v Err: %v", ti, keys, vals, err)
		if err != tt.err {
			t.Errorf("#%v Error got \"%v\", want \"%v\"", ti, err, tt.err)
		}

		for i, v := range keys {
			if tt.keys[i] != v {
				t.Errorf("#%v keys got \"%v\", want \"%v\"", ti, v, tt.keys[i])
			}
		}

		for i, v := range vals {
			switch vt := v.(type) {
			case nil:
				if tt.vals[i] != nil {
					t.Errorf("#%v values got \"%v\", want \"%v\"", ti, vt, tt.vals[i])
				}
			case int:
				if tt.vals[i].(int) != vt {
					t.Errorf("#%v values got \"%v\", want \"%v\"", ti, vt, tt.vals[i])
				}
			case string:
				if tt.vals[i].(string) != vt {
					t.Errorf("#%v values got \"%v\", want \"%v\"", ti, vt, tt.vals[i])
				}
			default:
				t.Errorf("#%v Unknown type for:  \"%v\"", ti, v)
			}
		}
	}
}

func TestSplitterMap(t *testing.T) {
	b := initSplitter()

	for ti, tt := range splitterMapTests {
		keys, vals, err := b.Split(tt.input, tt.exclude...)
		t.Logf("#%v Keys: %v Vals: %v Err: %v", ti, keys, vals, err)
		if err != tt.err {
			t.Errorf("#%v Error got \"%v\", want \"%v\"", ti, err, tt.err)
			return
		}

		if len(keys) != len(vals) {
			t.Errorf("#%v keys(%v) and values(%v) has different length", ti, len(keys), len(vals))
		}

		for i, key := range keys {
			if slices.Contains(tt.exclude, key) {
				t.Errorf("#%v key \"%v\", is excluded", ti, key)
				return
			}

			// find real key in map
			var mkey string
			for mk, _ := range tt.input {
				if testKeyModifier(mk) == key {
					mkey = mk
					break
				}
			}

			if mkey == "" {
				t.Errorf("#%v key \"%v\", not found in input", key, key)
				return
			}

			// check vals
			switch vt := vals[i].(type) {
			case nil:
				if vt != nil {
					t.Errorf("#%v valuesn got \"%v\", want \"%v\"", key, vt, tt.input[mkey])
				}
			case int:
				if vt != tt.input[mkey].(int) {
					t.Errorf("#%v valuesi got \"%v\", want \"%v\"", key, vt, tt.input[mkey].(int))
				}
			case string:
				if vt != tt.input[mkey].(string) {
					t.Errorf("#%v valuess got \"%v\", want \"%v\"", key, vt, tt.input[mkey])
				}
			default:
				t.Errorf("#%v Unknown type for: \"%v\"", key, key)
			}
		}
	}
}

func TestNewSplitter(t *testing.T) {
	s := NewSplitter()
	if s == nil {
		t.Error("Expected NewSplitter to return a non-nil value")
	}
	if s.tag != "sqlg" {
		t.Errorf("Expected default tag 'sqlg', got '%s'", s.tag)
	}
}

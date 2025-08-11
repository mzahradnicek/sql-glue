package sqlg

import (
	"reflect"
	"testing"
)

var tRETStruct = struct {
	Name string
}{
	Name: "Joe",
}

var tRETString string = "world"

var testsResolveElemType = []struct {
	name    string
	input   interface{}
	kind    []reflect.Kind
	wantVal interface{}
	wantPtr interface{}
}{
	{
		name:    "Struct pointer, expects struct",
		input:   &tRETStruct,
		kind:    []reflect.Kind{reflect.Struct},
		wantVal: tRETStruct,
		wantPtr: &tRETStruct,
	}, {
		name:    "String value, expects string",
		input:   "hello",
		kind:    []reflect.Kind{reflect.String},
		wantVal: "hello",
		wantPtr: nil,
	}, {
		name:    "String pointer, expects string",
		input:   &tRETString,
		kind:    []reflect.Kind{reflect.String},
		wantVal: "world",
		wantPtr: &tRETString,
	}, {
		name:    "Mismatched kind (int input, string expected)",
		input:   789,
		kind:    []reflect.Kind{reflect.String},
		wantVal: nil,
		wantPtr: nil,
	}, {
		name:    "Nil input",
		input:   nil,
		kind:    []reflect.Kind{reflect.Int},
		wantVal: nil,
		wantPtr: nil,
	},
}

func TestResolveElemType(t *testing.T) {
	for _, tc := range testsResolveElemType {
		t.Run(tc.name, func(t *testing.T) {
			val, ptr := resolveElemType(tc.input, tc.kind...)

			// Porovnanie výsledkov
			if val != tc.wantVal {
				t.Errorf("For val: got %v, want %v", val, tc.wantVal)
			}
			if ptr != tc.wantPtr {
				t.Errorf("For ptr: got %v, want %v", ptr, tc.wantPtr)
			}
		})
	}
}

package sqlg

import "testing"

var testsToSnake = []struct {
	in   string
	want string
}{
	{"ToSnake", "to_snake"},
	{"TSnake", "t_snake"},
	{"ToSnakeS", "to_snake_s"},
	{"toSnakeS", "to_snake_s"},
}

func TestToSnake(t *testing.T) {
	for _, tt := range testsToSnake {
		res := ToSnake(tt.in)
		if res != tt.want {
			t.Errorf("got %v, want %v", res, tt.want)
		}
	}
}

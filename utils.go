package sqlg

import (
	"reflect"
	"slices"
)

func resolveElemType(input interface{}, kind ...reflect.Kind) (val interface{}, ptr interface{}) {
	value := reflect.ValueOf(input)

	if value.Kind() == reflect.Pointer {
		ptr = value.Interface()
		value = value.Elem()
	}

	if slices.Contains(kind, value.Kind()) {
		val = value.Interface()
	}

	return
}

func errorChunkPreview(chunk string, pos int) string {
	if pos > len(chunk) {
		pos = len(chunk)
	}

	start := pos - 20
	if start < 0 {
		start = 0
	}

	return chunk[start:pos]
}

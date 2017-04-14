package slice

import (
	"reflect"
)

// Reverse can reverse an Array, Slice or String.
func Reverse(slice interface{}) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Ptr {
		return
	}

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		runes := []rune(v.String())
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		v.SetString(string(runes))
	case reflect.Array, reflect.Slice:
		for i, j := 0, v.Len()-1; i < j; i, j = i+1, j-1 {
			vi, vj := v.Index(i), v.Index(j)
			t := vi.Interface()
			vi.Set(vj)
			vj.Set(reflect.ValueOf(t))
		}
	}
}

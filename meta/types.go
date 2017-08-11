package meta

import (
	"reflect"
)

type Logic uint // zero value represents X undefined

const (
	X Logic = iota
	T
	F
	Z
)

func Add(a, b interface{}) reflect.Value {
	if Width(a) < Width(b) {
		return Add(b, a)
	}

	va := reflect.ValueOf(a)
	switch va.Type().Kind() {
	case reflect.Int64, reflect.Int:
		vb := reflect.ValueOf(b)
		return reflect.ValueOf(va.Int() + vb.Int())
	default:
		panic(va)
	}
	return reflect.ValueOf(nil)
}

func True(v interface{}) bool {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Bool:
		return v.(bool)
	case reflect.Int:
		return v.(int) != 0
	case reflect.Uint:
		return v.(uint) != 0
	}

	return false
}

func Width(v interface{}) int {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Bool:
		return 1
	}
	return t.Bits()
}

func Cat(values ...interface{}) []Logic {
	size := 0
	for _, v := range values {
		size += Width(v)
	}
	bits := make([]Logic, size)
	off := 0
	for _, v := range values {
		off += fillLogic(v, bits[off:])
	}
	return bits
}

func fillLogic(v interface{}, logic []Logic) int {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Bool:
		if v.(bool) {
			logic[0] = T
		} else {
			logic[0] = F
		}
		return 1
	case reflect.Int:
	case reflect.Uint:
	}

	return t.Bits()
}

func ToLogic(v interface{}) []Logic {
	w := Width(v)
	bits := make([]Logic, w)
	fillLogic(v, bits)
	return bits
}

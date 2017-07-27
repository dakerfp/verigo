package meta

import (
	"fmt"
	"reflect"
)

type Module interface {
	Meta() *Mod
	Sub(...Module)
}

type Mod struct {
	Name    string
	subs    []Module
	wires   [][]interface{}
	inputs  map[string]reflect.Value
	outputs map[string]reflect.Value
}

func Init(m Module) {
	meta := m.Meta()
	// build it bottom up
	for _, sub := range meta.subs {
		Init(sub)
	}
	data := reflect.Indirect(reflect.ValueOf(m))
	t := data.Type()
	meta.Name = t.Name()
	meta.inputs = make(map[string]reflect.Value)
	meta.outputs = make(map[string]reflect.Value)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// ignore Module embedding
		if field.Type == reflect.TypeOf(meta).Elem() {
			continue
		}

		switch string(field.Tag) {
		case "":
			// ignore field
		case "input":
			meta.inputs[field.Name] = data.FieldByIndex(field.Index)
		case "output":
			meta.outputs[field.Name] = data.FieldByIndex(field.Index)
		case "submodule":
			// XXX: init automatically
		default:
			panic(fmt.Errorf("io tag %q not supported in field %q", field.Tag.Get("io"), field.Name))
		}
	}
	fmt.Println(data, t, meta)
}

func (m *Mod) Meta() *Mod {
	return m
}

func (m *Mod) Assign(recv interface{}, f interface{}) {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic("assign does not gets function")
	}

	if t.NumIn() > 0 {
		panic("edge triggedered not yet supported") // XXX
	}

	if t.NumOut() != 1 {
		panic("assign must return a single function")
	}

	recvt := reftype(recv)
	rett := t.Out(0)
	if !rett.ConvertibleTo(recvt) {
		panic("assign to different types")
	}

	fv := reflect.ValueOf(f)
	fmt.Println(t.NumIn())
	fmt.Println(t.NumOut())
	fmt.Println(fv)
	fmt.Println(reflect.TypeOf(m))
}

func reftype(v interface{}) reflect.Type {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		panic("wire element must be a pointer to another based type")
	}

	e := reflect.Indirect(ptr)
	if e.Kind() == reflect.Ptr {
		panic("wire element must be a pointer to another based type")
	}

	return e.Type()
}

func (m *Mod) Wire(wire ...interface{}) {
	if len(wire) <= 1 {
		panic("wire must have at least 2 elements")
	}

	wt0 := reftype(wire[0])
	for _, we := range wire[1:] {
		wt := reftype(we)
		if !wt.ConvertibleTo(wt0) {
			panic(wt)
		}
	}

	m.wires = append(m.wires, wire)
}

func (m *Mod) Sub(subs ...Module) {
	m.subs = subs
}

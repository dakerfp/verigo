package meta

import (
	"fmt"
	"reflect"
)

type Value interface{}
type In interface{}
type Out interface{}

type Module interface {
	Meta() *Mod
	Sub(Module)
}

type Mod struct {
	subs    []Module
	wires   [][]interface{}
	inputs  map[string]reflect.Value
	outputs map[string]reflect.Value
}

func Init(m Module) {
	data := reflect.Indirect(reflect.ValueOf(m))
	t := data.Type()

	meta := m.Meta()
	meta.inputs = make(map[string]reflect.Value)
	meta.outputs = make(map[string]reflect.Value)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		switch field.Tag.Get("io") {
		case "input":
			meta.inputs[field.Name] = data.FieldByIndex(field.Index)
		case "output":
			meta.outputs[field.Name] = data.FieldByIndex(field.Index)
		}

		fmt.Println(field.Name)
	}
	fmt.Println(data, t, meta)

	for _, sub := range meta.subs {
		Init(sub)
	}
}

func (m *Mod) Meta() *Mod {
	return m
}

func (m *Mod) Input(v interface{}) Value {
	return nil
}

func (m *Mod) Output(v interface{}) Value {
	return nil
}

func (m *Mod) Assign(v interface{}, expr interface{}) {
	// panic(expr)
}

func wiretype(v interface{}) reflect.Type {
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

	wt0 := wiretype(wire[0])
	for _, we := range wire[1:] {
		wt := wiretype(we)
		if !wt.ConvertibleTo(wt0) {
			panic(wt)
		}
	}

	m.wires = append(m.wires, wire)
}

func (m *Mod) Sub(sub Module) {
	m.subs = append(m.subs, sub)
}

type Input interface{}
type Output interface{}

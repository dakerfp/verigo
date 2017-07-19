package verilog

import (
	"errors"
	"reflect"
	"time"
)

var (
	ErrMismatchingValueTypes = errors.New("mismatching types")
	ErrTypeNotSupported      = errors.New("type not supported")
)

type Module struct {
	Name            string
	Inputs, Outputs map[string]int64
	Signals         map[string]int64
	Delay           time.Duration
}

func sizeOfVType(t reflect.Type) int64 {
	f := reflect.New(t)
	return f.MethodByName("Size").Call(nil)[0].Int()
}

func GetModule(v interface{}) (m Module, err error) {
	m.Inputs = make(map[string]int64)
	m.Outputs = make(map[string]int64)

	tp := reflect.TypeOf(v)
	vType := reflect.TypeOf((*Value)(nil)).Elem()

	m.Name = tp.Name()

	// getting inputs
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		if !reflect.PtrTo(field.Type).Implements(vType) {
			err = ErrTypeNotSupported
			return
		}
		m.Inputs[field.Name] = sizeOfVType(field.Type)
	}

	// getting outputs
	for i := 0; i < tp.NumMethod(); i++ {
		method := tp.Method(i)
		fnt := method.Func.Type()
		if fnt.NumOut() != 1 {
			panic(fnt)
		}
		outTypePtr := reflect.PtrTo(fnt.Out(0))
		if !outTypePtr.Implements(vType) {
			panic(outTypePtr)
		}

		m.Outputs[method.Name] = sizeOfVType(fnt.Out(0))
	}

	return
}

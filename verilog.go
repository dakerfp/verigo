package verilog

import (
	"errors"
	"time"
	"reflect"
)

var (
	ErrMismatchingValueTypes = errors.New("mismatching types")
	ErrTypeNotSupported = errors.New("type not supported")
)

type Logic byte

type PosEdge Logic
type NegEdge Logic

const (
	Zero Logic = iota
	One
	X
	Z
)

func LogicAnd(a, b Logic) Logic {
	switch {
        case a == 0 || b == 0:
                return 0
        case a == 1 && b == 1:
                return 1
        default:
                 return X
        }

}

type Vector []Value

type Value interface {
	Size() int64
}

func (l *Logic) Set(v Value) (err error) {
	switch v.(type) {
	case *Logic:
		*l = *v.(*Logic)
	default:
		err = ErrMismatchingValueTypes
	}
	return
}

func (l *Logic) Size() int64 {
	return 0
}

func (v *Vector) Size() (size int64) {
	for i := 0; i < len(*v); i++ {
		size += (*v)[i].Size()
	}
	return
}

type Module struct {
	Name string
	Inputs, Outputs map[string]int64
	Signals map[string]int64
	Delay time.Duration
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
		if fnt.NumOut() != 1  {
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




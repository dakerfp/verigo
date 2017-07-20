package verilog

type Logic byte
type PosEdge Logic
type NegEdge Logic

const (
	Zero Logic = iota
	One
	X
	Z
)

var LogicValues = []Logic{0, 1, X, Z}

type Vector []Value

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

type Value interface {
	Size() int64
	Get() []Logic
	Init()
}

func (l *Logic) Init() {
	*l = X
}

func (l *Logic) Get() []Logic {
	return []Logic{*l}
}

func (l *Logic) Set(v Value) {
	switch v.(type) {
	case *Logic:
		*l = *v.(*Logic)
	default:
		panic(ErrMismatchingValueTypes)
	}
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

func (v *Vector) Init() {
	for i := 0; i < len(*v); i++ {
		(*v)[i].Init()
	}
}

func (vec *Vector) Get() (bits []Logic) {
	bits = make([]Logic, 0, vec.Size())
	for i := 0; i < len(*vec); i++ {
		bits = append(bits, (*vec).Get()...)
	}
	return
}

func (vev *Vector) Set(v Value) {
	panic(ErrMismatchingValueTypes)
}

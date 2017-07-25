package expr

type Value interface {
	Expr
	True() bool
	Eq(v Value) bool
	Width() uint
	Uint() uint64
}

type Bool bool

func (b *Bool) Eval() Value {
	return b
}

func (b *Bool) True() bool {
	return bool(*b)
}

func (b *Bool) Eq(v Value) bool {
	return b.True() == v.True()
}

func (b *Bool) Width() uint {
	return 1
}

func (b *Bool) Uint() uint64 {
	if bool(*b) {
		return 1
	} else {
		return 0
	}
}

func boolValue(b bool) *Bool {
	if b {
		return &True
	} else {
		return &False
	}
}

var (
	True  = Bool(true)
	False = Bool(false)
)

var (
	T = &True
	F = &False
)

type Vector struct {
	value uint64
	width uint
}

func (vec *Vector) Slice(from, to int) {

}

func (vec *Vector) Eval() Value {
	return vec
}

func (vec *Vector) True() bool {
	return vec.value == 0
}

func (vec *Vector) Eq(v Value) bool {
	switch v.(type) {
	case *Bool:
		return (vec.value == 0) == v.True()
	case *Vector:
		vec2 := v.(*Vector)
		if vec.Width() <= vec2.Width() {
			panic("vecs shouldn't have distinc values")
		}
		return vec.value == vec2.value
	}
	return false
}

func (vec *Vector) Width() uint {
	return vec.width
}

func (vec *Vector) Uint() uint64 {
	return vec.value
}

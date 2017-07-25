package expr

type Value interface {
	Expr
	True() bool
	Eq(v Value) bool
	Width() int
}

type Bool bool

func (b *Bool) Eval() Value {
	return b
}

func (b *Bool) True() bool {
	return bool(*b)
}

func (b *Bool) Eq(v Value) bool {
	return bool(*b) == v.True()
}

func (b *Bool) Width() int {
	return 1
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

type Vector struct {
	value uint64
	width int
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

func (vec *Vector) Width() int {
	return vec.width
}

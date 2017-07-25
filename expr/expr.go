package expr

type Expr interface {
	Eval() Value
}

func Eq(vl, vr Value) bool {
	if vl == nil || vr == nil {
		return false
	}
	if vl.Width() >= vr.Width() {
		return vl.Eq(vr)
	}
	return vr.Eq(vl)
}

type UnaryExpr struct {
	Expr Expr
	Op   func(v Value) Value
}

func (uo *UnaryExpr) Eval() Value {
	return uo.Op(uo.Expr.Eval())
}

type BinaryExpr struct {
	Expr1, Expr2 Expr
	Op           func(v1, v2 Value) Value
}

func (bo *BinaryExpr) Eval() Value {
	return bo.Op(bo.Expr1.Eval(), bo.Expr2.Eval())
}

func Not(expr Expr) *UnaryExpr {
	return &UnaryExpr{
		expr,
		func(v Value) Value { // XXX: must support Vector not
			return boolValue(!v.True())
		},
	}
}

func And(expr1, expr2 Expr) *BinaryExpr {
	return &BinaryExpr{
		expr1, expr2,
		func(v1 Value, v2 Value) Value { // XXX: must support Vector not
			return boolValue(v1.True() && v2.True())
		},
	}
}

func Or(expr1, expr2 Expr) *BinaryExpr {
	return &BinaryExpr{
		expr1, expr2,
		func(v1 Value, v2 Value) Value { // XXX: must support Vector not
			return boolValue(v1.True() || v2.True())
		},
	}
}

func max(a, b uint) uint {
	if a > b {
		return a
	} else {
		return b
	}
}

func Add(expr1, expr2 Expr) *BinaryExpr {
	return &BinaryExpr{
		expr1, expr2,
		func(v1 Value, v2 Value) Value { // XXX: must support Vector not
			a := v1.Uint()
			b := v2.Uint()
			return &Vector{a + b, max(v1.Width(), v2.Width())}
		},
	}
}

type IfExpr struct {
	Cond, If, Else Expr
}

func (ife *IfExpr) Eval() Value {
	if ife.Cond.Eval().True() {
		return ife.If.Eval()
	} else {
		return ife.Else.Eval()
	}
}

type WalkFunc func(Expr, []Expr) error

func Walk(root Expr, walkFunc WalkFunc) (err error) {
	visited := make(map[Expr]bool)
	list := []Expr{root}

	for len(list) > 0 {
		expr := list[0]
		list = list[1:]
		if visited[expr] {
			continue
		}
		visited[expr] = true

		var next []Expr
		switch expr.(type) {
		case Value:
		case *UnaryExpr:
			ue := expr.(*UnaryExpr)
			next = []Expr{ue.Expr}
		case *BinaryExpr:
			be := expr.(*BinaryExpr)
			next = []Expr{be.Expr1, be.Expr2}
		case *IfExpr:
			ife := expr.(*IfExpr)
			next = []Expr{ife.Cond, ife.If, ife.Else}
		}
		err = walkFunc(expr, next)
		if err != nil {
			return
		}
		list = append(list, next...)
	}
	return
}

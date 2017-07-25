package expr

type Expr interface {
	Eval() Value
}

func Eq(vl, vr Value) bool {
	if vl.Width() >= vr.Width() {
		return vl.Eq(vr)
	} else {
		return vr.Eq(vl)
	}
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

func Not(expr Expr) Expr {
	return &UnaryExpr{
		expr,
		func(v Value) Value { // XXX: must support Vector not
			return boolValue(!v.True())
		},
	}
}

func And(expr1, expr2 Expr) Expr {
	return &BinaryExpr{
		expr1, expr2,
		func(v1 Value, v2 Value) Value { // XXX: must support Vector not
			return boolValue(v1.True() && v2.True())
		},
	}
}

func Or(expr1, expr2 Expr) Expr {
	return &BinaryExpr{
		expr1, expr2,
		func(v1 Value, v2 Value) Value { // XXX: must support Vector not
			return boolValue(v1.True() || v2.True())
		},
	}
}

type Var struct {
	V Value
}

func (vr *Var) Update(v Value) bool {
	if vr.V.Eq(v) {
		return false
	}
	vr.V = v
	return true
}

package expr

import "testing"

func TestUnary(t *testing.T) {
	a := Bool(true)
	b := Bool(false)

	if b.True() {
		t.Fail()
	}

	if !a.True() {
		t.Fail()
	}

	notA := Not(&a)
	if notA.Eval().True() {
		t.Fail()
	}

	notB := Not(&b)
	if !notB.Eval().True() {
		t.Fail()
	}
}

func TestEq(t *testing.T) {
	if !Eq(T, T) {
		t.Fatal()
	}

	if !Eq(F, F) {
		t.Fatal()
	}

	if Eq(T, F) {
		t.Fatal()
	}

	if Eq(F, T) {
		t.Fatal()
	}
}

func TestBinary(t *testing.T) {
	a := Bool(true)
	b := Bool(false)

	if !a.True() {
		t.Fail()
	}

	if b.True() {
		t.Fail()
	}

	aAndB := And(&a, &b)
	if aAndB.Eval().True() {
		t.Fail()
	}

	aAndA := And(&a, &a)
	if !aAndA.Eval().True() {
		t.Fail()
	}

	bAndB := And(&b, &b)
	if bAndB.Eval().True() {
		t.Fail()
	}
}

func TestIf(t *testing.T) {
	for _, cond := range []bool{true, false} {
		for _, ifv := range []Value{&True, &False} {
			for _, elsev := range []Value{&True, &False} {
				ifExpr := IfExpr{boolValue(cond), ifv, elsev}

				var v Value
				if cond {
					v = ifv
				} else {
					v = elsev
				}
				if !Eq(ifExpr.Eval(), v) {
					t.Fail()
				}
			}
		}
	}
}

func TestWalk(t *testing.T) {
	a := Bool(true)
	b := Bool(false)
	c := Bool(true)
	d := Bool(false)

	expr := Not(Or(And(&a, &b), And(&c, &d)))

	count := 0
	err := Walk(expr, func(e Expr, deps []Expr) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if count != 8 {
		t.Fatal(count)
	}

	// cyclic case
	exprOr := Or(And(&a, &b), nil)
	exprOr.Expr2 = exprOr

	count = 0
	err = Walk(exprOr, func(e Expr, deps []Expr) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatal(count)
	}
}

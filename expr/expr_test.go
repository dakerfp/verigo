package expr

import "testing"

func TestUnary(t *testing.T) {
	a := Bool(true)
	b := Bool(false)

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

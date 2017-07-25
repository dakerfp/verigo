package expr

import "testing"

func TestUnary(t *testing.T) {
	a := Bool(true)

	if !a.True() {
		t.Fail()
	}

	notA := Not(&a)
	if notA.Eval().True() {
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

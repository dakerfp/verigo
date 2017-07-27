package meta

import (
	"testing"
)

func TestTrue(t *testing.T) {
	if !True(true) {
		t.Fatal(true)
	}

	if True(false) {
		t.Fatal(false)
	}

	if !True(1) {
		t.Fatal(1)
	}

	if !True(-1) {
		t.Fatal(-1)
	}

	if True(0) {
		t.Fatal(0)
	}
}

func TestWidth(t *testing.T) {
	if w := Width(true); w != 1 {
		t.Fatal(w)
	}

	if w := Width(int(42)); w != 64 {
		t.Fatal(w)
	}
}

func TestCat(t *testing.T) {
	if l := Cat(true, false, true, true); len(l) != 4 {
		t.Fatal(l)
	}

	if l := Cat(); len(l) != 0 {
		t.Fatal(l)
	}

	if l := Cat(42, true, true); len(l) != 64+2 {
		t.Fatal(l)
	}
}

// module mux2
// 	(input logic A, B, Sel
// 	 output logic Out)

// 	assign Out = (Sel) ? A : B;

// endmodule : mux2

type Mux2 struct {
	Mod

	A, B, Sel bool `io:"input"`
	Out       bool `io:"output"`
}

func mux2() *Mux2 {
	m := &Mux2{}

	m.Assign(&m.Out, func() bool {
		if True(m.Sel) {
			return m.B
		}
		return m.A
	})
	// or  m.Assign(&m.Out, m.output)

	return m
}

func (m *Mux2) output() bool {
	if True(m.Sel) {
		return m.B
	}
	return m.A
}

func TestMux2(t *testing.T) {
	mux := mux2()
	mux.A = false
	mux.B = true
	mux.Sel = true

	Init(mux)
	meta := mux.Meta()
	if len(meta.inputs) != 3 {
		t.Fail()
	}

	if len(meta.outputs) != 1 {
		t.Fail()
	}
}

// module mux4
// 	(input logic A, B, C, D, Sel0, Sel1
// 	 output logic Out)

// 	wire O1, O2;
// 	mux2 m0(A, B, Sel0, O1);
// 	mux2 m1(C, D, Sel0, O2);
// 	mux2 mo(O1, O2, Sel1, Out);

// endmodule : mux4

type Mux4 struct {
	Mod

	A, B, C, D, Sel0, Sel1 bool `io:"input"`
	Out                    bool `io:"output"`
}

func mux4() *Mux4 {
	m := &Mux4{}

	ml := mux2()
	mr := mux2()
	mo := mux2()

	m.Sub(ml, mr, mo)

	m.Wire(&ml.A, &m.A)
	m.Wire(&ml.B, &m.B)
	m.Wire(&ml.Sel, &m.Sel0)

	return m
}

func TestMux4(t *testing.T) {
	mux := mux4()

	Init(mux)
	meta := mux.Meta()
	if len(meta.inputs) != 6 {
		t.Fatal(len(meta.inputs))
	}

	if len(meta.outputs) != 1 {
		t.Fatal()
	}
}

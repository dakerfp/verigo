package meta

import (
	"testing"
)

// module mux2
// 	(input logic A, B, Sel
// 	 output logic Out)

// 	assign Out = (Sel) ? A : B;

// endmodule : mux2

type Mux2 struct {
	Mod

	A, B, Sel bool "input"
	Out       bool "output"
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
	if len(meta.Inputs) != 3 {
		t.Fail()
	}

	if len(meta.Outputs) != 1 {
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

	A, B, C, D, Sel0, Sel1 bool "input"
	Out                    bool "output"

	ml, mr, mo *Mux2 "submodule"
}

func mux4() *Mux4 {
	m := &Mux4{}

	m.ml = mux2()
	m.mr = mux2()
	m.mo = mux2()
	m.Sub(m.ml, m.mr, m.mo) // XXX: use struct tag

	// m.Wire(&m.ml.A, &m.A)
	// m.Wire(&m.ml.B, &m.B)
	// m.Wire(&m.ml.Sel, &m.Sel0)

	return m
}

func TestMux4(t *testing.T) {
	mux := mux4()

	Init(mux)
	meta := mux.Meta()
	if len(meta.Inputs) != 6 {
		t.Fatal(len(meta.Inputs))
	}

	if len(meta.Outputs) != 1 {
		t.Fatal()
	}
}

// module And
// 	(input bit A, B,
// 	 output O);

// 	assign O = A && B;

// endmodule : And

type And struct {
	Mod

	A, B bool "input"
	O    bool "output"
}

func and() *And {
	m := &And{}
	Init(m)
	m.Always(`O`, `A && B`)
	return m
}

func TestAnd(t *testing.T) {
	a := and()
	sig := a.Values["O"] // XXX
	v := sig.Update()
	if v.Bool() {
		t.Fatal(v)
	}

	a.A = true
	a.B = true
	v = sig.Update()
	if !v.Bool() {
		t.Fatal(v)
	}
}

type DFF struct {
	Mod

	Clk, In bool "input"
	Out     bool "output"
}

func dff() *DFF {
	m := &DFF{}
	Init(m)
	m.Always(`Out`, `In`, Pos(`Clk`))
	return m
}

func TestDFF(t *testing.T) { // XXX: create proper test
	m := dff()
	if m.Out {
		t.Fatal(m.Out)
	}
}

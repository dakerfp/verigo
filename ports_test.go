package verilog

import (
	"testing"
)

// +module
type And struct {
	A Logic // input
	B Logic // input
}

func (and And) Out() Logic {
	return LogicAnd(and.A, and.B)
}

// +module
type Mux2 struct {
	A      Logic
	B      Logic
	Select Logic
}

func (mux Mux2) Out() Logic {
	switch mux.Select {
	case 0:
		return mux.A
	case 1:
		return mux.B
	default:
		return X
	}
}

// +module
type Negate struct {
	Clk   Logic
	Input Logic
}

func (n Negate) Output(Clk PosEdge) Logic {
	switch n.Input {
	case 0:
		return 1
	case 1:
		return 0
	default:
		return X
	}
}

func TestPorts(t *testing.T) {
	var and And
	mand, err := GetModule(and)
	if err != nil {
		t.Fatal(err)
	}

	var n Negate
	mn, err := GetModule(n)
	if err != nil {
		t.Fatal(err)
	}

	var mux Mux2
	muxn, err := GetModule(mux)
	if err != nil {
		t.Fatal(err)
	}

	tts := map[string]struct {
		m                 Module
		nInputs, nOutputs int
	}{
		"And":    {mand, 2, 1},
		"Negate": {mn, 2, 1},
		"Mux2":   {muxn, 3, 1},
	}

	for name, tt := range tts {
		if tt.m.Name != name {
			t.Fatal(name)
		}
		if len(tt.m.Inputs) != tt.nInputs {
			t.Fatal(name)
		}
		if len(tt.m.Outputs) != tt.nOutputs {
			t.Fatal(name)
		}
	}
}

func TestTruthTable(t *testing.T) {
	var and And
	for _, a := range LogicValues {
		for _, b := range LogicValues {
			and.A = a
			and.B = b
			if and.Out() != LogicAnd(a, b) {
				t.Fatal(a, b)
			}
		}
	}
}

func TestInit(t *testing.T) {
	var neg Negate
	neg.Clk.Init()
	neg.Input.Init()

	if neg.Clk != X {
		t.Fatal(neg.Clk)
	}
	if neg.Input != X {
		t.Fatal(neg.Input)
	}
	if neg.Output(0) != X {
		t.Fatal(neg.Input)
	}
}

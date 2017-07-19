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
	if n.Input == 1 {
		return 0
	}
	return 1
}

func TestPorts(t *testing.T) {
	var n Negate
	m, err := GetModule(n)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "Negate" {
		t.Fail()
	}
	if len(m.Inputs) != 2 {
		t.Fail()
	}
	if len(m.Outputs) != 1 {
		t.Fail()
	}
}

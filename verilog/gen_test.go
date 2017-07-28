package verilog

import (
	"os"
	"testing"

	"github.com/dakerfp/verigo/meta"
)

type Mux2 struct {
	meta.Mod

	A, B, Sel bool "input"
	Out       bool "output"
}

func mux2() *Mux2 {
	m := &Mux2{}

	meta.Init(m)

	m.Assign(&m.Out, func() bool {
		if meta.True(m.Sel) {
			return m.B
		}
		return m.A
	})
	// or  m.Assign(&m.Out, m.output)

	return m
}

func TestGenModule(t *testing.T) {
	m := mux2()
	err := GenerateVerilog(os.Stdout, m)
	if err != nil {
		panic(err)
	}
}

package verilog

import (
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
	m.Always(`Out`, `Sel && B || ^Sel && A`)

	return m
}

func TestGenModule(t *testing.T) {
	// m := mux2()
	// err := GenerateVerilog(os.Stdout, m)
	// if err != nil {
	// 	panic(err)
	// }
}

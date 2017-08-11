package sim

import (
	"reflect"
	"testing"
	"time"

	"github.com/dakerfp/verigo/meta"
)

func Node(v0 reflect.Value, update meta.UpdateFunc) *meta.Node {
	return &meta.Node{
		V:      v0,
		Listen: nil,
		Notify: nil,
		Update: update,
	}
}

var (
	True  = true
	False = false
)

func T() reflect.Value {
	return reflect.ValueOf(True)
}

func F() reflect.Value {
	return reflect.ValueOf(False)
}

func Not(a *meta.Node) meta.UpdateFunc {
	return func() reflect.Value {
		return reflect.ValueOf(!a.V.Bool())
	}
}

func And(a *meta.Node, b *meta.Node) meta.UpdateFunc {
	return func() reflect.Value {
		return reflect.ValueOf(a.V.Bool() && b.V.Bool())
	}
}

func TestComb(t *testing.T) {
	// build nodes
	a := Node(F(), F)
	b := Node(F(), F)
	c := Node(F(), F)
	d := Node(F(), F)

	ab := Node(F(), And(a, b))
	meta.Connect(a, ab, meta.Anyedge)
	meta.Connect(b, ab, meta.Anyedge)

	cd := Node(F(), And(c, d))
	meta.Connect(c, cd, meta.Anyedge)
	meta.Connect(d, cd, meta.Anyedge)

	o := Node(F(), And(ab, cd))
	meta.Connect(ab, o, meta.Anyedge)
	meta.Connect(cd, o, meta.Anyedge)

	if o.Update().Bool() {
		t.Fatal(o.Update())
	}

	sim := NewSimulator()
	go func() {
		now := time.Now()
		sim.Set(a, True, now)
		sim.Set(b, True, now)
		sim.Set(c, True, now)
		sim.Set(d, True, now)
		sim.End()
	}()
	sim.Run()

	if !ab.V.Bool() {
		t.Fatal(ab.V)
	}

	if !cd.V.Bool() {
		t.Fatal(cd.V)
	}

	if !o.V.Bool() {
		t.Fatal(o.V)
	}
}

func TestClk(t *testing.T) {
	// build nodes
	// always_ff @(posedge clk)
	//     na <= ~a;
	clk := Node(F(), F)
	a := Node(F(), F)
	na := Node(T(), Not(a))
	meta.Connect(clk, na, meta.Posedge|meta.Block) // only on clock trigger

	if !na.Update().Bool() {
		t.Fatal(na.Update())
	}

	now := time.Now()

	sim := NewSimulator()
	go func() {
		sim.Set(a, True, now)
		sim.End()
	}()
	sim.Run()

	if !na.V.Bool() {
		t.Fatal(na.Update())
	}

	sim = NewSimulator()
	go func() {
		sim.Set(clk, True, now)
		sim.End()
	}()
	sim.Run()

	if na.V.Bool() {
		t.Fatal(na.V)
	}

	sim = NewSimulator()
	go func() {
		sim.Set(a, True, now)
		sim.Set(clk, False, now)
		sim.Set(clk, True, now.Add(1)) // trigger a <- true
		// should not trigger a <- false
		sim.Set(a, False, now.Add(2))
		sim.Set(clk, False, now.Add(3))
		sim.End()
	}()
	sim.Run()

	if na.V.Bool() {
		t.Fatal(na.V)
	}
}

type DFF struct {
	meta.Mod

	Clk, In bool "input"
	Out     bool "output"
}

func dff() *DFF {
	m := &DFF{}
	meta.Init(m)
	m.Always(`Out`, `In`, meta.Pos("Clk"))
	return m
}

func TestDFF(t *testing.T) { // XXX: create proper test
	m := dff()
	if m.Out {
		t.Fatal(m.Out)
	}

	mt := m.Meta()
	in := mt.Values["In"]
	clk := mt.Values["Clk"]
	out := mt.Values["Out"]

	if out.V.Bool() {
		t.Fatal(m.Out)
	}

	now := time.Now()
	sim := NewSimulator()
	go func() {
		sim.Set(in, true, now)
		sim.Set(clk, false, now)
		sim.Set(clk, true, now.Add(1)) // trigger a <- true
		// should not trigger a <- false
		sim.Set(in, false, now.Add(2))
		sim.Set(clk, false, now.Add(3))
		sim.End()
	}()
	sim.Run()

	if !out.V.Bool() {
		t.Fatal(m.Out)
	}
}

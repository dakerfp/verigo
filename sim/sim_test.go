package sim

import (
	"testing"
	"time"

	"github.com/dakerfp/verigo/expr"
)

func TestComb(t *testing.T) {
	// build nodes
	a := NewNode(expr.F)
	b := NewNode(expr.F)
	c := NewNode(expr.F)
	d := NewNode(expr.F)

	ab := NewNode(expr.And(a, b),
		Listen(a, Anyedge, false),
		Listen(b, Anyedge, false),
	)

	cd := NewNode(expr.And(c, d),
		Listen(c, Anyedge, false),
		Listen(d, Anyedge, false),
	)

	o := NewNode(expr.And(ab, cd),
		Listen(ab, Anyedge, false),
		Listen(cd, Anyedge, false),
	)

	if o.Eval().True() {
		t.Fatal(o.Eval())
	}

	sim := NewSim()
	go func() {
		now := time.Now()
		sim.Set(a, expr.T, now)
		sim.Set(b, expr.T, now)
		sim.Set(c, expr.T, now)
		sim.Set(d, expr.T, now)
		sim.End()
	}()
	sim.Run()

	if !ab.Eval().True() {
		t.Fatal(ab.Eval())
	}

	if !cd.Eval().True() {
		t.Fatal(ab.Eval())
	}

	if !o.Eval().True() {
		t.Fatal(o.Eval(), expr.T, expr.F)
	}
}

func TestClk(t *testing.T) {
	// build nodes
	// always_ff @(posedge clk)
	//     na <= ~a;
	clk := NewNode(expr.F)
	a := NewNode(expr.F)
	na := NewNode(expr.Not(a),
		Listen(clk, Posedge, false), // only on clock trigger
	)

	if !na.Eval().True() {
		t.Fatal(na.Eval())
	}

	now := time.Now()

	sim := NewSim()
	go func() {
		sim.Set(a, expr.T, now)
		sim.End()
	}()
	sim.Run()

	if !na.Eval().True() {
		t.Fatal(na.Eval())
	}

	sim = NewSim()
	go func() {
		sim.Set(clk, expr.T, now)
		sim.End()
	}()
	sim.Run()

	if na.Eval().True() {
		t.Fatal(na.Eval())
	}
}

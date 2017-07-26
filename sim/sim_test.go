package sim

import (
	"testing"
	"time"

	"github.com/dakerfp/verigo/expr"
)

func newNode(e expr.Expr, listen ...*signal) *Node {
	n := &Node{e: e, v: e.Eval(), listen: listen, notify: nil}
	for _, sig := range listen {
		sig.n = n
	}
	return n
}

func TestComb(t *testing.T) {
	// build nodes
	a := newNode(expr.F)
	b := newNode(expr.F)
	c := newNode(expr.F)
	d := newNode(expr.F)

	ab := newNode(expr.And(a, b),
		listen(a, Anyedge, false),
		listen(b, Anyedge, false),
	)

	cd := newNode(expr.And(c, d),
		listen(c, Anyedge, false),
		listen(d, Anyedge, false),
	)

	o := newNode(expr.And(ab, cd),
		listen(ab, Anyedge, false),
		listen(cd, Anyedge, false),
	)

	if o.Eval().True() {
		t.Fatal(o.Eval())
	}

	sim := NewSimulator()
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
	clk := newNode(expr.F)
	a := newNode(expr.F)
	na := newNode(expr.Not(a),
		listen(clk, Posedge, false), // only on clock trigger
	)

	if !na.Eval().True() {
		t.Fatal(na.Eval())
	}

	now := time.Now()

	sim := NewSimulator()
	go func() {
		sim.Set(a, expr.T, now)
		sim.End()
	}()
	sim.Run()

	if !na.Eval().True() {
		t.Fatal(na.Eval())
	}

	sim = NewSimulator()
	go func() {
		sim.Set(clk, expr.T, now)
		sim.End()
	}()
	sim.Run()

	if na.Eval().True() {
		t.Fatal(na.Eval())
	}

	sim = NewSimulator()
	go func() {
		sim.Set(a, expr.T, now)
		sim.Set(clk, expr.F, now)
		sim.Set(clk, expr.T, now.Add(1)) // trigger a <- true
		// should not trigger a <- false
		sim.Set(a, expr.F, now.Add(2))
		sim.Set(clk, expr.F, now.Add(3))
		sim.End()
	}()
	sim.Run()

	if na.Eval().True() {
		t.Fatal(na.Eval())
	}
}

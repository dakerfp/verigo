package sim

import (
	"testing"
	"time"

	"github.com/dakerfp/verigo/expr"
)

func TestSim(t *testing.T) {
	// build nodes
	a := NewNode(expr.F)
	b := NewNode(expr.F)
	c := NewNode(expr.F)
	d := NewNode(expr.F)

	ab := NewNode(expr.And(&a, &b),
		a.Listen(Anyedge, false),
		b.Listen(Anyedge, false),
	)

	cd := NewNode(expr.And(&c, &d),
		c.Listen(Anyedge, false),
		d.Listen(Anyedge, false),
	)

	o := NewNode(expr.And(&ab, &cd),
		ab.Listen(Anyedge, false),
		cd.Listen(Anyedge, false),
	)

	if o.Eval().True() {
		t.Fatal(o.Eval())
	}

	var sim simulator
	go sim.Run()

	time.Sleep(time.Second / 10)
	if o.Eval().True() {
		t.Fatal(o.Eval()) // test if does not close
	}

	now := time.Now()
	sim.Set(&a, expr.T, now)
	sim.Set(&b, expr.T, now)
	sim.Set(&c, expr.T, now)
	sim.Set(&d, expr.T, now)
	sim.Wait()

	if !o.Eval().True() {
		t.Fatal(o.Eval())
	}
}

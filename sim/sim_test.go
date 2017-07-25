package sim

import (
	"testing"
	"time"

	"github.com/dakerfp/verigo/expr"
)

func TestSim(t *testing.T) {
	// build nodes
	a := NewNode(&expr.False)
	b := NewNode(&expr.False)
	c := NewNode(&expr.False)
	d := NewNode(&expr.False)

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

	now := time.Now()
	evs := make(chan event, 4)
	evs <- a.Poke(&expr.True, now)
	evs <- b.Poke(&expr.True, now)
	evs <- c.Poke(&expr.True, now)
	evs <- d.Poke(&expr.True, now)
	go run(now, evs)

	time.Sleep(time.Second)
	close(evs)

	if !o.Eval().True() {
		t.Fatal(o.Eval())
	}
}

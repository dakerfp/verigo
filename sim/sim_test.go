package sim

import (
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)
	now := time.Now()
	evs := make(chan event, 4)
	go run(&wg, now, evs)
	evs <- a.Poke(&expr.True, now)
	evs <- b.Poke(&expr.True, now)
	evs <- c.Poke(&expr.True, now)
	evs <- d.Poke(&expr.True, now)
	close(evs)
	wg.Wait()

	if !o.Eval().True() {
		t.Fatal(o.Eval())
	}
}

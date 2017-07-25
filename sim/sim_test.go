package sim

import (
	"testing"
	"time"

	"github.com/dakerfp/verigo/expr"
)

func evalAnd(v expr.Value, sigs []*signal) expr.Value {
	a := sigs[0].n.Eval().True()
	b := sigs[1].n.Eval().True()
	no := new(expr.Bool)
	*no = expr.Bool(a && b)
	return no
}

func TestSim(t *testing.T) {
	a := expr.Var{&expr.False}
	b := expr.Var{&expr.False}
	c := expr.Var{&expr.False}
	d := expr.Var{&expr.False}
	ab := expr.Var{expr.And(&a, &b)}
	cd := expr.Var{expr.And(&c, &d)}
	o := expr.Var{expr.And(&ab, &cd)}

	// build nodes
	na := node{v: a}
	nb := node{v: b}
	nc := node{v: c}
	nd := node{v: d}

	nab := node{
		v: ab,
		listen: []*signal{
			na.Listen(Anyedge, false),
			nb.Listen(Anyedge, false),
		},
	}

	ncd := node{
		v: cd,
		listen: []*signal{
			nc.Listen(Anyedge, false),
			nd.Listen(Anyedge, false),
		},
	}

	no := node{
		v: o,
		listen: []*signal{
			nab.Listen(Anyedge, false),
			ncd.Listen(Anyedge, false),
		},
	}

	now := time.Now()
	evs := make(chan event, 4)
	evs <- na.Poke(&expr.True, now)
	evs <- nb.Poke(&expr.True, now)
	evs <- nc.Poke(&expr.True, now)
	evs <- nd.Poke(&expr.True, now)
	go run(now, evs)

	time.Sleep(time.Second)
	close(evs)

	if !no.Eval().True() {
		t.Fatal(no.Eval())
	}
}

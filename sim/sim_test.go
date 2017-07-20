package sim

import (
	"testing"
	"time"
)

func evalAnd(v Value, sigs []*signal) Value {
	a := sigs[0].n.Eval().True()
	b := sigs[1].n.Eval().True()
	no := new(Bool)
	*no = Bool(a && b)
	return no
}

func TestSim(t *testing.T) {
	var a, b, c, d Bool
	var ab, cd Bool
	var o Bool

	// build nodes
	na := node{v: &a}
	nb := node{v: &b}
	nc := node{v: &c}
	nd := node{v: &d}

	nab := node{
		v: &ab,
		listen: []*signal{
			na.Listen(Anyedge, false),
			nb.Listen(Anyedge, false),
		},
		eval: evalAnd,
	}
	ncd := node{
		v: &cd,
		listen: []*signal{
			nc.Listen(Anyedge, false),
			nd.Listen(Anyedge, false),
		},
		eval: evalAnd,
	}

	no := node{
		v: &o,
		listen: []*signal{
			nab.Listen(Anyedge, false),
			ncd.Listen(Anyedge, false),
		},
		eval: evalAnd,
	}

	now := time.Now()
	True := Bool(true)
	evs := make(chan event, 4)
	evs <- na.Poke(&True, now)
	evs <- nb.Poke(&True, now)
	evs <- nc.Poke(&True, now)
	evs <- nd.Poke(&True, now)
	go run(now, evs)

	time.Sleep(time.Second)
	close(evs)

	if !no.Eval().True() {
		t.Fatal(no.Eval())
	}
}

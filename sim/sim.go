package sim

import (
	"fmt"
	"sort"
	"time"

	"github.com/dakerfp/verigo/expr"
)

type Sensivity int

const (
	Noedge Sensivity = 1 << iota
	Anyedge
	Posedge
	Negedge
	Block
)

type Node struct {
	e      expr.Expr
	v      expr.Value
	listen []*signal
	notify []*signal
}

func NewNode(e expr.Expr) *Node {
	return &Node{e: e, v: e.Eval()}
}

func (n *Node) Eval() expr.Value {
	if n.v == nil {
		n.v = n.e.Eval()
	}
	return n.v
}

func Connect(a *Node, b *Node, s Sensivity) {
	sig := listen(a, s)
	sig.n = b
	b.listen = append(b.listen, sig)
}

type signal struct {
	n *Node
	s Sensivity
}

func (sig *signal) block() bool {
	return sig.s&Block != 0
}

type event struct {
	sig *signal
	ts  time.Time
}

func listen(n *Node, s Sensivity) *signal {
	sig := &signal{nil, s}
	n.notify = append(n.notify, sig)
	return sig
}

type Simulator struct {
	eventPool []event
	blocked   []event
	now       time.Time
	scheduler chan event
}

func NewSimulator() *Simulator {
	return &Simulator{
		scheduler: make(chan event),
	}
}

func (sim *Simulator) End() {
	sim.scheduler <- event{nil, sim.now}
}

func (sim *Simulator) Run() {
	for {
		select {
		case ev, ok := <-sim.scheduler:
			if !ok {
				panic("should not close this channel")
			}
			if ev.sig != nil {
				sim.putEvent(ev) // is a valid event
				continue
			}
			for sim.handleAnyEvent() {
				// execute until has no event left
			}
			return
		default:
			sim.handleAnyEvent()
		}
	}
}

func (sim *Simulator) Set(n *Node, v expr.Value, ts time.Time) {
	n.e = v
	sim.scheduler <- event{
		&signal{n, Anyedge},
		ts,
	}
}

func (sim *Simulator) updateNodeValue(n *Node, v expr.Value) {
	if expr.Eq(n.v, v) {
		return
	}
	n.v = v
	for _, sig := range n.notify {
		posedge := v.True()
		switch sig.s & ^Block {
		case Noedge:
			continue
		case Posedge:
			if !posedge {
				continue
			}
		case Negedge:
			if posedge {
				continue
			}
		case Anyedge:
			// just proceeed
		}
		// XXX: Add delay
		sim.putEvent(event{sig, sim.now})
	}
}

func (sim *Simulator) handleAnyEvent() (any bool) {
	switch {
	case len(sim.eventPool) > 0:
		sim.handleNextEvent()
		any = true
	case len(sim.blocked) > 0:
		sim.handleBlockedEvents()
		any = true
	default:
	}
	return
}

func (sim *Simulator) handleNextEvent() {
	ev := sim.eventPool[0] // first event

	if ev.ts.Before(sim.now) {
		panic(fmt.Errorf("ev.ts %v should never be before %v", ev.ts, sim.now))
	}

	if ev.ts.After(sim.now) && len(sim.blocked) > 0 {
		// ensure all blocked events are handled before
		sim.handleBlockedEvents()
		return
	}

	sim.eventPool = sim.eventPool[1:] // pop
	sim.now = ev.ts                   // step simulation time

	if ev.sig.block() {
		// append event in blocked queue
		sim.blocked = append(sim.blocked, ev)
	} else {
		// execute now
		n := ev.sig.n
		sim.updateNodeValue(n, n.e.Eval())
	}
}

func (sim *Simulator) handleBlockedEvents() {
	values := make([]expr.Value, len(sim.blocked))
	// eval
	for i, ev := range sim.blocked {
		values[i] = ev.sig.n.e.Eval()
	}
	// update values and schedule next evs
	for i, ev := range sim.blocked {
		n := ev.sig.n
		sim.updateNodeValue(n, values[i])
	}
	sim.blocked = nil
}

func (sim *Simulator) putEvent(ev event) {
	// sort by ts and then, blocking always comes later
	sim.eventPool = append(sim.eventPool, ev)
	sort.Slice(sim.eventPool, func(i, j int) bool {
		ei := sim.eventPool[i]
		ej := sim.eventPool[j]
		if ei.ts.Before(ej.ts) {
			return true
		} else if ei.ts.After(ej.ts) {
			return false
		}
		return !ei.sig.block()
	})
}

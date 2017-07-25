package sim

import (
	"fmt"
	"sort"
	"time"

	"github.com/dakerfp/verigo/expr"
)

type Sensivity int

const (
	Noedge Sensivity = iota
	Anyedge
	Posedge
	Negedge
)

type Node struct {
	e      expr.Expr
	v      expr.Value
	listen []*signal
	notify []*signal
}

func NewNode(e expr.Expr, listen ...*signal) *Node {
	n := &Node{e: e, v: e.Eval(), listen: listen, notify: nil}
	for _, sig := range listen {
		sig.n = n
	}
	return n
}

func (n *Node) Eval() expr.Value {
	if n.v == nil {
		n.v = n.e.Eval()
	}
	return n.v
}

type signal struct {
	n     *Node
	s     Sensivity
	block bool
}

type event struct {
	sig *signal
	ts  time.Time
}

func updateNodeEvent(n *Node, v expr.Value, ts time.Time) event {
	n.e = v
	return event{
		&signal{n, Anyedge, false},
		ts,
	}
}

func Listen(n *Node, s Sensivity, block bool) *signal {
	sig := &signal{nil, s, block}
	n.notify = append(n.notify, sig)
	return sig
}

func deferUpdate(n *Node, v expr.Value, now time.Time, sim *simulator) {
	if expr.Eq(n.v, v) {
		return
	}
	n.v = v
	for _, sig := range n.notify {
		posedge := v.True()
		switch sig.s {
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
		}
		// XXX: Add delay
		sim.putEvent(event{sig, now})
	}
}

func update(n *Node, now time.Time, sim *simulator) {
	deferUpdate(n, n.e.Eval(), now, sim)
}

type simulator struct {
	eventPool []event
	blocked   []event
	now       time.Time
	scheduler chan event
}

func NewSim() *simulator { // XXX
	return &simulator{
		scheduler: make(chan event),
	}
}

func (sim *simulator) End() {
	sim.scheduler <- event{nil, sim.now}
}

func (sim *simulator) Run() {
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
			for sim.executeAny() {
				// execute until has no event left
			}
			return
		default:
			sim.executeAny()
		}
	}
}

func (sim *simulator) Set(n *Node, v expr.Value, ts time.Time) {
	sim.scheduler <- updateNodeEvent(n, v, ts)
}

func (sim *simulator) executeAny() (any bool) {
	switch {
	case len(sim.eventPool) > 0:
		sim.executeEvent()
		any = true
	case len(sim.blocked) > 0:
		sim.updateAllBlockedEvents()
		any = true
	default:
	}
	return
}

func (sim *simulator) executeEvent() {
	ev := sim.eventPool[0]
	sim.eventPool = sim.eventPool[1:]

	if ev.ts.Before(sim.now) {
		panic(fmt.Errorf("ev.ts %v should never be before %v", ev.ts, sim.now))
	}

	if ev.ts.After(sim.now) {
		// ensure all blocked events execute before
		sim.updateAllBlockedEvents()
		sim.blocked = nil
		sim.now = ev.ts // step simulation time
	}

	if ev.sig.block {
		// append event in blocked queue
		sim.blocked = append(sim.blocked, ev)
	} else {
		// execute now
		update(ev.sig.n, sim.now, sim)
	}
}

func (sim *simulator) updateAllBlockedEvents() {
	values := make([]expr.Value, len(sim.blocked))
	// eval
	for i, ev := range sim.blocked {
		values[i] = ev.sig.n.e.Eval()
	}
	// update values and schedule next evs
	for i, ev := range sim.blocked {
		deferUpdate(ev.sig.n, values[i], sim.now, sim)
	}
}

func (sim *simulator) putEvent(ev event) {
	// sort by ts and then, leave blocking later
	sim.eventPool = append(sim.eventPool, ev)
	sort.Slice(sim.eventPool, func(i, j int) bool {
		ei := sim.eventPool[i]
		ej := sim.eventPool[j]
		if ei.ts.Before(ej.ts) {
			return true
		} else if ei.ts.After(ej.ts) {
			return false
		}
		return !ei.sig.block
	})
}

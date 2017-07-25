package sim

import (
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

type signal struct {
	n     *node
	s     Sensivity
	block bool
}

type node struct {
	expr.Var
	listen []*signal
	notify []*signal
}

type event struct {
	sig *signal
	nv  expr.Value
	ts  time.Time
}

func NewNode(e expr.Expr, listen ...*signal) node {
	var n node
	n.V = e
	n.listen = listen
	return n
}

func (n *node) poke(v expr.Value, ts time.Time) event {
	n.Update(v)
	return event{
		&signal{n, Anyedge, false},
		v,
		ts,
	}
}

func (n *node) Listen(s Sensivity, block bool) *signal {
	sig := &signal{n, s, block}
	n.notify = append(n.notify, sig)
	return sig
}

func (n *node) deferUpdate(v expr.Value, now time.Time, sim *simulator) {
	if !n.Update(v) {
		return
	}
	for _, sig := range n.notify {
		posedge := n.Eval().True()
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
		sim.putEvent(event{sig, n.Eval(), now})
	}
}

func (n *node) update(now time.Time, sim *simulator) {
	v := n.Eval()
	n.deferUpdate(v, now, sim)
}

type simulator struct {
	eventPool []event
	blocked   []event
	now       time.Time
	scheduler chan event
	wait      chan bool
	stop      bool
}

func (sim *simulator) Run() {
	sim.scheduler = make(chan event, 3)
	sim.wait = make(chan bool)
	for {
		select {
		case ev, ok := <-sim.scheduler:
			if ok {
				sim.putEvent(ev)
				continue
			}
			for sim.executeAny() {
				// execute until has no event left
			}
			sim.wait <- true
			return
		default:
			sim.executeAny()
		}
	}
}

func (sim *simulator) Wait() {
	close(sim.scheduler) // tell run to stop
	<-sim.wait           // wait run until is finished
}

func (sim *simulator) Set(n *node, v expr.Value, ts time.Time) {
	sim.scheduler <- n.poke(v, ts)
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
		panic("ev.ts should never be before now")
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
		ev.sig.n.update(sim.now, sim)
	}
}

func (sim *simulator) updateAllBlockedEvents() {
	values := make([]expr.Value, len(sim.blocked))
	// eval
	for i, ev := range sim.blocked {
		values[i] = ev.sig.n.Eval()
	}
	// update values and schedule next evs
	for i, ev := range sim.blocked {
		ev.sig.n.deferUpdate(values[i], sim.now, sim)
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

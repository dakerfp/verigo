package sim

import (
	"sort"
	"sync"
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

func (n *node) Poke(v expr.Value, ts time.Time) event {
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

func (n *node) deferUpdate(v expr.Value, now time.Time, scheduler chan<- event) {
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
		scheduler <- event{sig, n.Eval(), now} // XXX: Add delay
	}
}

func (n *node) update(now time.Time, scheduler chan<- event) {
	v := n.Eval()
	n.deferUpdate(v, now, scheduler)
}

type simulator struct {
	eventPool []event
	blocked   []event
	now       time.Time
}

func run(wg *sync.WaitGroup, startTime time.Time, scheduler chan event) {
	defer wg.Done()
	var sim simulator
	sim.now = startTime
	for {
		select {
		case ev, ok := <-scheduler:
			if !ok {
				return
			}
			sim.putEvent(ev)

		default:
			if len(sim.eventPool) == 0 {
				continue
			}
			sim.executeEvents(scheduler)
		}
	}
}

func (sim *simulator) executeEvents(scheduler chan event) {
	ev := sim.eventPool[0]
	sim.eventPool = sim.eventPool[1:]

	if ev.ts.Before(sim.now) {
		panic("ev.ts should never be before now")
	}

	if ev.ts.After(sim.now) {
		sim.updateAllBlockedEvs(scheduler)
		sim.blocked = nil
		sim.now = ev.ts // step simulation time
	}

	if ev.sig.block {
		sim.blocked = append(sim.blocked, ev)
	} else {
		ev.sig.n.update(sim.now, scheduler)
	}
}

func (sim *simulator) updateAllBlockedEvs(scheduler chan<- event) {
	values := make([]expr.Value, len(sim.blocked))
	// eval
	for i, ev := range sim.blocked {
		values[i] = ev.sig.n.Eval()
	}
	// update values and schedule next evs
	for i, ev := range sim.blocked {
		ev.sig.n.deferUpdate(values[i], sim.now, scheduler)
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

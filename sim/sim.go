package sim

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/dakerfp/verigo/meta"
)

type signal struct {
	n *meta.Node
	s meta.Sensivity
}

func (sig *signal) block() bool {
	return sig.s.Block()
}

type event struct {
	sig *signal
	ts  time.Time
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

func (sim *Simulator) Set(n *meta.Node, x interface{}, ts time.Time) {
	v := reflect.ValueOf(x)
	n.Update = func() reflect.Value { return v }
	sim.scheduler <- event{
		&signal{n, meta.Anyedge},
		ts,
	}
}

func (sim *Simulator) updateNodeValue(n *meta.Node, v reflect.Value) {
	if reflect.DeepEqual(n.V, v) { // XXX: implement Eq
		return
	}
	n.V = v
	for _, edge := range n.Notify {
		posedge := v.Bool() // XXX: may fail
		switch edge.Sensivity.Edge() {
		case meta.Noedge:
			continue
		case meta.Posedge:
			if !posedge {
				continue
			}
		case meta.Negedge:
			if posedge {
				continue
			}
		case meta.Anyedge:
			// just proceeed
		}
		// XXX: Add delay
		sig := &signal{edge.To, edge.Sensivity} // XXX
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
		sim.updateNodeValue(n, n.Update())
	}
}

func (sim *Simulator) handleBlockedEvents() {
	values := make([]reflect.Value, len(sim.blocked))
	// eval
	for i, ev := range sim.blocked {
		values[i] = ev.sig.n.Update()
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

package sim

import (
	"sort"
	"time"
)

type Value interface {
	True() bool
	Update(Value) bool
}

type Bool bool

func (b *Bool) True() bool {
	return bool(*b)
}

func (b *Bool) Update(v Value) bool {
	if b.True() == v.True() {
		return false
	}
	*b = !*b
	return true
}

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
	v      Value
	listen []*signal
	eval   func(Value, []*signal) Value
	notify []*signal
}

type event struct {
	sig *signal
	nv  Value
	ts  time.Time
}

func (n *node) Poke(v Value, ts time.Time) event {
	n.v = v
	return event{
		&signal{n, Anyedge, false},
		v,
		ts,
	}
}

func (n *node) Eval() Value {
	if n.eval != nil {
		return n.eval(n.v, n.listen)
	}
	switch len(n.listen) {
	case 0:
		return n.v
	case 1:
		return n.listen[0].n.Eval()
	default:
		panic(n.listen)
	}
}

func (n *node) Listen(s Sensivity, block bool) *signal {
	sig := &signal{n, s, block}
	n.notify = append(n.notify, sig)
	return sig
}

func (n *node) deferUpdate(v Value, now time.Time, scheduler chan<- event) {
	if !n.v.Update(v) {
		return
	}
	for _, sig := range n.notify {
		posedge := n.v.True()
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
		scheduler <- event{sig, n.v, now} // XXX: Add delay
	}
}

func (n *node) update(now time.Time, scheduler chan<- event) {
	v := n.Eval()
	n.deferUpdate(v, now, scheduler)
}

func updateAllBlockedEvs(blocked []event, now time.Time, scheduler chan<- event) {
	values := make([]Value, len(blocked))
	// eval
	for i, ev := range blocked {
		values[i] = ev.sig.n.Eval()
	}
	// update values and schedule next evs
	for i, ev := range blocked {
		ev.sig.n.deferUpdate(values[i], now, scheduler)
	}
}

func run(startTime time.Time, scheduler chan event) {
	var eventPool []event
	var blocked []event
	now := startTime
	for {
		select {
		case ev, ok := <-scheduler:
			if !ok {
				return
			}

			// sort by ts and then, leave blocking later
			eventPool = append(eventPool, ev)
			sort.Slice(eventPool, func(i, j int) bool {
				ei := eventPool[i]
				ej := eventPool[j]
				if ei.ts.Before(ej.ts) {
					return true
				} else if ei.ts.After(ej.ts) {
					return false
				}
				return !ei.sig.block
			})

		default:
			if len(eventPool) != 0 {
				return
			}
			ev := eventPool[0]
			eventPool = eventPool[1:]

			if ev.ts.Before(now) {
				panic("ev.ts should never be before now")
			}

			if ev.ts.After(now) {
				updateAllBlockedEvs(blocked, now, scheduler)
				blocked = nil
				now = ev.ts // step simulation time
			}

			if ev.sig.block {
				blocked = append(blocked, ev)
			} else {
				ev.sig.n.update(now, scheduler)
			}
		}
	}
}

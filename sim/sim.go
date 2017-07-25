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

func updateAllBlockedEvs(blocked []event, now time.Time, scheduler chan<- event) {
	values := make([]expr.Value, len(blocked))
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

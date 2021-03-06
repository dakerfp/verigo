package meta

import "reflect"

type Sensivity int

const (
	Posedge Sensivity = 1 << iota
	Negedge
	Block
	Anyedge = Posedge | Negedge
	Noedge  = 0
)

func (s Sensivity) String() string {
	var repr string
	if s.Block() {
		repr = "block "
	}
	switch s.Edge() {
	case Noedge:
		repr += "none"
	case Posedge:
		repr += "pos"
	case Negedge:
		repr += "neg"
	case Anyedge:
		repr += "any"
	}
	return repr
}

func (s Sensivity) Block() bool {
	return s&Block != 0
}

func (s Sensivity) Edge() Sensivity {
	return s &^ Block
}

type UpdateFunc func() reflect.Value

type Edge struct {
	From, To *Node
	Sensivity
}

type Node struct {
	T              reflect.Type
	V              reflect.Value
	Notify, Listen []*Edge
	Update         UpdateFunc
	Name           string
}

func Connect(from, to *Node, s Sensivity) {
	e := &Edge{from, to, s}
	from.Notify = append(from.Notify, e)
	to.Listen = append(to.Listen, e)
}

type Signal struct {
	Name string
	Sensivity
}

func Neg(name string) Signal {
	return Signal{name, Negedge}
}

func Pos(name string) Signal {
	return Signal{name, Posedge}
}

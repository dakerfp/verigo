package meta

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
)

type Module interface {
	Meta() *Mod
	Sub(...Module)
}

type Mod struct {
	Name string

	subs    []Module
	Values  map[string]*Node
	Inputs  []*Node
	Outputs []*Node
}

func Init(m Module) {
	meta := m.Meta()
	// build it bottom up
	for _, sub := range meta.subs {
		Init(sub)
	}
	data := reflect.Indirect(reflect.ValueOf(m))
	t := data.Type()
	meta.Name = t.Name()
	meta.Values = make(map[string]*Node)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// ignore Module embedding
		if field.Type == reflect.TypeOf(meta).Elem() {
			continue
		}

		if string(field.Tag) == "submodule" {
			// XXX: init automatically
			continue
		}

		v := data.FieldByIndex(field.Index)
		t := v.Type()
		n := &Node{T: t, V: v, Name: field.Name}
		meta.Values[field.Name] = n

		switch string(field.Tag) {
		case "":
			// ignore field
		case "input":
			meta.Inputs = append(meta.Inputs, n)
		case "output":
			meta.Outputs = append(meta.Outputs, n)
		default:
			panic(fmt.Errorf("io tag %q not supported in field %q", field.Tag.Get("io"), field.Name))
		}

	}
}

func (m *Mod) Meta() *Mod {
	return m
}

func (m *Mod) Assign(recv interface{}, f interface{}) {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic("assign does not gets function")
	}

	if t.NumIn() > 0 {
		panic("edge triggedered not yet supported") // XXX
	}

	if t.NumOut() != 1 {
		panic("assign must return a single function")
	}

	recvt := reftype(recv)
	rett := t.Out(0)
	if !rett.ConvertibleTo(recvt) {
		panic("assign to different types")
	}
}

func reftype(v interface{}) reflect.Type {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Ptr {
		panic("wire element must be a pointer to another based type")
	}

	e := reflect.Indirect(ptr)
	if e.Kind() == reflect.Ptr {
		panic("wire element must be a pointer to another based type")
	}

	return e.Type()
}

func (m *Mod) Sub(subs ...Module) {
	m.subs = append(m.subs, subs...)
}

func (m *Mod) Bind(recv string, x string) {
	if err := m.parseExpr(recv, nil, x); err != nil {
		panic(err)
	}
}

func (m *Mod) Always(recv string, x string, signals ...Signal) {
	if err := m.parseExpr(recv, signals, x); err != nil {
		panic(err)
	}
}

var (
	ErrInvalidIdentifier = errors.New("invalid identifier")
)

func (m *Mod) assembleExpr(e ast.Expr) (update UpdateFunc, dependsOn []string, err error) {
	switch e.(type) {
	case *ast.Ident:
		identExpr := e.(*ast.Ident)
		n, ok := m.Values[identExpr.Name]
		if !ok {
			err = ErrInvalidIdentifier
		}
		dependsOn = []string{identExpr.Name}
		update = func() reflect.Value {
			return n.V // XXX: test if closure closes in n
		}
	case *ast.BinaryExpr:
		binaryExpr := e.(*ast.BinaryExpr)
		updateX, depX, errX := m.assembleExpr(binaryExpr.X)
		if errX != nil {
			err = errX
			return
		}
		updateY, depY, errY := m.assembleExpr(binaryExpr.Y)
		if errY != nil {
			err = errY
			return
		}
		switch binaryExpr.Op {
		case token.LAND:
			update = func() reflect.Value {
				x := updateX()
				y := updateY()
				return reflect.ValueOf(x.Bool() && y.Bool())
			}
		default:
			panic(binaryExpr.Op.String())
		}

		dependsOn = append(depX, depY...)
	default:
		panic(reflect.TypeOf(e))
	}
	return
}

func (m *Mod) parseExpr(recv string, signals []Signal, x string) (err error) {
	var exp ast.Expr
	exp, err = parser.ParseExpr(x)
	if err != nil {
		return
	}
	recvN := m.Values[recv]
	update, deps, err := m.assembleExpr(exp)
	recvN.Update = update
	for _, dep := range deps {
		n := m.Values[dep]
		Connect(n, recvN, Anyedge) // XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx
	}
	return err
}

// func (m *Mod) Wire(wire ...interface{}) {
// 	if len(wire) <= 1 {
// 		panic("wire must have at least 2 elements")
// 	}

// 	wt0 := reftype(wire[0])
// 	for _, we := range wire[1:] {
// 		wt := reftype(we)
// 		if !wt.ConvertibleTo(wt0) {
// 			panic(wt)
// 		}
// 	}

// 	m.wires = append(m.wires, wire)
// }

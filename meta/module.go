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

	subs   []Module
	nodes  map[reflect.Value]*Node
	values map[string]*Node
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
	meta.nodes = make(map[reflect.Value]*Node)
	meta.values = make(map[string]*Node)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// ignore Module embedding
		if field.Type == reflect.TypeOf(meta).Elem() {
			continue
		}

		switch string(field.Tag) {
		case "submodule":
			// XXX: init automatically
			continue
		case "":
			// ignore field
		case "input":
			// meta.inputs[field.Name] = data.FieldByIndex(field.Index)
		case "output":
			// meta.outputs[field.Name] = data.FieldByIndex(field.Index)
		default:
			panic(fmt.Errorf("io tag %q not supported in field %q", field.Tag.Get("io"), field.Name))
		}
		v := data.FieldByIndex(field.Index)
		n := &Node{V: v}
		meta.values[field.Name] = n
		meta.nodes[v] = n
	}
	fmt.Println(data, t, meta)
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

	fv := reflect.ValueOf(f)
	fmt.Println(t.NumIn())
	fmt.Println(t.NumOut())
	fmt.Println(fv)
	fmt.Println(reflect.TypeOf(m))
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

func (m *Mod) Bind(recv interface{}, x string) {
	if err := m.parseExpr(recv, nil, x); err != nil {
		panic(err)
	}
}

func (m *Mod) Always(recv interface{}, x string, sigs ...Signal) {
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
		n, ok := m.values[identExpr.Name]
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

func (m *Mod) parseExpr(recv interface{}, signals []Signal, x string) (err error) {
	var exp ast.Expr
	exp, err = parser.ParseExpr(x)
	if err != nil {
		return
	}
	r := reflect.ValueOf(recv)
	recvN := m.nodes[r]
	update, deps, err := m.assembleExpr(exp)
	for _, dep := range deps {
		n := m.values[dep]
		Connect(n, recvN)
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
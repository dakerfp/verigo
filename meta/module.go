package meta

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
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

func (m *Mod) assign(recv interface{}, f interface{}) {
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

func (m *Mod) Always(recv string, x string, signals ...Signal) {
	if err := m.parseExpr(recv, signals, x); err != nil {
		panic(err)
	}
}

var (
	ErrInvalidIdentifier = errors.New("invalid identifier")
)

func (m *Mod) assembleExpr(e ast.Expr) (update UpdateFunc, dependsOn []Signal, err error) {
	switch e.(type) {
	case *ast.Ident:
		identExpr := e.(*ast.Ident)
		n, ok := m.Values[identExpr.Name]
		if !ok {
			err = ErrInvalidIdentifier
		}
		update = func() reflect.Value {
			return n.V // XXX: test if closure closes in n
		}
		dependsOn = []Signal{
			Signal{
				Name:      identExpr.Name,
				Sensivity: Anyedge,
			},
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
		case token.ADD:
			update = func() reflect.Value {
				x := updateX()
				y := updateY()
				return Add(x.Interface(), y.Interface())
			}
		default:
			panic(binaryExpr.Op.String())
		}

		dependsOn = append(depX, depY...)
	case *ast.BasicLit:
		// basic lit is constant no need to update dependsOn list
		basicLitExpr := e.(*ast.BasicLit)
		switch basicLitExpr.Kind {
		case token.INT:
			var i int64
			i, err = strconv.ParseInt(basicLitExpr.Value, 10, 64)
			update = func() reflect.Value {
				return reflect.ValueOf(i)
			}
		default:
			panic(basicLitExpr)
		}
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

	// if there is any explicit signal, use it
	// otherwise, use deps as if it is combinational
	if len(signals) == 0 {
		signals = deps
	}
	for _, signal := range signals {
		n := m.Values[signal.Name]
		Connect(n, recvN, signal.Sensivity)
	}
	return err
}

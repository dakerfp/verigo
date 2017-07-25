package comp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	// "go/types"
	"reflect"
)

type DataType int

const (
	NoType DataType = iota
	BoolType
	UintType
	IntType
)

var basicTypeToDataType = map[string]DataType{
	"bool": BoolType,
	"uint": UintType,
	"int":  IntType,
}

type Module struct {
	Name     string
	Inputs   map[string]DataType
	Outputs  map[string]DataType
	Internal map[string]DataType
	Signals  map[string]DataType
}

func genModule(tpspec *ast.TypeSpec) (m *Module) {
	if tpspec == nil {
		return
	}

	switch tpspec.Type.(type) {
	default:
		return
	case *ast.StructType:

	}

	m = &Module{}
	m.Name = tpspec.Name.String()
	m.Inputs = make(map[string]DataType)
	m.Outputs = make(map[string]DataType)
	m.Internal = make(map[string]DataType)
	m.Signals = make(map[string]DataType)

	str := tpspec.Type.(*ast.StructType)
	if !str.Struct.IsValid() {
		panic(str.Struct)
	}

	for _, field := range str.Fields.List {
		var dtype DataType
		ident := field.Type.(*ast.Ident)
		if ident == nil {
			panic(reflect.TypeOf(ident))
		}

		tag := ""
		if field.Tag != nil {
			tag = field.Tag.Value
		}
		dtype = basicTypeToDataType[ident.Name]

		for _, name := range field.Names {
			switch tag {
			case "`input`":
				m.Inputs[name.Name] = dtype
			case "`output`":
				fmt.Println(">>>", tag, name.Name)
				m.Outputs[name.Name] = dtype
			default:
				fmt.Println(">>>", tag)
				m.Internal[name.Name] = dtype
			}
			m.Signals[name.Name] = dtype
		}
	}
	return
}

func identName(expr ast.Expr) string {
	return expr.(*ast.Ident).Name
}

func genModuleExpression(m *Module, modRef string, expr ast.Expr) DataType {
	switch expr.(type) {
	case *ast.SelectorExpr:
		selectorExpr := expr.(*ast.SelectorExpr)
		if identName(selectorExpr.X) != modRef {
			panic("invalid reference to module")
		}
		v := identName(selectorExpr.Sel)
		dt, ok := m.Signals[v]
		if !ok {
			panic("no signal in module")
		}
		return dt
	default:
		fmt.Println(reflect.TypeOf(expr), expr)
	}
	return NoType
}

func genModuleStmt(m *Module, modRef string, stmt ast.Stmt) DataType {
	switch stmt.(type) {
	case *ast.ReturnStmt:
		retStmt := stmt.(*ast.ReturnStmt)
		if len(retStmt.Results) != 1 {
			panic("more than one stmt")
		}
		expr := retStmt.Results[0]
		return genModuleExpression(m, modRef, expr)
	case *ast.IfStmt:
		ifStmt := stmt.(*ast.IfStmt)
		expr := ifStmt.Cond
		selDt := genModuleExpression(m, modRef, expr)
		if selDt != BoolType {
			panic("invalid if condition")
		}
		ifDt := genModuleStmt(m, modRef, ifStmt.Body)
		elseDt := genModuleStmt(m, modRef, ifStmt.Else)
		if ifDt != elseDt {
			panic("mismatch return type")
		}
		return ifDt
	case *ast.BlockStmt:
		blockStmt := stmt.(*ast.BlockStmt)
		if len(blockStmt.List) != 1 {
			panic("unsupported multistmt block")
		}
		return genModuleStmt(m, modRef, blockStmt.List[0])
	default:
		fmt.Println(reflect.TypeOf(stmt))
	}
	return NoType
}

func parse(filename string) error {
	fset := token.NewFileSet()
	fast, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	modules := make(map[string]*Module)

	// Get module declaration
	for _, decl := range fast.Decls {
		n := ast.Node(decl)
		switch n.(type) {
		case *ast.GenDecl:
			decl := n.(*ast.GenDecl)
			for _, spec := range decl.Specs {
				tpspec := spec.(*ast.TypeSpec)
				if tpspec == nil {
					continue
				}
				m := genModule(tpspec)
				modules[m.Name] = m
			}
		}
	}
	fmt.Println(modules)

	for _, decl := range fast.Decls {
		n := ast.Node(decl)
		switch n.(type) {
		case *ast.FuncDecl:
			fndecl := n.(*ast.FuncDecl)
			fntype := fndecl.Type
			if fndecl.Recv.NumFields() != 1 {
				continue
			}
			recvType := fndecl.Recv.List[0].Type.(*ast.Ident)
			mod, ok := modules[recvType.Name]
			if !ok {
				fmt.Println(recvType, mod)
				continue
			}

			fname := fndecl.Name.Name
			_, ok = mod.Inputs[fname]
			if ok {
				panic("overriding input")
			}

			outFieldType, ok := mod.Outputs[fname]
			if !ok {
				fmt.Println(fname, mod)
				panic("no output specified")
			}

			results := fndecl.Type.Results
			if results.NumFields() != 1 {
				panic("shoud return single type")
			}

			rettype := results.List[0].Type.(*ast.Ident)
			rDataType, ok := basicTypeToDataType[rettype.Name]
			if !ok {
				panic("return type not supported")
			}

			if rDataType != outFieldType {
				panic("different return type from spec")
			}

			params := fntype.Params
			if params.NumFields() != 0 {
				continue
			}

			recvNames := fndecl.Recv.List[0].Names
			if len(recvNames) != 1 {
				panic("recv must have a name")
			}
			recvName := recvNames[0].Name

			for _, stmt := range fndecl.Body.List {
				fmt.Println(stmt)
			}

			stmts := fndecl.Body.List
			if len(stmts) != 1 {
				panic("too many stmts")
			}
			stmt := stmts[0]
			dt := genModuleStmt(mod, recvName, stmt)

			if dt != mod.Signals[fname] {
				fmt.Println("ERR", dt, mod.Signals[fname], fname)
				// panic("mismatching return type", dt)
			}

			// mtype := params.List[0]
		}
	}

	// for _, decl := range fast.Decls {
	// 	n := ast.Node(decl)
	// 	switch n.(type) {
	// 	case *ast.FuncDecl:
	// 		fn := n.(*ast.FuncDecl)
	// 		fmt.Println("Function:", fn.Name)
	// 	case *ast.GenDecl:
	// 		decl := n.(*ast.GenDecl)
	// 		fmt.Println(reflect.TypeOf(decl.Specs))
	// 		for _, spec := range decl.Specs {
	// 			fmt.Println(spec)
	// 		}
	// 	case *ast.StructType:
	// 		st := n.(*ast.StructType)
	// 		fmt.Println(st)
	// 	default:
	// 		fmt.Println(reflect.TypeOf(n))
	// 	}
	// }

	return err
}

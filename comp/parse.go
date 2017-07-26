package comp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
)

type Parser struct {
	fileSet    *token.FileSet
	modules    map[string]*Module
	functions  map[string]*Function
	currmodule *Module
}

func (p *Parser) parseTopLevel(n ast.Node) (err error) {
	switch n.(type) {
	case *ast.FuncDecl:
		err = p.parseFunction(n.(*ast.FuncDecl))
	case *ast.GenDecl:
		decl := n.(*ast.GenDecl)
		for _, spec := range decl.Specs {
			switch spec.(type) {
			case *ast.TypeSpec:
				err = p.parseModule(spec.(*ast.TypeSpec))
			default:
				err = fmt.Errorf("invalid spec %v", decl)
			}
			if err != nil {
				return
			}
		}
	default:
		err = fmt.Errorf("invalid node %v", n)
	}
	return
}

func (p *Parser) parseStructField(field *ast.Field) error {
	ident := field.Type.(*ast.Ident)
	if ident == nil {
		return fmt.Errorf("unexpected type %v", reflect.TypeOf(ident))
	}

	tag := ""
	if field.Tag != nil {
		tag = field.Tag.Value
	}

	dtype := basicTypeToDataType[ident.Name]

	for _, name := range field.Names {
		switch tag {
		case "`input`":
			p.currmodule.Inputs[name.Name] = dtype
		case "`output`":
			p.currmodule.Outputs[name.Name] = dtype
		default:
			p.currmodule.Internal[name.Name] = dtype
		}
		p.currmodule.Signals[name.Name] = dtype
	}
	return nil
}

func (p *Parser) parseModule(spec *ast.TypeSpec) error {
	switch spec.Type.(type) {
	case *ast.StructType:
		m := &Module{
			Name:     spec.Name.String(),
			Inputs:   make(map[string]DataType),
			Outputs:  make(map[string]DataType),
			Internal: make(map[string]DataType),
			Signals:  make(map[string]DataType),
		}

		str := spec.Type.(*ast.StructType)
		if !str.Struct.IsValid() {
			return fmt.Errorf("invalid struct %v", str.Struct)
		}

		p.currmodule = m
		defer func() { p.currmodule = nil }()
		for _, field := range str.Fields.List {
			if err := p.parseStructField(field); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("type spec for %v is not implemented", spec.Type)
	}
	return nil
}

func idname(expr ast.Expr) string {
	return expr.(*ast.Ident).Name
}

func (p *Parser) parseParams(fl *ast.FieldList) (dts map[string]DataType, err error) {
	dts = make(map[string]DataType)
	for _, field := range fl.List {
		tp := basicTypeToDataType[idname(field.Type)]
		for _, name := range field.Names {
			dts[idname(name)] = tp
		}
	}
	return
}

func (p *Parser) parseCombFunc(funcDecl *ast.FuncDecl) error {
	name := idname(funcDecl.Name)
	params, err := p.parseParams(funcDecl.Type.Params)
	p.functions[name] = &Function{
		Name:   name,
		Params: params,
	}
	return err
}

func (p *Parser) parseModuleFunc(funcDecl *ast.FuncDecl) error {
	return nil
	// fntype := fndecl.Type
	// recvType := fndecl.Recv.List[0].Type.(*ast.Ident)
	// mod, ok := modules[recvType.Name]
	// if !ok {
	// 	fmt.Println(recvType, mod)
	// 	continue
	// }

	// fname := fndecl.Name.Name
	// _, ok = mod.Inputs[fname]
	// if ok {
	// 	panic("overriding input")
	// }

	// outFieldType, ok := mod.Outputs[fname]
	// if !ok {
	// 	fmt.Println(fname, mod)
	// 	panic("no output specified")
	// }

	// results := fndecl.Type.Results
	// if results.NumFields() != 1 {
	// 	panic("shoud return single type")
	// }

	// rettype := results.List[0].Type.(*ast.Ident)
	// rDataType, ok := basicTypeToDataType[rettype.Name]
	// if !ok {
	// 	panic("return type not supported")
	// }

	// if rDataType != outFieldType {
	// 	panic("different return type from spec")
	// }

	// params := fntype.Params
	// if params.NumFields() != 0 {
	// 	continue
	// }

	// recvNames := fndecl.Recv.List[0].Names
	// if len(recvNames) != 1 {
	// 	panic("recv must have a name")
	// }
	// recvName := recvNames[0].Name

	// for _, stmt := range fndecl.Body.List {
	// 	fmt.Println(stmt)
	// }

	// stmts := fndecl.Body.List
	// if len(stmts) != 1 {
	// 	panic("too many stmts")
	// }
	// stmt := stmts[0]
	// dt := genModuleStmt(mod, recvName, stmt)

	// if dt != mod.Signals[fname] {
	// 	fmt.Println("ERR", dt, mod.Signals[fname], fname)
	// 	// panic("mismatching return type", dt)
	// }

	// // mtype := params.List[0]
	// return nil
}

func (p *Parser) parseFunction(funcDecl *ast.FuncDecl) (err error) {
	switch funcDecl.Recv.NumFields() {
	case 0:
		err = p.parseCombFunc(funcDecl)
	case 1:
		err = p.parseModuleFunc(funcDecl)
	default:
		err = fmt.Errorf("invalid num fields in recv %v", funcDecl)
	}
	return
}

func (p *Parser) parseExpr(expr ast.Expr) (dt DataType, err error) {
	var ok bool
	switch expr.(type) {
	case *ast.SelectorExpr:
		selectorExpr := expr.(*ast.SelectorExpr)
		if idname(selectorExpr.X) != p.currmodule.Name {
			err = fmt.Errorf("invalid reference to module")
		}
		v := idname(selectorExpr.Sel)
		dt, ok = p.currmodule.Signals[v]
		if !ok {
			err = fmt.Errorf("no signal in module")
		}
		return
	default:
		err = fmt.Errorf("%v: %v", reflect.TypeOf(expr), expr)
	}
	return
}

func (p *Parser) parseStmt(stmt ast.Stmt) (DataType, error) {
	switch stmt.(type) {
	case *ast.ReturnStmt:
		retStmt := stmt.(*ast.ReturnStmt)
		if len(retStmt.Results) != 1 {
			panic("more than one stmt")
		}
		return p.parseExpr(retStmt.Results[0])
	case *ast.IfStmt:
		ifStmt := stmt.(*ast.IfStmt) // XXX
		selDt, err := p.parseExpr(ifStmt.Cond)
		if err != nil {
			panic(err)
		}
		if selDt != BoolType {
			panic("invalid if condition")
		}
		ifDt, err := p.parseStmt(ifStmt.Body)
		if err != nil {
			panic(err)
		}
		elseDt, err := p.parseStmt(ifStmt.Else)
		if err != nil {
			panic(err)
		}
		if ifDt != elseDt {
			panic("mismatch return type")
		}
		return ifDt, nil
	case *ast.BlockStmt:
		blockStmt := stmt.(*ast.BlockStmt)
		if len(blockStmt.List) != 1 {
			panic("unsupported multistmt block")
		}
		return p.parseStmt(blockStmt.List[0])
	default:
		return NoType, fmt.Errorf("%v", reflect.TypeOf(stmt))
	}
	return NoType, nil
}

func (p *Parser) parseFile(astFile *ast.File) (err error) {
	// Get declarations
	for _, decl := range astFile.Decls {
		n := ast.Node(decl)
		if n == nil {
			err = fmt.Errorf("error on declaration %v", decl)
			return
		}
		p.parseTopLevel(n)
	}
	return
}

func parse(filename string) (modules map[string]*Module, functions map[string]*Function, err error) {
	var p Parser
	p.fileSet = token.NewFileSet()

	var astFile *ast.File
	astFile, err = parser.ParseFile(p.fileSet, filename, nil, parser.AllErrors)
	if err != nil {
		return
	}
	p.modules = make(map[string]*Module)
	p.functions = make(map[string]*Function)
	err = p.parseFile(astFile)
	modules = p.modules
	functions = p.functions
	return
}

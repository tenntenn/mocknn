package mockgen

func eidtFunc(pkg *packages.Package, cmap ast.CommentMap, file *ast.File, decl *ast.FuncDecl) ([]*Edit, error) {
	org, ok := parseComment(cmap, decl)
	if !ok {
		return nil
	}

	orgObj := pkg.Types.Scope().Lookup(org)
	orgdecl := funcDecl(pkg, orgObj)
	if orgdecl == nil {
		return nil
	}

	funcType, _ := orgObj.Type().(*types.Signature)
	mockType, _ := pkg.TypesInfo.TypeOf(decl.Name).(*types.Signature)
	if funcType == nil || mockType == nil {
		return nil
	}

	args := make([]string, 0, funcType.Params().Len())
	for _, f := range orgdecl.Type.Params.List {
		for _, n := range f.Names {
			args = append(args, n.Name)
		}
	}

	callexpr, err := parser.ParseExpr(fmt.Sprintf("%s(%s)", decl.Name.Name, strings.Join(args, ",")))
	if err != nil {
		return fmt.Errorf("replace func: %w", err)
	}

	var stmt ast.Stmt
	if funcType.Results().Len() == 0 {
		stmt = &ast.ExprStmt{X: callexpr}
	} else {
		stmt = &ast.ReturnStmt{
			Results: []ast.Expr{callexpr},
		}
	}

	edits := []*Edit{&Edit{
			For: orgObj.Pkg(),
			Do: func(cur *astutil.Cursor) (bool, error) {
				if cur.Node() != orgspec {
					return true, nil
				}

				cur.Replace(&ast.TypeSpec{
					Doc:  orgspec.Doc,
					Name: orgspec.Name,
					// TODO: Fix for go.dev/issue/46477 in Go 1.20
					TypeParams: nil,
					// +1 is space
					Assign:  orgspec.Name.Pos() + token.Pos(len(orgspec.Name.Name)) + 1,
					Type:    ast.NewIdent(spec.Name.Name),
					Comment: orgspec.Comment,
				})

				return false, nil
			},
		}}

	body := &ast.BlockStmt{
		Lbrace: orgdecl.Body.Lbrace,
		List:   []ast.Stmt{stmt},
		Rbrace: orgdecl.Body.Lbrace + callexpr.End(),
	}

	r.replaces[orgdecl.Body] = body

	return nil
}

func funcDecl(pkg *packages.Package, o types.Object) *ast.FuncDecl {
	switch o.(type) {
	case *types.Func: // ok
	default:
		return nil
	}

	id := ident(pkg, o)
	if id == nil {
		return nil
	}

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			decl, _ := decl.(*ast.FuncDecl)
			if decl != nil && decl.Name == id {
				return decl
			}
		}
	}

	return nil
}

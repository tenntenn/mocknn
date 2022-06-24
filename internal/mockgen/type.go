package mockgen

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func editType(byPkg *packages.Package, cmap ast.CommentMap, file *ast.File, decl *ast.GenDecl) ([]*Edit, error) {
	org, ok := parseComment(cmap, decl)
	if !ok {
		return nil, nil
	}

	var edits []*Edit
	for _, spec := range decl.Specs {
		spec, _ := spec.(*ast.TypeSpec)
		if spec == nil || spec.TypeParams != nil {
			// TODO: Fix for go.dev/issue/46477 in Go 1.20
			continue
		}

		org := org
		if o, ok := parseComment(cmap, spec); ok {
			org = o
		}

		if org == "" {
			continue
		}

		var orgObj types.Object
		if i := strings.Index(org, "."); i >= 0 {
			orgObj = analysisutil.LookupFromImports(byPkg.Types.Imports(), org[:i], org[i+1:])
		} else {
			orgObj = byPkg.Types.Scope().Lookup(org)
		}

		orgspec := typeSpec(byPkg, orgObj)
		if orgspec == nil {
			continue
		}

		typ, _ := orgObj.Type().(*types.Named)
		if typ == nil {
			continue
		}

		edits = append(edits, &Edit{
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
		})

		for i := 0; i < typ.NumMethods(); i++ {
			m := typ.Method(i)
			n := funcDecl(pkg, m)
			if n == nil {
				continue
			}

			edits = append(edits, &Edit{
				For: orgObj.Pkg(),
				Do: func(cur *astutil.Cursor) (bool, error) {
					if cur.Node() != n {
						return true, nil
					}
					cur.Delete()
					return false, nil
				},
			})
		}
	}

	return edits, nil
}

func typeSpec(pkg *packages.Package, o types.Object) *ast.TypeSpec {
	if _, ok := o.(*types.TypeName); !ok {
		return nil
	}

	pkg = findPkg(pkg, o)
	if pkg == nil {
		return nil
	}

	id := findIdent(pkg, o)
	if id == nil {
		return nil
	}

	file := findFile(pkg, id)
	if file == nil {
		return nil
	}

	for _, decl := range file.Decls {
		decl, _ := decl.(*ast.GenDecl)
		if decl == nil || decl.Tok != token.TYPE {
			continue
		}

		for _, spec := range decl.Specs {
			spec, _ := spec.(*ast.TypeSpec)
			if spec != nil && spec.Name == id {
				return spec
			}
		}
	}

	return nil
}

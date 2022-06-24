package mockgen

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func findFile(pkg *packages.Package, n ast.Node) *ast.File {
	for _, file := range pkg.Syntax {
		if file.Pos() <= n.Pos() && n.End() <= file.End() {
			return file
		}
	}
	return nil
}

func findPkg(root *packages.Package, o types.Object) *packages.Package {
	if root.Types == o.Pkg() {
		return root
	}

	for _, pkg := range root.Imports {
		if pkg.Types == o.Pkg() {
			return pkg
		}
	}

	return nil
}

func findIdent(pkg *packages.Package, o types.Object) *ast.Ident {
	for id, obj := range pkg.TypesInfo.Defs {
		if o == obj {
			return id
		}
	}

	return nil
}

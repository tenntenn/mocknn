package overlay

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/gostaticanalysis/analysisutil"
	"go.uber.org/multierr"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

const directive = "//mocknn:"

type Generator struct {
	Dir     string
	Pkgs    []*packages.Package
	Overlay *packages.OverlayJSON
}

func (g *Generator) Generate() (*packages.OverlayJSON, error) {
	r := &replacer{
		pkgs:     g.Pkgs,
		dir:      g.Dir,
		json:     g.Overlay,
		replaces: make(map[ast.Node]ast.Node),
		deletes:  make(map[ast.Node]bool),
	}

	if r.dir == "" {
		tmpdir, err := os.MkdirTemp("", "mocknn-*")
		if err != nil {
			return nil, err
		}
		r.dir = tmpdir
	}

	if r.json == nil {
		r.json = &packages.OverlayJSON{
			Replace: make(map[string]string),
		}
	}

	for _, pkg := range g.Pkgs {
		if err := r.replacePkg(pkg); err != nil {
			return nil, err
		}
	}

	return r.json, nil
}

type replacer struct {
	dir      string
	pkgs     []*packages.Package
	json     *packages.OverlayJSON
	replaces map[ast.Node]ast.Node
	deletes  map[ast.Node]bool
}

func (r *replacer) replacePkg(pkg *packages.Package) error {

	for _, file := range pkg.Syntax {
		if err := r.replaceFile(pkg, file); err != nil {
			return err
		}
	}

	for _, file := range pkg.Syntax {
		if err := r.createFile(pkg.Fset, file); err != nil {
			return err
		}
	}

	for _, pkg := range pkg.Imports {
		for _, file := range pkg.Syntax {
			if err := r.createFile(pkg.Fset, file); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *replacer) replaceFile(pkg *packages.Package, file *ast.File) error {

	cmap := ast.NewCommentMap(pkg.Fset, file, file.Comments)
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.TYPE:
				r.replaceType(pkg, cmap, file, decl)
			case token.VAR, token.CONST:
				r.replaceValue(pkg, cmap, decl)
			}
		case *ast.FuncDecl:
			if err := r.replaceFunc(pkg, cmap, file, decl); err != nil {
				return err
			}
		}

	}

	return nil
}

func (r *replacer) createFile(fset *token.FileSet, file *ast.File) (rerr error) {

	tfile := fset.File(file.Pos())
	if tfile == nil {
		return nil
	}

	var fixed bool
	newfile := astutil.Apply(file, func(c *astutil.Cursor) bool {
		if r.deletes[c.Node()] {
			c.Delete()
			fixed = true
			return false
		}

		if n := r.replaces[c.Node()]; n != nil {
			c.Replace(n)
			fixed = true
			return false
		}

		return true
	}, nil)

	if !fixed {
		return nil
	}

	orgFileName := tfile.Name()
	dstFileName := filepath.Join(r.dir, filepath.Base(orgFileName))
	f, err := os.Create(dstFileName)
	if err != nil {
		return err
	}

	defer func() {
		rerr = multierr.Append(rerr, f.Close())
	}()

	if err := format.Node(f, fset, newfile); err != nil {
		return err
	}

	r.json.Replace[orgFileName] = dstFileName

	return nil
}

func (r *replacer) replaceValue(pkg *packages.Package, cmap ast.CommentMap, decl *ast.GenDecl) {
	org, ok := parseComment(cmap, decl)
	if !ok {
		return
	}

	for _, spec := range decl.Specs {
		spec, _ := spec.(*ast.ValueSpec)
		if spec == nil || len(spec.Names) != 1 {
			continue
		}

		org := org
		if o, ok := parseComment(cmap, spec); ok {
			org = o
		}

		if org == "" {
			continue
		}

		orgspec := valueSepc(pkg, pkg.Types.Scope().Lookup(org))
		if orgspec == nil {
			continue
		}

		r.replaces[orgspec] = &ast.ValueSpec{
			Doc:     orgspec.Doc,
			Names:   []*ast.Ident{ast.NewIdent(org)},
			Type:    orgspec.Type,
			Values:  []ast.Expr{ast.NewIdent(spec.Names[0].Name)},
			Comment: orgspec.Comment,
		}
	}
}

func valueSepc(pkg *packages.Package, o types.Object) *ast.ValueSpec {
	switch o.(type) {
	case *types.Var, *types.Const: // ok
	default:
		return nil
	}

	id := ident(pkg, o)
	if id == nil {
		return nil
	}

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			decl, _ := decl.(*ast.GenDecl)
			if decl == nil || (decl.Tok != token.VAR && decl.Tok != token.CONST) {
				continue
			}

			for _, spec := range decl.Specs {
				spec, _ := spec.(*ast.ValueSpec)
				if spec != nil && len(spec.Names) == 1 && spec.Names[0] == id {
					return spec
				}
			}
		}
	}

	return nil
}

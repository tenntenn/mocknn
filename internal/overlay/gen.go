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
		pkgs: g.Pkgs,
		dir:  g.Dir,
		json: g.Overlay,
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

func (r *replacer) replaceType(pkg *packages.Package, cmap ast.CommentMap, file *ast.File, decl *ast.GenDecl) {
	org, ok := parseComment(cmap, decl)
	if !ok {
		return
	}

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

		orgspec := typeSepc(pkg, pkg.Types.Scope().Lookup(org))
		if orgspec == nil {
			continue
		}

		typ, _ := pkg.TypesInfo.TypeOf(orgspec.Name).(*types.Named)
		if typ == nil {
			continue
		}

		r.replaces[orgspec] = &ast.TypeSpec{
			Doc:  orgspec.Doc,
			Name: orgspec.Name,
			// TODO: Fix for go.dev/issue/46477 in Go 1.20
			TypeParams: nil,
			// +1 is space
			Assign:  orgspec.Name.Pos() + token.Pos(len(orgspec.Name.Name)) + 1,
			Type:    ast.NewIdent(spec.Name.Name),
			Comment: orgspec.Comment,
		}

		for i := 0; i < typ.NumMethods(); i++ {
			m := typ.Method(i)
			n := funcDecl(pkg, m)
			if n != nil {
				r.deletes[n] = true
			}
		}
	}
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

func (r *replacer) replaceFunc(pkg *packages.Package, cmap ast.CommentMap, file *ast.File, decl *ast.FuncDecl) error {
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

	body := &ast.BlockStmt{
		Lbrace: orgdecl.Body.Lbrace,
		List:   []ast.Stmt{stmt},
		Rbrace: orgdecl.Body.Lbrace + callexpr.End(),
	}

	r.replaces[orgdecl.Body] = body

	return nil
}

func typeSepc(pkg *packages.Package, o types.Object) *ast.TypeSpec {
	switch o.(type) {
	case *types.TypeName: // ok
	default:
		return nil
	}

	id := ident(pkg.TypesInfo, o)
	if id == nil {
		return nil
	}

	for _, file := range pkg.Syntax {
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
	}

	return nil
}

func valueSepc(pkg *packages.Package, o types.Object) *ast.ValueSpec {
	switch o.(type) {
	case *types.Var, *types.Const: // ok
	default:
		return nil
	}

	id := ident(pkg.TypesInfo, o)
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

func funcDecl(pkg *packages.Package, o types.Object) *ast.FuncDecl {
	switch o.(type) {
	case *types.Func: // ok
	default:
		return nil
	}

	id := ident(pkg.TypesInfo, o)
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

func ident(info *types.Info, o types.Object) *ast.Ident {
	for id, obj := range info.Defs {
		if o == obj {
			return id
		}
	}
	return nil
}

func parseComment(cmap ast.CommentMap, n ast.Node) (mock string, ok bool) {
	cgs, ok := cmap[n]
	if !ok {
		return "", false
	}

	for _, cg := range cgs {
		for _, c := range cg.List {
			i := strings.Index(c.Text, directive)
			if i < 0 {
				continue
			}
			i += len(directive)
			s := strings.Split(strings.TrimSpace(c.Text[i:]), " ")
			return s[0], true
		}
	}

	return "", false
}

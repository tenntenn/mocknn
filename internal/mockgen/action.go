package mockgen

import (
	"go/ast"
	"go/token"
	"sync"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type Edit struct {
	By  *packages.Package
	For *ast.File
	Do  func(cur *astutil.Cursor) (bool, error)
}

type Action struct {
	Once  sync.Once
	By    *packages.Package
	Edits []*edit
	Err   error
}

func (act *Action) Do() {
	act.once(func() {
		act.Err = act.do()
	})
}

func (act *Action) do() error {
	for _, file := range act.By.Syntax {
		if err := act.doFile(file); err != nil {
			return err
		}
	}

	return nil
}

func (act *Action) doFile(file *ast.File) error {
	cmap := ast.NewCommentMap(act.By.Fset, file, file.Comments)

	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.TYPE:
				if err := editType(act.By, cmap, file, decl); err != nil {
					return err
				}
			case token.VAR, token.CONST:
				if err := editValue(act.By, cmap, decl); err != nil {
					return err
				}
			}
		case *ast.FuncDecl:
			if err := act.editFunc(act.By, cmap, file, decl); err != nil {
				return err
			}
		}

	}

	return nil
}

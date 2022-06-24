package mockgen

import (
	"go/ast"
	"strings"
)

const directive = "//mocknn:"

func parseComment(cmap ast.CommentMap, n ast.Node) (string, bool) {
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

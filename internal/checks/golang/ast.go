package golang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"

	"aint/internal/check"
)

func parse(content []byte) (*token.FileSet, *ast.File, bool) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, 0)
	if err != nil {
		return nil, nil, false
	}
	return fset, file, true
}

func matchAt(fset *token.FileSet, pos token.Pos) check.Match {
	p := fset.Position(pos)
	return check.Match{Line: p.Line, Column: p.Column}
}

func errorDirectComparison(_ string, content []byte) []check.Match {
	fset, file, ok := parse(content)
	if !ok {
		return nil
	}

	var matches []check.Match
	ast.Inspect(file, func(node ast.Node) bool {
		expr, ok := node.(*ast.BinaryExpr)
		if !ok || (expr.Op != token.EQL && expr.Op != token.NEQ) {
			return true
		}
		if isNil(expr.X) || isNil(expr.Y) {
			return true
		}
		if looksLikeError(expr.X) || looksLikeError(expr.Y) {
			matches = append(matches, matchAt(fset, expr.OpPos))
		}
		return true
	})
	return matches
}

func errorStringComparison(_ string, content []byte) []check.Match {
	fset, file, ok := parse(content)
	if !ok {
		return nil
	}

	seen := make(map[token.Pos]bool)
	var matches []check.Match
	add := func(pos token.Pos) {
		if !seen[pos] {
			seen[pos] = true
			matches = append(matches, matchAt(fset, pos))
		}
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch expr := node.(type) {
		case *ast.BinaryExpr:
			if (expr.Op == token.EQL || expr.Op == token.NEQ) &&
				(isErrorStringCall(expr.X) || isErrorStringCall(expr.Y)) {
				add(expr.OpPos)
			}
		case *ast.CallExpr:
			selector, ok := expr.Fun.(*ast.SelectorExpr)
			if !ok || !isIdent(selector.X, "strings") {
				return true
			}
			switch selector.Sel.Name {
			case "Contains", "ContainsAny", "HasPrefix", "HasSuffix", "EqualFold":
				for _, arg := range expr.Args {
					if isErrorStringCall(arg) {
						add(expr.Pos())
						break
					}
				}
			}
		}
		return true
	})
	return matches
}

func panicWithError(_ string, content []byte) []check.Match {
	fset, file, ok := parse(content)
	if !ok {
		return nil
	}

	var matches []check.Match
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isIdent(call.Fun, "panic") || len(call.Args) != 1 {
			return true
		}
		if looksLikeError(call.Args[0]) || isErrorConstructor(call.Args[0]) {
			matches = append(matches, matchAt(fset, call.Pos()))
		}
		return true
	})
	return matches
}

func deferInLoop(_ string, content []byte) []check.Match {
	return statementInLoop(content, func(node ast.Node) (token.Pos, bool) {
		stmt, ok := node.(*ast.DeferStmt)
		if !ok {
			return token.NoPos, false
		}
		return stmt.Defer, true
	})
}

func goroutineInLoop(_ string, content []byte) []check.Match {
	return statementInLoop(content, func(node ast.Node) (token.Pos, bool) {
		stmt, ok := node.(*ast.GoStmt)
		if !ok {
			return token.NoPos, false
		}
		return stmt.Go, true
	})
}

func statementInLoop(content []byte, target func(ast.Node) (token.Pos, bool)) []check.Match {
	fset, file, ok := parse(content)
	if !ok {
		return nil
	}

	seen := make(map[token.Pos]bool)
	var matches []check.Match
	inspectBody := func(body *ast.BlockStmt) {
		ast.Inspect(body, func(node ast.Node) bool {
			if _, nestedFunction := node.(*ast.FuncLit); nestedFunction {
				return false
			}
			if pos, found := target(node); found && !seen[pos] {
				seen[pos] = true
				matches = append(matches, matchAt(fset, pos))
			}
			return true
		})
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.ForStmt:
			inspectBody(stmt.Body)
		case *ast.RangeStmt:
			inspectBody(stmt.Body)
		}
		return true
	})
	return matches
}

func getterPrefix(_ string, content []byte) []check.Match {
	fset, file, ok := parse(content)
	if !ok {
		return nil
	}

	var matches []check.Match
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || !isGetterName(fn.Name.Name) ||
			fieldCount(fn.Type.Params) != 0 || fieldCount(fn.Type.Results) != 1 {
			continue
		}
		matches = append(matches, matchAt(fset, fn.Name.Pos()))
	}
	return matches
}

func largeInterface(_ string, content []byte) []check.Match {
	fset, file, ok := parse(content)
	if !ok {
		return nil
	}

	var matches []check.Match
	ast.Inspect(file, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		iface, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok || interfaceMemberCount(iface) <= 5 {
			return true
		}
		matches = append(matches, matchAt(fset, typeSpec.Name.Pos()))
		return false
	})
	return matches
}

func isNil(expr ast.Expr) bool {
	if paren, ok := expr.(*ast.ParenExpr); ok {
		return isNil(paren.X)
	}
	return isIdent(expr, "nil")
}

func looksLikeError(expr ast.Expr) bool {
	switch expr := expr.(type) {
	case *ast.ParenExpr:
		return looksLikeError(expr.X)
	case *ast.Ident:
		return isErrorName(expr.Name)
	case *ast.SelectorExpr:
		return isErrorName(expr.Sel.Name)
	}
	return false
}

func isErrorName(name string) bool {
	return name == "err" || strings.HasSuffix(name, "Err") || strings.HasSuffix(name, "Error")
}

func isErrorStringCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "Error"
}

func isErrorConstructor(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return (isIdent(selector.X, "fmt") && selector.Sel.Name == "Errorf") ||
		(isIdent(selector.X, "errors") && selector.Sel.Name == "New")
}

func isIdent(expr ast.Expr, name string) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == name
}

func isGetterName(name string) bool {
	if !strings.HasPrefix(name, "Get") || len(name) == len("Get") {
		return false
	}
	r, _ := utf8.DecodeRuneInString(name[len("Get"):])
	return unicode.IsUpper(r)
}

func fieldCount(fields *ast.FieldList) int {
	if fields == nil {
		return 0
	}
	count := 0
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			count++
		} else {
			count += len(field.Names)
		}
	}
	return count
}

func interfaceMemberCount(iface *ast.InterfaceType) int {
	return fieldCount(iface.Methods)
}

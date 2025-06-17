package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type CheckerResult struct {
	FileName string
	Line     int
	Column   int
	Message  string
}

type Checker struct{}

func NewChecker() *Checker {
	return &Checker{}
}

func (c *Checker) Check(directoryName string) ([]*CheckerResult, error) {
	results := []*CheckerResult{}

	err := filepath.Walk(directoryName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		dirResults, err := c.checkDirectory(path)
		if err != nil {
			return fmt.Errorf("ディレクトリ %s のチェック中にエラー: %w", path, err)
		}

		results = append(results, dirResults...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (c *Checker) checkDirectory(directoryName string) ([]*CheckerResult, error) {
	fileSet := token.NewFileSet()

	filter := func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() &&
			!strings.HasPrefix(name, ".") &&
			strings.HasSuffix(name, ".go")
	}

	pkgs, err := parser.ParseDir(fileSet, directoryName, filter, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	results := []*CheckerResult{}

	for _, pkg := range pkgs {
		_results, err := c.checkPackage(pkg, fileSet)
		if err != nil {
			return nil, fmt.Errorf("パッケージ %s のチェック中にエラー: %w", pkg.Name, err)
		}

		results = append(results, _results...)
	}

	return results, nil
}

func (c *Checker) checkPackage(pkg *ast.Package, fileSet *token.FileSet) ([]*CheckerResult, error) {
	testFuncs := map[string]*ast.FuncDecl{}
	publicFuncs := map[string]*ast.FuncDecl{}

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if funcDecl.Name == nil {
				continue
			}

			funcName := funcDecl.Name.Name
			if strings.HasPrefix(funcName, "Test") {
				testFuncs[funcName] = funcDecl
			} else if unicode.IsUpper(rune(funcName[0])) {
				publicFuncs[funcName] = funcDecl
			}
		}
	}

	results := []*CheckerResult{}

	for funcName, funcDecl := range publicFuncs {
		if !c.isTested(funcDecl, testFuncs) {
			result := &CheckerResult{
				FileName: fileSet.Position(funcDecl.Pos()).Filename,
				Line:     fileSet.Position(funcDecl.Pos()).Line,
				Column:   fileSet.Position(funcDecl.Pos()).Column,
				Message:  fmt.Sprintf("関数 %s にテストが実装されていません", funcName),
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func (c *Checker) isTested(funcDecl *ast.FuncDecl, testFuncs map[string]*ast.FuncDecl) bool {
	for _, testFunc := range testFuncs {
		if c.isCalledInRun(testFunc, funcDecl.Name.Name) {
			return true
		}
	}

	return false
}

func (c *Checker) isCalledInRun(testFunc *ast.FuncDecl, targetFuncName string) bool {
	var found bool
	ast.Inspect(testFunc.Body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if c.isRunCall(callExpr) {
			if len(callExpr.Args) >= 2 {
				funcLit, ok := callExpr.Args[1].(*ast.FuncLit)
				if !ok {
					return true
				}

				if c.isFunctionCalledInBlock(funcLit.Body, targetFuncName) {
					found = true
					return false
				}
			}
		}
		return true
	})

	return found
}

func (c *Checker) isRunCall(callExpr *ast.CallExpr) bool {
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return selExpr.Sel.Name == "Run"
}

func (c *Checker) isFunctionCalledInBlock(block *ast.BlockStmt, targetFuncName string) bool {
	var found bool
	ast.Inspect(block, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		ident, ok := callExpr.Fun.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == targetFuncName {
			found = true
			return false
		}
		return true
	})
	return found
}

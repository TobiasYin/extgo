package err_handle

import (
	"fmt"
	"github.com/TobiasYin/extgo/go-parser/ast"
	"github.com/TobiasYin/extgo/go-parser/token"

	"github.com/TobiasYin/extgo"
)

type Plugin struct {
	annotations []string
	initDecl    *ast.FuncDecl
}
var _ extgo.Plugin = &Plugin{}


type FuncDeclVisitor struct {
	funcNode    *ast.FuncDecl
	fileVisitor *Plugin
}

func (v *Plugin) Backprocessor(file *ast.File) {

}


func (v *Plugin) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case *ast.BlockStmt:
		return &BlockVisitor{rnode}
	default:
	}
	return v
}

type BlockVisitor struct {
	blockNode *ast.BlockStmt
}

var _ ast.Visitor = &BlockVisitor{}

func (v *BlockVisitor) Visit(node ast.Node) (w ast.Visitor) {
	fmt.Println("hello")
	switch rnode := node.(type) {
	case *ast.ExprStmt:
		if unary, ok := rnode.X.(*ast.UnaryExpr); ok && unary.Op == token.NOT {
			if unary2, ok := unary.X.(*ast.UnaryExpr); ok && unary2.Op == token.NOT {
				switch ident := unary2.X.(type) {
				case *ast.Ident:
					if ident.Name == "err" {
						for i, s := range v.blockNode.List {
							if s == rnode {
								v.blockNode.List[i] = &ast.IfStmt{
									Cond: &ast.BinaryExpr{
										X:  ident,
										Op: token.NEQ,
										Y:  &ast.Ident{Name: "nil"},
									},
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											&ast.ReturnStmt{Results: []ast.Expr{ident}},
										},
									},
								}
								break
							}
						}
					}
				default:

				}
			}
		}
	default:
	}
	return v
}

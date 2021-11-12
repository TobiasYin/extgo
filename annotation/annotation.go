package annotation

import (
	"github.com/TobiasYin/extgo"
	"fmt"
	"github.com/TobiasYin/extgo/go-parser/ast"
	"github.com/TobiasYin/extgo/go-parser/token"
	"strings"
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

var _ ast.Visitor = &FuncDeclVisitor{}


func (p *Plugin) Backprocessor(f *ast.File) {
	if p.initDecl == nil {
		p.initDecl = &ast.FuncDecl{
			Recv: nil,
			Name: &ast.Ident{
				Name: "init",
			},
			Type: &ast.FuncType{},
			Body: &ast.BlockStmt{},
		}
		f.Decls = append(f.Decls, p.initDecl)
	}

	for i, anno := range p.annotations {
		left := -1
		for i, c := range anno {
			if c == '(' {
				left = i
			}
		}
		funcName := anno
		if funcName == "Annotation" {
			continue
		}
		var paramsExpr []ast.Expr
		if left > 0 {
			funcName = anno[:left]
			param := strings.Trim(anno[left:], "()")
			params := strings.Split(param, ",")

			for _, p := range params {
				paramsExpr = append(paramsExpr, &ast.Ident{Name: strings.Trim(p, " ")})
			}
		}

		p.initDecl.Body.List = append(p.initDecl.Body.List, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: getFuncName(i),
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.Ident{
						Name: funcName,
					},
					Args: paramsExpr,
				},
			},
		})

		f.Decls = append(f.Decls, &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: getFuncName(i),
						},
					},
					Type: &ast.Ident{
						Name: "func()",
					},
					Values: nil,
				},
			},
		})

	}
}

func (p *Plugin) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case nil:
	case *ast.FuncDecl:
		if rnode.Name.Name == "init" {
			p.initDecl = rnode
			return nil
		}
		return &FuncDeclVisitor{rnode, p}

	default:
	}
	return p
}

func (v *FuncDeclVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case *ast.CommentGroup:
		//fmt.Println(*rnode.List[0])
		var ids []int
		for _, c := range rnode.List {
			comment := strings.Trim(c.Text, " /")
			if comment[0] == '@' {
				v.fileVisitor.annotations = append(v.fileVisitor.annotations, comment[1:])
				ids = append(ids, len(v.fileVisitor.annotations)-1)
			}
		}
		var statements []ast.Stmt
		for _, id := range ids {
			//fmt.Println(id)
			statement := ast.CallExpr{
				Fun: &ast.Ident{
					Name: getFuncName(id),
				},
			}
			statements = append(statements, &ast.ExprStmt{X: &statement})
		}
		for _, s := range v.funcNode.Body.List {
			statements = append(statements, s)
		}
		v.funcNode.Body.List = statements
		return nil
	default:
	}
	return v
}

func getFuncName(id int) string {
	return fmt.Sprintf("AnnotationGenerateFunc_%d", id)
}

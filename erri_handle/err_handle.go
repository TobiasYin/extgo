package erri_handle

import (
	"fmt"
	"github.com/TobiasYin/extgo/go-parser/ast"
	"github.com/TobiasYin/extgo/go-parser/token"
	"strings"
	"sync/atomic"

	"github.com/TobiasYin/extgo"
)

type File struct {
	file         *ast.File
	addImport    bool
	useErri      bool
	shouldHandle map[*ast.FuncDecl]struct{}
}

type Plugin struct {
	file    *File
	Fset    *token.FileSet
	UseErri bool
}

var _ extgo.Plugin = &Plugin{}

type FuncDeclVisitor struct {
	funcNode    *ast.FuncDecl
	fileVisitor *Plugin
	Fset        *token.FileSet
	File        *File
}

func (v *Plugin) Backprocessor(file *ast.File) {
	if v.UseErri && v.file.addImport {
		has := false
		for _, spec := range file.Imports {
			if spec.Name != nil && spec.Name.Name == "erri" {
				has = true
				break
			}
			path := strings.Trim(spec.Path.Value, "\"")
			if strings.HasSuffix(path, "/erri") {
				has = true
				break
			}
		}
		if !has {
			spec := &ast.ImportSpec{
				Path: &ast.BasicLit{Value: fmt.Sprintf("\"github.com/TobiasYin/extgo/erri\"")},
			}
			file.Imports = append(file.Imports, spec)
			newDecls := []ast.Decl{}
			newDecls = append(newDecls, &ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{spec}})
			file.Decls = append(newDecls, file.Decls...)
		}
	}

	for f, _ := range v.file.shouldHandle {
		handleFunc(f)
	}
}
func handleFunc(f *ast.FuncDecl) {
	if f.Type.Results == nil || f.Type.Results.List == nil {
		return
	}
	if f.Body == nil || f.Body.List == nil {
		return
	}
	last := f.Type.Results.List[len(f.Type.Results.List)-1]
	if idt, ok := last.Type.(*ast.Ident); !ok || idt.Name != "error" {
		return
	}
	result := []ast.Stmt{}
	i := 0
	for _, r := range f.Type.Results.List {
		if len(r.Names) == 0 {
			result = append(result, &ast.DeclStmt{Decl: &ast.GenDecl{
				Tok:   token.VAR,
				Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{{Name: fmt.Sprintf("ret_nil_gen_%d", i)}}, Type: r.Type}},
			}})
			i += 1
		}
		for _, _ = range r.Names {
			result = append(result, &ast.DeclStmt{Decl: &ast.GenDecl{
				Tok:   token.VAR,
				Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{{Name: fmt.Sprintf("ret_nil_gen_%d", i)}}, Type: r.Type}},
			}})
			i += 1
		}
	}
	result = result[:len(result)-1]
	f.Body.List = insert(f.Body.List, 0, result)
}

func (v *Plugin) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case *ast.File:
		v.file = &File{
			file:         rnode,
			addImport:    false,
			useErri:      v.UseErri,
			shouldHandle: map[*ast.FuncDecl]struct{}{},
		}
	case *ast.FuncDecl:
		return &FuncDeclVisitor{rnode, v, v.Fset, v.file}
	default:
	}
	return v
}

func (v *FuncDeclVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case *ast.BlockStmt:
		b := &BlockVisitor{blockNode: rnode, funcVisitor: v}
		b.Walk()
		return nil
	default:
	}
	return v
}

type BlockVisitor struct {
	blockNode   *ast.BlockStmt
	funcVisitor *FuncDeclVisitor
	errNum      int64
	tmpNum      int64
	curStmt     int
}

func (v *BlockVisitor) Walk() {
	for i := 0; i < len(v.blockNode.List); i++ {
		v.curStmt = i
		ast.Walk(v, v.blockNode.List[i])
		i = v.curStmt
	}
}

var _ ast.Visitor = &BlockVisitor{}

func (v *BlockVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case *ast.AssignStmt:
		return &AssignStmtVisitor{rnode, v}
	case *ast.DeclStmt:
		return &DeclStmtVisitor{rnode, v}
	case *ast.ExprStmt:
		return &ExprStmtVisitor{rnode, v}
	case *ast.FuncDecl:
		return &FuncDeclVisitor{rnode, nil, v.funcVisitor.Fset, v.funcVisitor.File} // TODO
	case *ast.BlockStmt:
		b := &BlockVisitor{blockNode: rnode, funcVisitor: v.funcVisitor}
		b.Walk()
		return nil
	case *ast.ReturnStmt:
		v.handleReturnStmt(rnode)
	case *ast.QuesExpr:
		h := QuesExprHandler{
			BlockVisitor: v,
			quesNode:     rnode,
		}
		h.Handle()
	case *ast.CallExpr:
		v.handleCallExpr(rnode)
	default:
	}
	return v
}

type DeclStmtVisitor struct {
	decl         *ast.DeclStmt
	BlockVisitor *BlockVisitor
}

func (v *DeclStmtVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case *ast.QuesExpr:
		h := QuesExprHandler{
			DeclStmtVisitor: v,
			BlockVisitor:    v.BlockVisitor,
			quesNode:        rnode,
		}
		h.Handle()
		return nil
	case *ast.SelectorExpr:
		return &SelectorExprVisitor{
			selector:        rnode,
			DeclStmtVisitor: v,
			BlockVisitor:    v.BlockVisitor,
		}

	default:
	}
	return v
}

type AssignStmtVisitor struct {
	assignNode   *ast.AssignStmt
	BlockVisitor *BlockVisitor
}

func (v *AssignStmtVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case *ast.QuesExpr:
		h := QuesExprHandler{
			AssignStmtVisitor: v,
			BlockVisitor:      v.BlockVisitor,
			quesNode:          rnode,
		}
		h.Handle()
		return nil
	case *ast.CallExpr:
		v.BlockVisitor.handleCallExpr(rnode)
	case *ast.SelectorExpr:
		return &SelectorExprVisitor{
			selector:          rnode,
			AssignStmtVisitor: v,
			BlockVisitor:      v.BlockVisitor,
		}

	default:
	}
	return v
}

type ExprStmtVisitor struct {
	exprStmt     *ast.ExprStmt
	BlockVisitor *BlockVisitor
}

func (v *ExprStmtVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case *ast.QuesExpr:
		h := QuesExprHandler{
			ExprStmtVisitor: v,
			BlockVisitor:    v.BlockVisitor,
			quesNode:        rnode,
		}
		h.Handle()
		return nil
	case *ast.SelectorExpr:
		return &SelectorExprVisitor{
			selector:          rnode,
			AssignStmtVisitor: nil,
			ExprStmtVisitor:   v,
			BlockVisitor:      v.BlockVisitor,
		}
	case *ast.CallExpr:
		v.BlockVisitor.handleCallExpr(rnode)
	default:
	}
	return v
}

type SelectorExprVisitor struct {
	selector          *ast.SelectorExpr
	AssignStmtVisitor *AssignStmtVisitor
	DeclStmtVisitor   *DeclStmtVisitor
	ExprStmtVisitor   *ExprStmtVisitor
	BlockVisitor      *BlockVisitor
}

func (v *SelectorExprVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch rnode := node.(type) {
	case *ast.QuesExpr:
		h := QuesExprHandler{
			SelectorExprVisitor: v,
			ExprStmtVisitor:     v.ExprStmtVisitor,
			AssignStmtVisitor:   v.AssignStmtVisitor,
			DeclStmtVisitor:     v.DeclStmtVisitor,
			BlockVisitor:        v.BlockVisitor,
			quesNode:            rnode,
		}
		h.Handle()
		return nil
	case *ast.SelectorExpr:
		return &SelectorExprVisitor{
			selector:          rnode,
			AssignStmtVisitor: v.AssignStmtVisitor,
			ExprStmtVisitor:   v.ExprStmtVisitor,
			DeclStmtVisitor:   v.DeclStmtVisitor,
			BlockVisitor:      v.BlockVisitor,
		}
	default:
	}
	return v
}

type QuesExprHandler struct {
	SelectorExprVisitor *SelectorExprVisitor
	AssignStmtVisitor   *AssignStmtVisitor
	DeclStmtVisitor     *DeclStmtVisitor
	ExprStmtVisitor     *ExprStmtVisitor
	BlockVisitor        *BlockVisitor
	quesNode            *ast.QuesExpr
}

func (q *QuesExprHandler) isLeft() bool {
	if q.SelectorExprVisitor == nil {
		return false
	}
	if l, ok := q.SelectorExprVisitor.selector.X.(*ast.QuesExpr); ok {
		return l == q.quesNode
	}
	return false
}

func remove(s []ast.Stmt, r int) []ast.Stmt {
	newStmts := make([]ast.Stmt, 0, len(s)-1)
	for i, stmt := range s {
		if i != r {
			newStmts = append(newStmts, stmt)
		}
	}
	return newStmts
}

func insert(s []ast.Stmt, start int, s1 []ast.Stmt) []ast.Stmt {
	newStmts := make([]ast.Stmt, 0, len(s)+len(s1))
	if start == len(s) {
		newStmts = append(newStmts, s...)
		newStmts = append(newStmts, s1...)
	} else {
		for i, stmt := range s {
			if i == start {
				newStmts = append(newStmts, s1...)
			}
			newStmts = append(newStmts, stmt)
		}
	}
	return newStmts
}

func (q *QuesExprHandler) Handle() {
	errName, errHandle := q.buildErrRet()
	stmts := []ast.Stmt{}

	if q.isLeft() {
		names := q.BlockVisitor.genTempAssignName(1)
		assignStmt := buildAssign(names, errName, q.quesNode.X, token.DEFINE)
		q.SelectorExprVisitor.selector.X = &ast.Ident{Name: names[0]}
		stmts = append(stmts, assignStmt)
		stmts = append(stmts, errHandle)
		// stmts add cur stmt
		q.BlockVisitor.blockNode.List = insert(q.BlockVisitor.blockNode.List, q.BlockVisitor.curStmt, stmts)
		q.BlockVisitor.curStmt -= 1
		return
	}

	if q.ExprStmtVisitor != nil {
		/*
			A()?
			Expand To
			err := A()
			if err != nil {
				return err
			}
		*/
		assignStmt := buildAssign(nil, errName, q.quesNode.X, token.DEFINE)
		stmts = append(stmts, assignStmt, errHandle)
		q.BlockVisitor.blockNode.List = remove(q.BlockVisitor.blockNode.List, q.BlockVisitor.curStmt)
		q.BlockVisitor.blockNode.List = insert(q.BlockVisitor.blockNode.List, q.BlockVisitor.curStmt, stmts)
		q.BlockVisitor.curStmt -= 1
		return
	}
	if q.DeclStmtVisitor != nil {
		if d, ok := q.DeclStmtVisitor.decl.Decl.(*ast.GenDecl); !ok {
			return
		} else {
			if d.Tok != token.VAR {
				return
			}
			for _, s := range d.Specs {
				if p, ok := s.(*ast.ValueSpec); ok {
					if p.Type == nil {
						lhs := []ast.Expr{}
						for _, name := range p.Names {
							lhs = append(lhs, name)
						}

						rhs := []ast.Expr{}
						for _, value := range p.Values {
							rhs = append(rhs, value)
						}
						stmts = append(stmts, &ast.AssignStmt{
							Tok: token.DEFINE,
							Lhs: lhs,
							Rhs: rhs,
						})
						continue
					}
					stmts = append(stmts,
						&ast.DeclStmt{Decl: &ast.GenDecl{
							Tok: token.VAR,
							Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: p.Names,
									Type:  p.Type,
								},
							},
						}})

					lhs := []ast.Expr{}
					for _, name := range p.Names {
						lhs = append(lhs, name)
					}

					rhs := []ast.Expr{}
					for _, value := range p.Values {
						rhs = append(rhs, value)
					}
					stmts = append(stmts, &ast.AssignStmt{
						Tok: token.ASSIGN,
						Lhs: lhs,
						Rhs: rhs,
					})
				}
			}
			q.BlockVisitor.blockNode.List = remove(q.BlockVisitor.blockNode.List, q.BlockVisitor.curStmt)
			q.BlockVisitor.blockNode.List = insert(q.BlockVisitor.blockNode.List, q.BlockVisitor.curStmt, stmts)
			q.BlockVisitor.curStmt -= 1

		}
		return
	}
	if q.AssignStmtVisitor != nil {
		if q.AssignStmtVisitor.assignNode.Tok == token.ASSIGN {
			q.BlockVisitor.blockNode.List = insert(q.BlockVisitor.blockNode.List, q.BlockVisitor.curStmt, []ast.Stmt{
				&ast.DeclStmt{Decl: &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{&ast.ValueSpec{
						Names: []*ast.Ident{{Name: errName}},
						Type:  &ast.Ident{Name: "error"},
					}}},
				},
			})
			q.BlockVisitor.curStmt += 1
		}
		q.AssignStmtVisitor.assignNode.Lhs = append(q.AssignStmtVisitor.assignNode.Lhs, &ast.Ident{Name: errName})
		q.AssignStmtVisitor.assignNode.Rhs[0] = q.quesNode.X
		stmts = append(stmts, errHandle)
		q.BlockVisitor.blockNode.List = insert(q.BlockVisitor.blockNode.List, q.BlockVisitor.curStmt+1, stmts)
		q.BlockVisitor.curStmt -= 1
	}
}

func buildAssign(assignNames []string, errName string, l ast.Expr, tok token.Token) *ast.AssignStmt {
	assignExpr := []ast.Expr{}
	for _, name := range assignNames {
		assignExpr = append(assignExpr, &ast.Ident{Name: name})
	}
	if errName != "" {
		assignExpr = append(assignExpr, &ast.Ident{Name: errName})
	}
	return &ast.AssignStmt{
		Lhs: assignExpr,
		Tok: tok,
		Rhs: []ast.Expr{l},
	}
}

func (q *BlockVisitor) genTempAssignName(size int) []string {
	names := []string{}
	for i := 0; i < size; i++ {
		tmpNum := atomic.AddInt64(&q.tmpNum, 1)
		tmp := fmt.Sprintf("tmp_gen_%d", tmpNum)
		names = append(names, tmp)
	}
	return names
}

func (q *QuesExprHandler) buildErrRet() (string, ast.Stmt) {
	errNum := atomic.AddInt64(&q.BlockVisitor.errNum, 1)
	err := fmt.Sprintf("err_gen_%d", errNum)

	fun := q.BlockVisitor.funcVisitor

	usePanic := true
	pos := q.BlockVisitor.funcVisitor.Fset.Position(q.quesNode.Pos())
	line := pos.String()

	var state ast.Stmt
	res := fun.funcNode.Type.Results
	if res != nil {
		if len(res.List) != 0 {
			last := res.List[len(res.List)-1]
			typ := last.Type
			if rTyp, ok := typ.(*ast.Ident); ok {
				if rTyp.Name == "error" {
					usePanic = false
				}
			}
		}
	}

	var errRet ast.Expr
	if q.BlockVisitor.funcVisitor.File.useErri {
		q.BlockVisitor.funcVisitor.File.addImport = true
		errRet = &ast.CallExpr{
			Fun: &ast.Ident{Name: "erri.ErrorWithLine"},
			Args: []ast.Expr{
				&ast.BasicLit{Value: fmt.Sprintf("\"%s\"", line)},
				&ast.Ident{Name: err},
			},
		}
	} else {
		errRet = &ast.Ident{
			Name: err,
		}
	}

	if usePanic {
		state = &ast.ExprStmt{X: &ast.CallExpr{
			Fun:  &ast.Ident{Name: "panic"},
			Args: []ast.Expr{errRet},
		}}
	} else {
		ret := 0
		for _, i := range res.List {
			ret += len(i.Names)
			if len(i.Names) == 0 {
				ret += 1
			}
		}
		result := []ast.Expr{}
		for i := 0; i < ret-1; i++ {
			result = append(result, &ast.Ident{
				Name: fmt.Sprintf("ret_nil_gen_%d", i),
			})
		}
		if len(result) > 0 {
			f := q.BlockVisitor.funcVisitor.funcNode
			q.BlockVisitor.funcVisitor.File.shouldHandle[f] = struct{}{}
		}

		result = append(result, errRet)
		state = &ast.ReturnStmt{
			Results: result,
		}
	}

	return err, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X: &ast.Ident{
				Name: err,
			},
			Op: token.NEQ,
			Y: &ast.Ident{
				Name: "nil",
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				state,
			},
		},
	}
}

func (v *BlockVisitor) handleReturnStmt(s *ast.ReturnStmt) {
	if !v.funcVisitor.File.useErri {
		return
	}
	f := v.funcVisitor.funcNode
	res := f.Type.Results
	if res != nil {
		if len(res.List) != 0 {
			last := res.List[len(res.List)-1]
			typ := last.Type
			if rTyp, ok := typ.(*ast.Ident); ok {
				if rTyp.Name != "error" {
					return
				}
			}
		}
	}

	retCount := 0
	names := []ast.Expr{}
	for _, r := range f.Type.Results.List {
		if len(r.Names) == 0 {
			retCount += 1
		}
		for _, n := range r.Names {
			retCount += 1
			names = append(names, n)
		}
	}

	if len(s.Results) == 0 {
		if retCount == 0 {
			return
		}

		s.Results = names
	}

	// TODO handle direct Return Funcall（）、board Return
	if len(s.Results) == 1 && retCount != 1 {
		names := v.genTempAssignName(retCount)
		assign := buildAssign(names, "", s.Results[0], token.DEFINE)
		v.blockNode.List = insert(v.blockNode.List, v.curStmt, []ast.Stmt{assign})
		v.curStmt += 1
		res := []ast.Expr{}
		for _, n := range names {
			res = append(res, &ast.Ident{Name: n})
		}
		s.Results = res
	}

	last := s.Results[len(s.Results)-1]
	if call, ok := last.(*ast.CallExpr); ok {
		if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
			if x, ok := fun.X.(*ast.Ident); ok {
				if x.Name == "erri" {
					return
				}
			}
		}
		if fun, ok := call.Fun.(*ast.Ident); ok {
			if strings.HasPrefix(fun.Name, "erri.") {
				return
			}
		}

	}
	if last, ok := last.(*ast.Ident); ok {
		if last.Name == "nil" {
			return
		}
	}
	errRet := &ast.CallExpr{
		Fun: &ast.Ident{Name: "erri.ErrorWithLine"},
		Args: []ast.Expr{
			&ast.BasicLit{Value: fmt.Sprintf("\"%s\"", v.funcVisitor.Fset.Position(s.Return))},
			last,
		},
	}
	s.Results[len(s.Results)-1] = errRet

}

func (v *BlockVisitor) handleCallExpr(s *ast.CallExpr) {
	if !v.funcVisitor.File.useErri {
		return
	}
	if fun, ok := s.Fun.(*ast.Ident); ok {
		if strings.HasPrefix(fun.Name, "erri.") {
			return
		}
	}
	if fun, ok := s.Fun.(*ast.SelectorExpr); ok {
		if x, ok := fun.X.(*ast.Ident); ok {
			if x.Name != "erri" {
				return
			}
			name := fun.Sel.Name
			if strings.HasSuffix(name, "WithLine") {
				return
			}
			fun.Sel.Name = name + "WithLine"
			newArgs := []ast.Expr{&ast.BasicLit{Value: fmt.Sprintf("\"%s\"", v.funcVisitor.Fset.Position(s.Pos()))}}
			newArgs = append(newArgs, s.Args...)
			s.Args = newArgs
		}
	}

}

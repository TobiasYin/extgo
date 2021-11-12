package extgo

import (
	"github.com/TobiasYin/extgo/go-parser/ast"
)

type Plugin interface {
	ast.Visitor
	Backprocessor(file *ast.File)
}

var _ ast.Visitor = &Visitor{}

type Visitor struct {
	plugin  []Plugin
	Visitor []ast.Visitor
	start   bool
	parent  *Visitor
}

func (v *Visitor) Register(visitor Plugin) {
	v.plugin = append(v.plugin, visitor)
}

func (v *Visitor) Visit(node ast.Node) (w ast.Visitor) {
	if !v.start {
		for _, m := range v.plugin {
			v.Visitor = append(v.Visitor, m)
		}
		v.start = true
	}
	var newMiddlewares []ast.Visitor
	for _, m := range v.Visitor {
		r := m.Visit(node)
		if r != nil {
			newMiddlewares = append(newMiddlewares, r)
		}
	}
	if len(newMiddlewares) == 0 {
		return nil
	}
	newVisitor := Visitor{
		plugin:  v.plugin,
		Visitor: newMiddlewares,
		start:   true,
		parent:  v,
	}
	return &newVisitor
}

func (v *Visitor) Backprocessor(f *ast.File) {
	for _, m := range v.plugin {
		m.Backprocessor(f)
	}
}

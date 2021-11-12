package main

import (
	"bytes"
	"github.com/TobiasYin/extgo/go-parser/ast"
	"github.com/TobiasYin/extgo/go-parser/parser"
	"github.com/TobiasYin/extgo/go-parser/printer"
	"github.com/TobiasYin/extgo/go-parser/token"
	"os"
	"path"
	"strings"

	"github.com/TobiasYin/extgo"
	"github.com/TobiasYin/extgo/erri_handle"
)

func main() {
	v := extgo.Visitor{}
	//v.Register(&annotation.Plugin{})
	v.Register(&erri_handle.Plugin{})

	parseFile("../example/main2.go")

}

func getFullName(filename string) string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return path.Join(wd, filename)
}

func parseFile(filename string) {
	sources, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, getFullName(filename), sources, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	//ast.Print(fset, f)
	v := extgo.Visitor{}
	v.Register(&erri_handle.Plugin{Fset: fset, UseErri: true})

	ast.Walk(&v, f)
	v.Backprocessor(f)



	buf := bytes.Buffer{}
	err = printer.Fprint(&buf, fset, f)
	if err != nil {
		panic(err)
	}
	sources = buf.Bytes()
	f, err = parser.ParseFile(fset, "", sources, 0)
	if err != nil {
		panic(err)
	}
	outputFile := getFileName(filename)
	_ = os.MkdirAll(outputFile[:strings.LastIndex(outputFile, "/")], 0755)
	out, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	err = printer.Fprint(out, fset, f)
	//err = printer.Fprint(os.Stdout, fset, f)
	if err != nil {
		panic(err)
	}
}

func getFileName(f string) string {
	return "output/src/" + f
}
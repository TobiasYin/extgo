package main

import (
	"fmt"

	"github.com/TobiasYin/extgo/erri"
)

func A() error {
	return erri.Errorf("hello")
}

func B() (int, error) {
	return 1, nil
}

type AA struct{}

func (a2 AA) AAA() (int, error) {
	return 0, erri.Errorf("hello")
}

func NewAA() (AA, error) {
	return AA{}, nil
}

func f() (int, error) {
	A()?
	a := B()?
	fmt.Println(a)
	b := NewAA()?.AAA()?
	fmt.Println(b)
	c := NewAA()?
	fmt.Println(c)
	return 1, nil
}

func ff() error {
	return nil
}

func f4() error {
	return ff()
}

func ff2() (int, error) {
	return 0, nil
}

func f5() (int, error) {
	return ff2()
}

func f6() (int, error) {
	return 0, erri.Errorf("a", "b")
}

func f7() (i int, err error) {
	return
}

func main() {
	var a int = f()?
	var b = f()?
	fmt.Println(a)
	fmt.Println(b)
	A()?
}

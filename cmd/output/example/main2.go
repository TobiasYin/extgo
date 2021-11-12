package main

import (
	"fmt"

	"github.com/TobiasYin/extgo/erri"
)

func A() error {
	return erri.ErrorfWithLine("/Users/tobias/projects/extgo/example/main2.go:10:9", "hello")
}

func B() (int, error) {
	return 1, nil
}

type AA struct{}

func (a2 AA) AAA() (int, error) {
	return 0, erri.ErrorfWithLine("/Users/tobias/projects/extgo/example/main2.go:20:12", "hello")
}

func NewAA() (AA, error) {
	return AA{}, nil
}

func f() (int, error) {
	var ret_nil_gen_0 int
	A()
	a, err_gen_1 := B()
	if err_gen_1 != nil {
		return ret_nil_gen_0, erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:29:7", err_gen_1)
	}
	fmt.Println(a)
	tmp_gen_1, err_gen_3 := NewAA()
	if err_gen_3 != nil {
		return ret_nil_gen_0, erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:31:7", err_gen_3)
	}
	b, err_gen_2 := tmp_gen_1.AAA()
	if err_gen_2 != nil {
		return ret_nil_gen_0, erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:31:7", err_gen_2)
	}
	fmt.Println(b)
	c, err_gen_4 := NewAA()
	if err_gen_4 != nil {
		return ret_nil_gen_0, erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:33:7", err_gen_4)
	}
	fmt.Println(c)
	return 1, nil
}

func ff() error {
	return nil
}

func f4() error {
	return erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:43:2", ff())
}

func ff2() (int, error) {
	return 0, nil
}

func f5() (int, error) {
	tmp_gen_1, tmp_gen_2 := ff2()
	return tmp_gen_1, erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:51:2", tmp_gen_2)
}

func f6() (int, error) {
	return 0, erri.ErrorfWithLine("/Users/tobias/projects/extgo/example/main2.go:55:12", "a", "b")
}

func f7() (i int, err error) {
	return i, erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:59:2", err)

}

func main() {
	var a int
	var err_gen_2 error
	a, err_gen_2 = f()
	if err_gen_2 != nil {
		panic(erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:63:14", err_gen_2))
	}
	b, err_gen_4 := f()
	if err_gen_4 != nil {
		panic(erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:64:10", err_gen_4))
	}
	fmt.Println(a)
	fmt.Println(b)
	err_gen_5 := A()
	if err_gen_5 != nil {
		panic(erri.ErrorWithLine("/Users/tobias/projects/extgo/example/main2.go:67:2", err_gen_5))
	}
}

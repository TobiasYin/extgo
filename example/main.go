package main

import (
	"errors"
	"fmt"
)

var a int = 1

func init()  {
	fmt.Println("init")
}

func BoldWrapper() func() {
	fmt.Println("use bold")
	return func() {
		fmt.Println("bold wrapper")
	}
}

func Wrapper(i int) func(){
	fmt.Println("register", i)
	return func (){
		fmt.Println("exec", i)
	}
}

func Wrapper2(i int, j int) func(){
	fmt.Println("register wrapper2", i)
	return func (){
		fmt.Println("exec wrapper2", i, j)
	}
}

// @Wrapper(1)
func f1(){
	fmt.Println("f1")
}

func f0() error{
	var err error
	if err != nil {
		return err
	}
	return errors.New("a")
}

// @BoldWrapper
// @Wrapper2(2, 3)
func f2() error {
	err := f0()
	!!err
	fmt.Println("f2")
	f1()
	fmt.Println("f2 end")
	return nil
}

func main() {
	fmt.Println("hello world")
	f1()
	f2()
}


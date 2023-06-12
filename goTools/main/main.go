package main

import (
	"fmt"
	"runtime"
)

func Recover() {
	if err := recover(); err != nil {
		buf := make([]byte, 102400)
		runtime.Stack(buf, false)
		fmt.Printf("[Recovery] panic1\n")
		//fmt.Printf("[Recovery] panic1 recovered: %s\n%s", err, buf)
	}
}

type testA struct {
	a int
}

func main() {
	a := &testA{}
	a.a = 1
	test1(&a.a)
	fmt.Println("after", a)
}

func test1(a *int) {
	a, err := test2()
	fmt.Println(*a, err)
}

func test2() (*int, error) {
	b := 2
	return &b, nil
}

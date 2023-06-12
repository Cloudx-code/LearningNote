package main

import (
	"encoding/json"
	"fmt"
)

type test struct {
	A int `json:"a,omitempty"`
	b int `json:"b,omitempty"`
}

func main() {
	test1 := &test{}

	aStr := `{"a":1}`
	err := json.Unmarshal([]byte(aStr), test1)
	if err != nil {
		fmt.Println("fail to unmarshal,err:", err)
		return
	}
	fmt.Println("success to unmarshal,test:", test1)
	bStr := `{"b":2}`
	err = json.Unmarshal([]byte(bStr), test1)
	if err != nil {
		fmt.Println("fail to unmarshal,err:", err)
		return
	}
	fmt.Println("success to unmarshal,test:", test1)
}

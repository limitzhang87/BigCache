package main

import (
	"fmt"
	"github.com/limitzhang87/BigCache"
	"log"
)

type Foo struct {
	Name string
	Age  uint16
}

func main() {
	f := Foo{
		Name: "zhangjixian",
		Age:  26,
	}

	cache := BigCache.NewInstance()

	key := "aa"
	err := cache.Set(key, f)

	if err != nil {
		log.Fatal("set value err : ", err)
	}
	fmt.Printf("set success value(%v)\n", f)

	tt, err := cache.Get(key)
	if err != nil {
		log.Fatal("get value err : ", err)
	}

	newValue, ok := tt.(Foo)
	if !ok {
		log.Fatal("get value err : ", err)
	}

	fmt.Printf("get value success value(%v)", newValue)
}

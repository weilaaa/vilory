package main

import (
	"fmt"
	"time"
)

// implement any process you want
func main() {
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Second)
		fmt.Println(i)
	}
}

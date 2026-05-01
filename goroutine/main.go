package main

import (
	"fmt"
	"time"
)

func sayHello(name string) {
	for i := 0; i < 5; i++ {
		fmt.Printf("Hello %s\n", name)
	}
}

func main() {
	// Goroutine
	// go sayHello("Nguyen")

	// Using goroutine with anonymous function
	go func() {
		for i := 0; i < 5; i++ {
			fmt.Println(1)
		}
	}()

	// Non goroutine
	// sayHello("Hao")

	time.Sleep(time.Second)
}

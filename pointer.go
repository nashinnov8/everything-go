package main

import "fmt"

func pointerDemo() {
	var x int = 10
	var p *int = &x               // p is a pointer to an integer, and it holds the address of x
	fmt.Println("Value of x:", x) // Output: Value of x: 10
	fmt.Println("Address of x:", &x)
	fmt.Println("Pointer p points to:", p)
	fmt.Println("Pointer value: ", *p)
}

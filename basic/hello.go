package main

import "fmt"

var name string = "Nguyen Van A"

func main() {
	// name := "Nguyen Van B"
	// fmt.Println("Hello, World!")
	// fmt.Println("My name is", Hello(name))

	// a, b := 10, 20
	// fmt.Printf("The sum of %d and %d is %d\n", a, b, add(a, b))

	// x, y := 10.0, 0.0
	// result, err := Divine(x, y)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	// fmt.Printf("The result of dividing %.2f by %.2f is %.2f\n", x, y, result)
	pointerDemo()
}

func Hello(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// This is a simple function to add two integers
// You can also declare a function like (a, b int) instead of (a int, b int)
func add(a int, b int) int {
	return a + b
}

func Divine(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}

	return a / b, nil
}

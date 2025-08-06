package main

import (
	"fmt"
	"os"
)

const version = "0.0.1-alpha"

func main() {
	fmt.Printf("pyqol v%s - Python Quality of Life\n", version)
	fmt.Println("Coming soon: Advanced Python static analysis with CFG and APTED")
	os.Exit(0)
}
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pyqol/pyqol/internal/version"
)

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (short)")
	flag.Parse()

	if showVersion {
		fmt.Println(version.Info())
		os.Exit(0)
	}

	fmt.Printf("pyqol %s - Python Quality of Life\n", version.Short())
	fmt.Println("Coming soon: Advanced Python static analysis with CFG and APTED")
}

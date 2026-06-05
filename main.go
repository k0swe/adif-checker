package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <adif-file>\n", os.Args[0])
		os.Exit(2)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read file: %v\n", err)
		os.Exit(1)
	}

	warnings, err := validateADIFWithWarnings(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid ADIF: %v\n", err)
		os.Exit(1)
	}
	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", warning)
	}

	fmt.Println("ADIF is well-formed")
}

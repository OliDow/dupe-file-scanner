package main

import (
	"dupe-file-checker/pkg/reporter"
	"dupe-file-checker/pkg/scanner"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dupe-checker <directory>")
		os.Exit(1)
	}

	root := os.Args[1]

	s := scanner.New()
	duplicates, err := s.Scan(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	reporter.PrintDuplicates(duplicates)
}

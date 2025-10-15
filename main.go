package main

import (
	"dupe-file-checker/pkg/reporter"
	"dupe-file-checker/pkg/scanner"
	"flag"
	"fmt"
	"os"
)

func main() {
	onlyImages := flag.Bool("only-images", false, "Only check image files (jpg, jpeg, png, gif, heic, heif, webp, bmp)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: dupe-checker [--only-images] <directory>")
		os.Exit(1)
	}

	root := flag.Arg(0)

	s := scanner.New()
	duplicates, err := s.Scan(root, *onlyImages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	reporter.PrintDuplicates(duplicates)
}

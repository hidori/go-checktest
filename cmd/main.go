package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hidori/go-checktest/checker"
)

func main() {
	err := run(os.Args)
	if err != nil {
		log.Fatalf("An error occurred: %v", err)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		fmt.Println("usage: checktest <directory>")
		return nil
	}

	checker := checker.NewChecker()

	results, err := checker.Check(args[1])
	if err != nil {
		return fmt.Errorf("error checking %s: %w", args[1], err)
	}

	for _, result := range results {
		fmt.Printf("%s:%d:%d: %s\n", result.FileName, result.Line, result.Column, result.Message)
	}

	return nil
}

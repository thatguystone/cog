package main

import (
	"fmt"
	"os"

	"github.com/iheartradio/cog/encoding/path"
)

func main() {
	var files []string

	if len(os.Args) > 1 {
		files = os.Args[1:]
	}

	if len(files) == 0 {
		// GOFILE is set by go generate
		file := os.Getenv("GOFILE")
		if file != "" {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		fmt.Printf("Usage: %s <FILE...>\n", os.Args[0])
		os.Exit(1)
	}

	for _, file := range files {
		err := path.GenerateFrom(file)
		if err != nil {
			fmt.Printf("Failed to generate for %s: %v\n", file, err)
		}
	}
}

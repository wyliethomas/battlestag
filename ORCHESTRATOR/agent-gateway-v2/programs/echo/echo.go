package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Input struct {
	Message string `json:"message"`
}

func main() {
	// Read JSON from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Echo the message
	fmt.Println(input.Message)
}

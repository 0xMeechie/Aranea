package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/0xMeechie/agenttest/pkg/agent"
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	a := agent.New(apiKey)

	fmt.Println("Agent ready. Type your request and press Enter. Type 'exit' to quit.")
	fmt.Println(strings.Repeat("-", 60))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" {
			fmt.Println("Goodbye.")
			break
		}

		response, err := a.Run(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\nAgent: %s\n", response)
	}
}

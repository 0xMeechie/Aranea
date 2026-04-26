package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/0xMeechie/Aranea/testagent/pkg/agent"
	"github.com/0xMeechie/Aranea/pkg/sdk"
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	aranea, err := sdk.InitFromFile("./testagent.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialize Aranea: %v\n", err)
		os.Exit(1)
	}
	defer aranea.Shutdown()

	a := agent.New(apiKey, &aranea.Tools)

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

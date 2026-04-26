package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xMeechie/Aranea/pkg/runtime"
	"github.com/0xMeechie/Aranea/pkg/sdk"
)

// toolDefinitions is the list of tools exposed to Claude
var toolDefinitions = []ToolDefinition{
	{
		Name:        "read_file",
		Description: "Read the contents of a file at the given path",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"path": {
					Type:        "string",
					Description: "Absolute or relative path to the file to read",
				},
			},
			Required: []string{"path"},
		},
	},
	{
		Name:        "write_file",
		Description: "Write content to a file at the given path, creating it if it doesn't exist",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"path": {
					Type:        "string",
					Description: "Absolute or relative path to the file to write",
				},
				"content": {
					Type:        "string",
					Description: "Content to write to the file",
				},
			},
			Required: []string{"path", "content"},
		},
	},
	{
		Name:        "http_request",
		Description: "Make an HTTP request to a URL and return the response body",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"url": {
					Type:        "string",
					Description: "The URL to make the request to",
				},
				"method": {
					Type:        "string",
					Description: "HTTP method to use",
					Enum:        []string{"GET", "POST", "PUT", "DELETE"},
				},
				"body": {
					Type:        "string",
					Description: "Request body (optional, for POST/PUT)",
				},
			},
			Required: []string{"url", "method"},
		},
	},
	{
		Name:        "run_command",
		Description: "Run a shell command and return its output. Use with caution.",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"command": {
					Type:        "string",
					Description: "The shell command to execute",
				},
			},
			Required: []string{"command"},
		},
	},
}

// dispatchTool routes a tool call through Aranea before executing it.
// Denied calls are never executed; a clear denial message is printed and returned instead.
func dispatchTool(ts *sdk.ToolSet, name string, rawInput json.RawMessage) string {
	fmt.Printf("\n  [tool call] %s — input: %s\n", name, string(rawInput))

	var result string
	var err error

	switch name {
	case "read_file":
		result, err = toolReadFile(ts, rawInput)
	case "write_file":
		result, err = toolWriteFile(ts, rawInput)
	case "http_request":
		result, err = toolHTTPRequest(ts, rawInput)
	case "run_command":
		result, err = toolRunCommand(ts, rawInput)
	default:
		os.Exit(2)
	}

	if err != nil {
		var denied *runtime.DeniedError
		if errors.As(err, &denied) {
			fmt.Printf("  [DENIED] %s\n", denied.Error())
			result = fmt.Sprintf("DENIED: %s", denied.Error())
		} else {
			result = fmt.Sprintf("error: %v", err)
		}
	}

	fmt.Printf("  [tool result] %s\n", truncate(result, 120))
	return result
}

func toolReadFile(ts *sdk.ToolSet, raw json.RawMessage) (string, error) {
	var input struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	return ts.ReadFile(filepath.Clean(input.Path))
}

func toolWriteFile(ts *sdk.ToolSet, raw json.RawMessage) (string, error) {
	var input struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	clean := filepath.Clean(input.Path)
	if err := ts.WriteFile(clean, input.Content); err != nil {
		return "", err
	}
	return fmt.Sprintf("successfully wrote %d bytes to %s", len(input.Content), clean), nil
}

func toolHTTPRequest(ts *sdk.ToolSet, raw json.RawMessage) (string, error) {
	var input struct {
		URL    string `json:"url"`
		Method string `json:"method"`
		Body   string `json:"body"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	return ts.HTTPRequest(input.Method, input.URL, input.Body)
}

func toolRunCommand(ts *sdk.ToolSet, raw json.RawMessage) (string, error) {
	var input struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	return ts.RunCommand(input.Command)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

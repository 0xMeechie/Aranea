package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

// dispatchTool routes a tool call to the correct implementation
func dispatchTool(name string, rawInput json.RawMessage) string {
	fmt.Printf("\n  [tool call] %s — input: %s\n", name, string(rawInput))

	var result string
	var err error

	switch name {
	case "read_file":
		result, err = toolReadFile(rawInput)
	case "write_file":
		result, err = toolWriteFile(rawInput)
	case "http_request":
		result, err = toolHTTPRequest(rawInput)
	case "run_command":
		result, err = toolRunCommand(rawInput)
	default:
		result = fmt.Sprintf("unknown tool: %s", name)
	}

	if err != nil {
		result = fmt.Sprintf("error: %v", err)
	}

	fmt.Printf("  [tool result] %s\n", truncate(result, 120))
	return result
}

// --- Tool implementations ---

func toolReadFile(raw json.RawMessage) (string, error) {
	var input struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	clean := filepath.Clean(input.Path)
	data, err := os.ReadFile(clean)
	if err != nil {
		return "", fmt.Errorf("reading file %q: %w", clean, err)
	}

	return string(data), nil
}

func toolWriteFile(raw json.RawMessage) (string, error) {
	var input struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	clean := filepath.Clean(input.Path)

	// ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(clean), 0755); err != nil {
		return "", fmt.Errorf("creating directories: %w", err)
	}

	if err := os.WriteFile(clean, []byte(input.Content), 0644); err != nil {
		return "", fmt.Errorf("writing file %q: %w", clean, err)
	}

	return fmt.Sprintf("successfully wrote %d bytes to %s", len(input.Content), clean), nil
}

func toolHTTPRequest(raw json.RawMessage) (string, error) {
	var input struct {
		URL    string `json:"url"`
		Method string `json:"method"`
		Body   string `json:"body"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}

	var bodyReader io.Reader
	if input.Body != "" {
		bodyReader = strings.NewReader(input.Body)
	}

	req, err := http.NewRequest(input.Method, input.URL, bodyReader)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	if input.Body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024)) // cap at 32KB
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	return fmt.Sprintf("status: %d\nbody:\n%s", resp.StatusCode, string(respBody)), nil
}

func toolRunCommand(raw json.RawMessage) (string, error) {
	var input struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	cmd := exec.Command("bash", "-c", input.Command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// return output even on non-zero exit — often useful context
		return fmt.Sprintf("exit error: %v\noutput:\n%s", err, string(out)), nil
	}

	return string(out), nil
}

// truncate shortens a string for display
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

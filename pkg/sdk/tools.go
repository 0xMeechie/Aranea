package sdk

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/0xMeechie/Aranea/pkg/runtime"
)

// ToolSet exposes instrumented tool operations.
// Every call is evaluated against the policy engine before execution.
// A denied call never executes the underlying operation.
type ToolSet struct {
	agentID string
	rt      *runtime.Runtime
}

func newToolSet(agentID string, rt *runtime.Runtime) ToolSet {
	return ToolSet{agentID: agentID, rt: rt}
}

// evaluate calls runtime.Evaluate and returns the error directly (nil on allow, *DeniedError on deny).
func (ts *ToolSet) evaluate(tool, action, target string, extra map[string]any) error {
	args := map[string]any{
		"action": action,
		"target": target,
	}
	for k, v := range extra {
		args[k] = v
	}
	_, err := ts.rt.Evaluate(runtime.ToolCallRequest{
		AgentID: ts.agentID,
		Tool:    tool,
		Args:    args,
	})
	return err
}

// ReadFile reads the file at path if policy allows it.
func (ts *ToolSet) ReadFile(path string) (string, error) {
	if err := ts.evaluate("read_file", "read", path, nil); err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

// WriteFile writes content to path if policy allows it.
func (ts *ToolSet) WriteFile(path, content string) error {
	if err := ts.evaluate("write_file", "write", path, nil); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// HTTPRequest performs an HTTP request to rawURL if policy allows it.
// The domain extracted from rawURL is checked against the policy's allowed domains.
func (ts *ToolSet) HTTPRequest(method, rawURL, body string) (string, error) {
	host, err := extractHost(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	if err := ts.evaluate("http_request", method, host, map[string]any{"url": rawURL}); err != nil {
		return "", err
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return string(data), nil
}

// RunCommand executes command in a shell if policy allows it.
func (ts *ToolSet) RunCommand(command string) (string, error) {
	if err := ts.evaluate("run_command", "run", command, nil); err != nil {
		return "", err
	}
	out, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		return "", fmt.Errorf("run command: %w", err)
	}
	return string(out), nil
}

func extractHost(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("no host in url %q", rawURL)
	}
	return u.Hostname(), nil
}

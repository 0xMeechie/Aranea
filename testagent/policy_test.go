package main

import (
	"errors"
	"os"
	"testing"

	"github.com/0xMeechie/Aranea/pkg/runtime"
	"github.com/0xMeechie/Aranea/pkg/sdk"
)

// TestPolicy loads the real testagent.yaml and runs tool calls through the
// Aranea runtime so we can verify allow/deny behaviour without a live API key.
func TestPolicy(t *testing.T) {
	// Run from the testagent directory so the relative config path resolves.
	if err := os.Chdir("testagent"); err != nil {
		// Already in testagent/ if test is run with `go test ./testagent/...`
		// from the repo root; Chdir will fail harmlessly.
		_ = err
	}

	aranea, err := sdk.InitFromFile("./testagent.yaml")
	if err != nil {
		t.Fatalf("InitFromFile: %v", err)
	}
	defer aranea.Shutdown()

	ts := &aranea.Tools

	cases := []struct {
		name    string
		call    func() error
		wantErr bool // true = expect DeniedError
	}{
		{
			name:    "read_file allowed",
			call:    func() error { _, err := ts.ReadFile("testagent.yaml"); return err },
			wantErr: false,
		},
		{
			name:    "write_file allowed",
			call:    func() error { return ts.WriteFile("_test_out.txt", "hello") },
			wantErr: false,
		},
		{
			name:    "http GET httpbin.org allowed",
			call:    func() error { _, err := ts.HTTPRequest("GET", "https://httpbin.org/get", ""); return err },
			wantErr: false,
		},
		{
			name:    "http POST httpbin.org denied (action not whitelisted)",
			call:    func() error { _, err := ts.HTTPRequest("POST", "https://httpbin.org/post", "{}"); return err },
			wantErr: true,
		},
		{
			name:    "http GET non-whitelisted domain denied",
			call:    func() error { _, err := ts.HTTPRequest("GET", "https://example.com/", ""); return err },
			wantErr: true,
		},
		{
			name:    "run_command denied (no rule, default deny)",
			call:    func() error { _, err := ts.RunCommand("echo hello"); return err },
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			var denied *runtime.DeniedError
			isDenied := errors.As(err, &denied)

			if tc.wantErr && !isDenied {
				t.Fatalf("expected DeniedError, got: %v", err)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if isDenied {
				t.Logf("DENIED: %s", denied.Error())
			} else {
				t.Logf("ALLOWED")
			}
		})
	}

	// Clean up write test artifact
	_ = os.Remove("_test_out.txt")
}

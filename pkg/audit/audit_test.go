package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func makeEvent(id string) Event {
	return Event{
		EventID:      id,
		Timestamp:    time.Now().UTC(),
		NodeID:       "node-1",
		AgentID:      "agent-1",
		AgentVersion: "0.1.0",
		Tool:         "read_file",
		Args:         map[string]any{"path": "/tmp/a.txt"},
		Decision:     "allow",
		Reason:       "rule matched",
		RuleMatched:  "allow-workspace",
		PolicyVersion: "v1",
		Signature:    "sig-abc",
		Context: EventContext{
			PromptSummary:    "summarise the file",
			ConversationTurn: 3,
		},
	}
}

func readLines(t *testing.T, path string) [][]byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open log file: %v", err)
	}
	defer f.Close()
	var lines [][]byte
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		b := make([]byte, len(sc.Bytes()))
		copy(b, sc.Bytes())
		lines = append(lines, b)
	}
	return lines
}

func logPath(dir, agentID string) string {
	return filepath.Join(dir, agentID+".log")
}

func TestNewLogCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "audit")
	l, err := NewLog(dir, "agent-1")
	if err != nil {
		t.Fatalf("NewLog: %v", err)
	}
	defer l.Close()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatal("log directory was not created")
	}
}

func TestWriteAppendsValidJSONLine(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLog(dir, "agent-1")
	if err != nil {
		t.Fatalf("NewLog: %v", err)
	}
	defer l.Close()

	if err := l.Write(makeEvent("evt-1")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	l.Close()

	lines := readLines(t, logPath(dir, "agent-1"))
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	var e Event
	if err := json.Unmarshal(lines[0], &e); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if e.EventID != "evt-1" {
		t.Errorf("eventID: got %q", e.EventID)
	}
}

func TestWriteThreeTimesProducesThreeLines(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLog(dir, "agent-1")
	if err != nil {
		t.Fatalf("NewLog: %v", err)
	}

	for i, id := range []string{"e1", "e2", "e3"} {
		if err := l.Write(makeEvent(id)); err != nil {
			t.Fatalf("Write %d: %v", i, err)
		}
	}
	l.Close()

	lines := readLines(t, logPath(dir, "agent-1"))
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
}

func TestEachLineContainsAllFields(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLog(dir, "agent-1")
	if err != nil {
		t.Fatalf("NewLog: %v", err)
	}
	l.Write(makeEvent("full-event"))
	l.Close()

	lines := readLines(t, logPath(dir, "agent-1"))
	var e Event
	if err := json.Unmarshal(lines[0], &e); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	checks := map[string]bool{
		"eventID":      e.EventID != "",
		"timestamp":    !e.Timestamp.IsZero(),
		"nodeID":       e.NodeID != "",
		"agentID":      e.AgentID != "",
		"agentVersion": e.AgentVersion != "",
		"tool":         e.Tool != "",
		"args":         e.Args != nil,
		"decision":     e.Decision != "",
		"reason":       e.Reason != "",
		"signature":    e.Signature != "",
		"context":      e.Context.PromptSummary != "",
	}
	for field, ok := range checks {
		if !ok {
			t.Errorf("field %q is empty or zero", field)
		}
	}
}

func TestLogNotTruncatedBetweenRestarts(t *testing.T) {
	dir := t.TempDir()

	// first "process"
	l1, err := NewLog(dir, "agent-1")
	if err != nil {
		t.Fatalf("NewLog first: %v", err)
	}
	l1.Write(makeEvent("e1"))
	l1.Write(makeEvent("e2"))
	l1.Close()

	// second "process" — reopen the same file
	l2, err := NewLog(dir, "agent-1")
	if err != nil {
		t.Fatalf("NewLog second: %v", err)
	}
	l2.Write(makeEvent("e3"))
	l2.Close()

	lines := readLines(t, logPath(dir, "agent-1"))
	if len(lines) != 3 {
		t.Fatalf("expected 3 accumulated lines, got %d", len(lines))
	}
}

func TestConcurrentWritesDoNotCorrupt(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLog(dir, "agent-1")
	if err != nil {
		t.Fatalf("NewLog: %v", err)
	}
	defer l.Close()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			l.Write(makeEvent(string(rune('A' + i))))
		}(i)
	}
	wg.Wait()
	l.Close()

	lines := readLines(t, logPath(dir, "agent-1"))
	if len(lines) != goroutines {
		t.Fatalf("expected %d lines, got %d", goroutines, len(lines))
	}
	for i, line := range lines {
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

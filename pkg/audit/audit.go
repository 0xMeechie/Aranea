package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventContext holds metadata about the agent invocation that triggered the tool call.
type EventContext struct {
	PromptSummary    string `json:"promptSummary,omitempty"`
	ConversationTurn int    `json:"conversationTurn,omitempty"`
}

// Event is a single immutable audit log entry.
type Event struct {
	EventID       string         `json:"eventID"`
	Timestamp     time.Time      `json:"timestamp"`
	NodeID        string         `json:"nodeID"`
	AgentID       string         `json:"agentID"`
	AgentVersion  string         `json:"agentVersion"`
	Tool          string         `json:"tool"`
	Args          map[string]any `json:"args"`
	Decision      string         `json:"decision"`
	Reason        string         `json:"reason"`
	RuleMatched   string         `json:"ruleMatched,omitempty"`
	PolicyVersion string         `json:"policyVersion,omitempty"`
	Signature     string         `json:"signature"`
	Context       EventContext   `json:"context"`
}

// Signable is the subset of Event fields covered by the signature.
// Signature itself is excluded to avoid a circular dependency.
type Signable struct {
	EventID      string         `json:"eventID"`
	Timestamp    time.Time      `json:"timestamp"`
	NodeID       string         `json:"nodeID"`
	AgentID      string         `json:"agentID"`
	AgentVersion string         `json:"agentVersion"`
	Tool         string         `json:"tool"`
	Args         map[string]any `json:"args"`
	Decision     string         `json:"decision"`
	Reason       string         `json:"reason"`
}

// Log appends audit events as newline-delimited JSON to a per-agent flat file.
type Log struct {
	mu   sync.Mutex
	file *os.File
}

// NewLog creates logDir if it does not exist, then opens (or creates) {agentID}.log
// for append-only writing. The file is never truncated.
func NewLog(logDir, agentID string) (*Log, error) {
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return nil, fmt.Errorf("create audit directory: %w", err)
	}
	path := filepath.Join(logDir, agentID+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open audit log: %w", err)
	}
	return &Log{file: f}, nil
}

// Write serialises the event as a single JSON line and appends it to the log.
// It is safe to call concurrently.
func (l *Log) Write(e Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = fmt.Fprintf(l.file, "%s\n", data)
	return err
}

// Close releases the underlying file handle.
func (l *Log) Close() error {
	return l.file.Close()
}

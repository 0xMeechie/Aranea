package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0xMeechie/Aranea/pkg/audit"
	"github.com/0xMeechie/Aranea/pkg/config"
	"github.com/0xMeechie/Aranea/pkg/policy"
)

const araneaDir = ".aranea"

type Runtime struct {
	cfg      map[string]*config.AgentConfig
	engine   *policy.PolicyEngine
	auditLog *audit.Log
	key      *policy.KeyPair
	nodeID   string
}

// CallContext holds metadata about the agent invocation that triggered the tool call.
type CallContext struct {
	PromptSummary    string
	ConversationTurn int
}

// ToolCallRequest is an incoming request to evaluate a tool call against policy.
// Action and Target within Args drive policy evaluation:
//
//	Args["action"] — the verb (GET, POST, DELETE, SELECT, read, write, etc.)
//	Args["target"] — what it acts on (domain, file path, table name, etc.)
type ToolCallRequest struct {
	AgentID     string
	Tool        string
	Args        map[string]any
	CallContext CallContext
}

// New wires up the Runtime from the given agent config.
// It initialises the policy engine, audit log, key pair, and node ID.
func New(nodeID string, keys policy.KeyPair) (*Runtime, error) {
	cfgs := make(map[string]*config.AgentConfig)
	pe := policy.NewPolicyEngine(cfgs)

	return &Runtime{
		engine: pe,
		cfg:    cfgs,
		key:    &keys,
		nodeID: nodeID,
	}, nil
}

func (r *Runtime) AddAgent(cfg *config.AgentConfig) {
}

func (r *Runtime) Sign(payload audit.Signable) (string, error) {
	return r.key.Sign(payload)
}

// Evaluate checks the tool call against policy, signs and writes an audit event,
// and returns the decision.
func (r *Runtime) Evaluate(req ToolCallRequest) (policy.Decision, error) {
	action, _ := req.Args["action"].(string)
	target, _ := req.Args["target"].(string)

	decision := r.engine.Evaluate(policy.RuleRequest{
		CallingAgent: req.AgentID,
		Tool:         req.Tool,
		Action:       action,
		Target:       target,
	})

	eventID := generateEventID()
	now := time.Now().UTC()

	signable := audit.Signable{
		EventID:   eventID,
		Timestamp: now,
		AgentID:   req.AgentID,
		Tool:      req.Tool,
		Args:      req.Args,
		Decision:  string(decision.Action),
		Reason:    decision.Reason,
	}

	sig, err := r.Sign(signable)
	if err != nil {
		return policy.Decision{}, fmt.Errorf("unable to sign request %s", err)
	}
	event := audit.Event{
		EventID:     eventID,
		Timestamp:   now,
		AgentID:     req.AgentID,
		Tool:        req.Tool,
		Args:        req.Args,
		Decision:    string(decision.Action),
		Reason:      decision.Reason,
		RuleMatched: decision.DeterminingRule,
		Signature:   sig,
		Context: audit.EventContext{
			PromptSummary:    req.CallContext.PromptSummary,
			ConversationTurn: req.CallContext.ConversationTurn,
		},
	}

	if err := r.auditLog.Write(event); err != nil {
		return policy.Decision{}, fmt.Errorf("write audit event: %w", err)
	}

	if decision.Action == policy.Deny {
		return decision, &DeniedError{
			Tool:        req.Tool,
			Reason:      decision.Reason,
			RuleMatched: decision.DeterminingRule,
		}
	}

	return decision, nil
}

// Close flushes and releases any resources held by the Runtime.
func (r *Runtime) Close() error {
	return r.auditLog.Close()
}

// generateEventID returns a cryptographically random 16-byte hex string.
func generateEventID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// getOrCreateNodeID reads the node ID from .aranea/node-id.
// If the file does not exist, a new random ID is generated and persisted.
// Repeated calls return the same ID.
func getOrCreateNodeID() (string, error) {
	path := filepath.Join(araneaDir, "node-id")

	data, err := os.ReadFile(path)
	if err == nil {
		if id := strings.TrimSpace(string(data)); id != "" {
			return id, nil
		}
	}

	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read node-id: %w", err)
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate node-id: %w", err)
	}
	id := hex.EncodeToString(b)

	if err := os.MkdirAll(araneaDir, 0o700); err != nil {
		return "", fmt.Errorf("create .aranea dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(id), 0o600); err != nil {
		return "", fmt.Errorf("write node-id: %w", err)
	}

	return id, nil
}

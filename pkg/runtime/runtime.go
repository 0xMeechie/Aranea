package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/0xMeechie/Aranea/pkg/audit"
	"github.com/0xMeechie/Aranea/pkg/config"
	"github.com/0xMeechie/Aranea/pkg/policy"
)

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
func New(nodeID string, keys policy.KeyPair, logDir string) (*Runtime, error) {
	cfgs := make(map[string]*config.AgentConfig)
	pe := policy.NewPolicyEngine(cfgs)

	log, err := audit.NewLog(logDir, nodeID)
	if err != nil {
		return nil, fmt.Errorf("open audit log: %w", err)
	}

	return &Runtime{
		engine:   pe,
		cfg:      cfgs,
		auditLog: log,
		key:      &keys,
		nodeID:   nodeID,
	}, nil
}

func (r *Runtime) AddAgent(cfg *config.AgentConfig) {
	r.cfg[cfg.AgentConfig.ID] = cfg
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

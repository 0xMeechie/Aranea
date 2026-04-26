package policy

import (
	"strings"

	"github.com/0xMeechie/Aranea/pkg/config"
)

// Action represents the outcome of a policy evaluation.
type Action string

const (
	Allow Action = "allow"
	Deny  Action = "deny"
)

// Decision is the result of evaluating a tool call against the policy.
type Decision struct {
	Action          Action
	Reason          string
	DeterminingRule string
}

type PolicyEngine struct {
	AgentConfigs map[string]*config.AgentConfig
}

// RuleRequest is an incoming evaluation of a tool call.
// Action is the verb the agent wants to perform (GET, POST, DELETE, SELECT, read, write, etc.)
// Target is what it's acting on (domain, file path, table name, etc.)
type RuleRequest struct {
	CallingAgent string
	Tool         string
	Action       string
	Target       string
}

// NewPolicyEngine creates a PolicyEngine from the given config.
func NewPolicyEngine(cfgs map[string]*config.AgentConfig) *PolicyEngine {
	return &PolicyEngine{
		AgentConfigs: cfgs,
	}
}

// AddAgentConfig adds a new config to the policy engine. This should be used when a new agent comes online.
func (pe *PolicyEngine) AddAgentConfig(agentName string, cfg *config.AgentConfig) {
	_, exist := pe.AgentConfigs[agentName]
	if !exist {
		pe.AgentConfigs[agentName] = cfg
	}
}

// Evaluate checks whether the tool call is permitted under the policy.
func (pe *PolicyEngine) Evaluate(req RuleRequest) Decision {
	cfg, exist := pe.AgentConfigs[req.CallingAgent]
	if !exist {
		return Decision{Action: Deny, Reason: "no agent config found for agent"}
	}

	rule, exist := cfg.PolicyConfig.RuleMap[req.Tool]
	if !exist {
		if cfg.PolicyConfig.DefaultBehavior == string(Allow) {
			return Decision{Action: Allow, Reason: "allowed by default policy"}
		}
		return Decision{Action: Deny, Reason: "tool not permitted by policy"}
	}

	// validate path-based tools (read_file, write_file)
	if len(rule.Paths) > 0 {
		if d, ok := checkPaths(req, rule); !ok {
			d.DeterminingRule = req.Tool
			return d
		}
	}

	// validate network-based tools (http_request)
	if len(rule.Domains) > 0 {
		if d, ok := checkDomains(rule.Domains, req.Target); !ok {
			d.DeterminingRule = req.Tool
			return d
		}
	}

	// validate table-based tools (sql_command)
	if len(rule.Tables) > 0 {
		if d, ok := checkTables(rule.Tables, req.Target); !ok {
			d.DeterminingRule = req.Tool
			return d
		}
	}

	// validate allowed actions (http_request, sql_command)
	if len(rule.AllowedActions) > 0 {
		if d, ok := checkAllowedActions(rule.AllowedActions, req.Action); !ok {
			d.DeterminingRule = req.Tool
			return d
		}
	}

	return Decision{Action: Allow, Reason: "rule matched", DeterminingRule: req.Tool}
}

// checkPaths validates the target path against the rule's allowed path patterns.
func checkPaths(req RuleRequest, policy config.ParamConfig) (Decision, bool) {
	if req.Target == "" {
		return Decision{Action: Deny, Reason: "path target is required"}, false
	}
	for _, path := range policy.Paths {
		if matchGlob(path, req.Target) {
			return Decision{Action: Allow, Reason: "allowed path found"}, true
		}
	}
	return Decision{Action: Deny, Reason: "path not permitted by policy"}, false
}

// checkDomains validates the target domain against the rule's allowed domains.
func checkDomains(allowed []string, target string) (Decision, bool) {
	if target == "" {
		return Decision{Action: Deny, Reason: "domain target is required"}, false
	}
	for _, domain := range allowed {
		if strings.EqualFold(domain, target) {
			return Decision{Action: Allow, Reason: "allowed domain found"}, true
		}
	}
	return Decision{Action: Deny, Reason: "domain not permitted by policy"}, false
}

// checkTables validates the target table against the rule's allowed tables.
func checkTables(allowed []string, target string) (Decision, bool) {
	if target == "" {
		return Decision{Action: Deny, Reason: "table target is required"}, false
	}
	for _, table := range allowed {
		if strings.EqualFold(table, target) {
			return Decision{}, true
		}
	}
	return Decision{Action: Deny, Reason: "table not permitted by policy"}, false
}

// checkAllowedActions validates the requested action against the rule's allowed actions.
func checkAllowedActions(allowed []string, action string) (Decision, bool) {
	if action == "" {
		return Decision{Action: Deny, Reason: "action is required"}, false
	}
	for _, a := range allowed {
		if strings.EqualFold(a, action) {
			return Decision{}, true
		}
	}
	return Decision{Action: Deny, Reason: "action not permitted by policy"}, false
}

// matchGlob supports /** suffix: pattern /tmp/workspace/** matches any path
// rooted at /tmp/workspace/. Exact match is also supported.
func matchGlob(pattern, path string) bool {
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}
	return pattern == path
}

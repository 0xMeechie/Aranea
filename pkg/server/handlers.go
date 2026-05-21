package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/0xMeechie/Aranea/pkg/policy"
	"github.com/0xMeechie/Aranea/pkg/runtime"
)

// Version is the runtime API version string. Bump alongside breaking HTTP
// contract changes so clients can detect them.
const Version = "0.1.0"

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, r, http.StatusOK, map[string]string{"version": Version})
}

// EvaluateRequest is the JSON envelope POSTed to /v1/evaluate. Args follow the
// runtime contract (Args["action"], Args["target"]).
type EvaluateRequest struct {
	AgentID          string         `json:"agentID"`
	Tool             string         `json:"tool"`
	Args             map[string]any `json:"args"`
	PromptSummary    string         `json:"promptSummary,omitempty"`
	ConversationTurn int            `json:"conversationTurn,omitempty"`
}

type RegisterAgentRequest struct {
	AgentID          string         `json:"agentID"`
	Tool             string         `json:"tool"`
	Args             map[string]any `json:"args"`
	PromptSummary    string         `json:"promptSummary,omitempty"`
	ConversationTurn int            `json:"conversationTurn,omitempty"`
}

// EvaluateResponse mirrors policy.Decision for clients.
type EvaluateResponse struct {
	Action          string `json:"action"`
	Reason          string `json:"reason"`
	DeterminingRule string `json:"determiningRule,omitempty"`
}

func (s *Server) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var req EvaluateRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// An authenticated agentID overrides the body so callers cannot impersonate.
	if id := AgentIDFromContext(r.Context()); id != "" {
		req.AgentID = id
	}
	if req.AgentID == "" || req.Tool == "" {
		WriteError(w, r, http.StatusBadRequest, "agentID and tool are required")
		return
	}

	decision, err := s.runtime.Evaluate(runtime.ToolCallRequest{
		AgentID: req.AgentID,
		Tool:    req.Tool,
		Args:    req.Args,
		CallContext: runtime.CallContext{
			PromptSummary:    req.PromptSummary,
			ConversationTurn: req.ConversationTurn,
		},
	})

	var denied *runtime.DeniedError
	if errors.As(err, &denied) {
		WriteJSON(w, r, http.StatusForbidden, EvaluateResponse{
			Action:          string(policy.Deny),
			Reason:          denied.Reason,
			DeterminingRule: denied.RuleMatched,
		})
		return
	}
	if err != nil {
		s.log.Error("evaluate failed",
			slog.String("request_id", RequestIDFromContext(r.Context())),
			slog.Any("err", err),
		)
		WriteError(w, r, http.StatusInternalServerError, "evaluation failed")
		return
	}

	WriteJSON(w, r, http.StatusOK, EvaluateResponse{
		Action:          string(decision.Action),
		Reason:          decision.Reason,
		DeterminingRule: decision.DeterminingRule,
	})
}

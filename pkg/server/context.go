package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/0xMeechie/Aranea/pkg/runtime"
)

type ctxKey int

const (
	ctxKeyRequestID ctxKey = iota
	ctxKeyRuntime
	ctxKeyAgentID
	ctxKeyPrincipal
)

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, id)
}

// RequestIDFromContext returns the ID assigned by the RequestID middleware,
// or "" if none was set.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyRequestID).(string)
	return id
}

func withRuntime(ctx context.Context, rt *runtime.Runtime) context.Context {
	return context.WithValue(ctx, ctxKeyRuntime, rt)
}

// RuntimeFromContext returns the runtime handle injected by WithRuntime.
// Handlers registered via Server.Handle/HandleFunc can rely on it being present.
func RuntimeFromContext(ctx context.Context) *runtime.Runtime {
	rt, _ := ctx.Value(ctxKeyRuntime).(*runtime.Runtime)
	return rt
}

func withPrincipal(ctx context.Context, p Principal) context.Context {
	ctx = context.WithValue(ctx, ctxKeyPrincipal, p)
	return context.WithValue(ctx, ctxKeyAgentID, p.AgentID)
}

// PrincipalFromContext returns the authenticated Principal or the zero value
// if the route is unauthenticated.
func PrincipalFromContext(ctx context.Context) Principal {
	p, _ := ctx.Value(ctxKeyPrincipal).(Principal)
	return p
}

// AgentIDFromContext returns the authenticated agent's ID, populated by
// RequireAuth. Empty when the route is unauthenticated.
func AgentIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyAgentID).(string)
	return id
}

func newRequestID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

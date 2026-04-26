package server

import (
	"errors"
	"net/http"
	"strings"
)

// ErrUnauthorized is returned by Authenticator implementations when credentials
// are missing or invalid. RequireAuth maps it to a uniform 401.
var ErrUnauthorized = errors.New("unauthorized")

// Principal is the result of a successful authentication. AgentID is required;
// implementations may attach arbitrary claims via Metadata.
type Principal struct {
	AgentID  string
	Metadata map[string]string
}

// Authenticator validates inbound credentials and returns the resolved principal.
// Implementations must return ErrUnauthorized for invalid credentials.
type Authenticator interface {
	Authenticate(r *http.Request) (Principal, error)
}

// AuthenticatorFunc adapts a plain function into an Authenticator.
type AuthenticatorFunc func(r *http.Request) (Principal, error)

func (f AuthenticatorFunc) Authenticate(r *http.Request) (Principal, error) {
	return f(r)
}

// BearerTokenAuth builds an Authenticator from a token → agentID map. Use it
// as a starting point; replace with a real credential store for production.
func BearerTokenAuth(tokens map[string]string) Authenticator {
	return AuthenticatorFunc(func(r *http.Request) (Principal, error) {
		h := r.Header.Get("Authorization")
		if h == "" {
			return Principal{}, ErrUnauthorized
		}
		scheme, token, ok := strings.Cut(h, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") {
			return Principal{}, ErrUnauthorized
		}
		agentID, ok := tokens[strings.TrimSpace(token)]
		if !ok {
			return Principal{}, ErrUnauthorized
		}
		return Principal{AgentID: agentID}, nil
	})
}

// RequireAuth wraps a handler with the supplied Authenticator. A nil
// Authenticator is a no-op, letting callers disable auth in development without
// conditional wiring.
func RequireAuth(a Authenticator) Middleware {
	return func(next http.Handler) http.Handler {
		if a == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, err := a.Authenticate(r)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Bearer realm="aranea"`)
				WriteError(w, r, http.StatusUnauthorized, "unauthorized")
				return
			}
			next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
		})
	}
}

package server

import "net/http"

// routes registers the built-in routes. Callers register additional routes via
// Server.Handle/HandleFunc/HandleAuth — those mount on the same mux and share
// the middleware chain configured in Handler().
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /version", s.handleVersion)

	s.HandleAuth("POST /v1/evaluate", http.HandlerFunc(s.handleEvaluate))
}

// Handle registers an http.Handler on the underlying mux. Pattern follows
// Go 1.22+ syntax (e.g. "GET /widgets/{id}").
func (s *Server) Handle(pattern string, h http.Handler) {
	s.mux.Handle(pattern, h)
}

// HandleFunc is the http.HandlerFunc variant of Handle.
func (s *Server) HandleFunc(pattern string, h http.HandlerFunc) {
	s.mux.HandleFunc(pattern, h)
}

// HandleAuth registers a handler that requires successful authentication
// against Config.Authenticator. With a nil Authenticator this is equivalent
// to Handle.
func (s *Server) HandleAuth(pattern string, h http.Handler) {
	s.mux.Handle(pattern, RequireAuth(s.cfg.Authenticator)(h))
}

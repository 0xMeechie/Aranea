package server

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	arruntime "github.com/0xMeechie/Aranea/pkg/runtime"
)

// Middleware is the canonical http.Handler decorator. Compose with Chain or
// pass via Config.Middleware to register on every route.
type Middleware func(http.Handler) http.Handler

// Chain wraps h with the given middleware. The first entry is the outermost —
// it sees the request first and the response last.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RequestID assigns each request a stable identifier and exposes it via the
// request context (RequestIDFromContext) and the X-Request-ID response header.
// An inbound X-Request-ID is preserved.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), id)))
	})
}

// LoggerMiddleware emits one structured log line per request.
func LoggerMiddleware(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			log.LogAttrs(r.Context(), slog.LevelInfo, "http request",
				slog.String("request_id", RequestIDFromContext(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Int("bytes", sw.bytes),
				slog.String("remote", r.RemoteAddr),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// Recoverer turns a panicking handler into a 500 response and logs the stack.
func Recoverer(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic in handler",
						slog.String("request_id", RequestIDFromContext(r.Context())),
						slog.Any("panic", rec),
						slog.String("stack", string(debug.Stack())),
					)
					WriteError(w, r, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// WithRuntime injects the runtime handle into the request context so any
// handler — including ones registered by callers — can fetch it via
// RuntimeFromContext without holding a closure.
func WithRuntime(rt *arruntime.Runtime) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(withRuntime(r.Context(), rt)))
		})
	}
}

// Timeout cancels the request context after d and writes msg as the body if
// the handler does not return in time. Wraps the stdlib http.TimeoutHandler so
// callers get the same semantics in a Middleware shape.
func Timeout(d time.Duration, msg string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, msg)
	}
}

// statusWriter captures the response status and byte count for the logger.
type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
	wrote  bool
}

func (w *statusWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.status = code
	w.wrote = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.wrote = true
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

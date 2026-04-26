package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xMeechie/Aranea/pkg/runtime"
)

const (
	defaultAddr            = ":8080"
	defaultReadTimeout     = 15 * time.Second
	defaultWriteTimeout    = 30 * time.Second
	defaultIdleTimeout     = 120 * time.Second
	defaultShutdownTimeout = 30 * time.Second
	defaultMaxHeaderBytes  = 1 << 20
)

// Config controls HTTP server behaviour. Zero values fall back to safe defaults.
type Config struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int
	Logger          *slog.Logger

	// Authenticator validates inbound credentials. Nil disables authentication
	// for routes registered with HandleAuth — convenient for local development.
	Authenticator Authenticator

	// Middleware is applied to every route, after the built-in chain
	// (RequestID → Logger → Recoverer → WithRuntime). The first entry is the
	// outermost wrapper.
	Middleware []Middleware
}

// Server is the HTTP entry point for the Aranea runtime. Construct it with New,
// register routes via Handle/HandleFunc/HandleAuth, then call Start.
type Server struct {
	cfg     Config
	runtime *runtime.Runtime
	log     *slog.Logger
	http    *http.Server
	mux     *http.ServeMux
}

// New constructs a Server bound to rt. The server is not started until Start.
func New(rt *runtime.Runtime, cfg Config) *Server {
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = defaultWriteTimeout
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = defaultIdleTimeout
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = defaultShutdownTimeout
	}
	if cfg.MaxHeaderBytes == 0 {
		cfg.MaxHeaderBytes = defaultMaxHeaderBytes
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	s := &Server{
		cfg:     cfg,
		runtime: rt,
		log:     cfg.Logger,
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s
}

// Runtime returns the runtime handle bound to this server.
func (s *Server) Runtime() *runtime.Runtime { return s.runtime }

// Logger returns the structured logger used by the server and its middleware.
func (s *Server) Logger() *slog.Logger { return s.log }

// Handler returns the fully wrapped http.Handler with the middleware chain applied.
// Useful for tests (httptest.NewServer) or for mounting Aranea inside a parent mux.
func (s *Server) Handler() http.Handler {
	chain := []Middleware{
		RequestID,
		LoggerMiddleware(s.log),
		Recoverer(s.log),
		WithRuntime(s.runtime),
	}
	chain = append(chain, s.cfg.Middleware...)
	return Chain(s.mux, chain...)
}

// Start runs the server until SIGINT or SIGTERM, then performs a graceful
// shutdown bounded by Config.ShutdownTimeout.
func (s *Server) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return s.StartContext(ctx)
}

// StartContext is Start with a caller-provided cancellation context.
func (s *Server) StartContext(ctx context.Context) error {
	s.http = &http.Server{
		Handler:           s.Handler(),
		ReadTimeout:       s.cfg.ReadTimeout,
		ReadHeaderTimeout: s.cfg.ReadTimeout,
		WriteTimeout:      s.cfg.WriteTimeout,
		IdleTimeout:       s.cfg.IdleTimeout,
		MaxHeaderBytes:    s.cfg.MaxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(s.log.Handler(), slog.LevelError),
	}

	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.cfg.Addr, err)
	}

	serveErr := make(chan error, 1)
	go func() {
		s.log.Info("server listening", slog.String("addr", ln.Addr().String()))
		if err := s.http.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case <-ctx.Done():
		s.log.Info("server shutdown requested")
		return s.shutdown()
	case err := <-serveErr:
		return err
	}
}

func (s *Server) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	s.log.Info("server stopped")
	return nil
}

package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/history"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/verification"
)

// Server coordinates HTTP routing, middleware, and handlers.
type Server struct {
	addr         string
	fleetService *fleet.Service
	verifier     verification.Verifier
	releases     history.Repository
	logger       *slog.Logger
	server       *http.Server
}

// NewServer initializes a new API server.
func NewServer(
	addr string,
	fleetService *fleet.Service,
	verifier verification.Verifier,
	releases history.Repository,
	logger *slog.Logger,
) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		addr:         addr,
		fleetService: fleetService,
		verifier:     verifier,
		releases:     releases,
		logger:       logger,
	}
}

// Handler configures and returns the HTTP request multiplexer.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/fleets", s.handleListFleets)
	mux.HandleFunc("GET /v1/fleets/{owner}/{tag}", s.handleGetFleet)
	mux.HandleFunc("GET /v1/fleets/{owner}/{tag}/members", s.handleListMembers)
	mux.HandleFunc("GET /v1/fleets/{owner}/{tag}/history", s.handleFleetHistory)
	mux.HandleFunc("GET /v1/fleets/{owner}/{tag}/releases", s.handleFleetReleases)
	mux.HandleFunc("GET /v1/fleets/{owner}/{tag}/verify", s.handleVerifyFleet)
	mux.HandleFunc("GET /v1/contracts/{contract_id}", s.handleInspectContract)

	// Wrap with common middleware (logging, panic recovery, CORS)
	return s.recoveryMiddleware(s.loggingMiddleware(mux))
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("starting http api server", "addr", s.addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

// Shutdown gracefully terminates the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Debug("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.logger.Error("panic recovered in http handler", "panic", rec)
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

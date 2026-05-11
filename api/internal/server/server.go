package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/abhi-jeet589/mocksmith/internal/config"
	"github.com/abhi-jeet589/mocksmith/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	cfg  config.Config
	repo handler.Store
}

func New(cfg config.Config, repo handler.Store) *Server {
	return &Server{cfg: cfg, repo: repo}
}

// Handler builds the HTTP handler tree. Exposed so tests can exercise routing
// without binding a real listener.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	admin := &handler.Admin{Repo: s.repo}
	r.Route("/admin/mocks", admin.Routes)

	// Any path not claimed above falls through to the dynamic mock handler,
	// which looks the request up by (method, path) in the store.
	mockH := &handler.Mock{Repo: s.repo}
	r.NotFound(mockH.Serve)
	r.MethodNotAllowed(mockH.Serve)

	return r
}

func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              ":" + s.cfg.Port,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s", s.cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Println("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

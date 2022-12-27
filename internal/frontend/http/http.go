package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"template/internal/config"
	"template/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type HTTP struct {
	cfg    config.Frontend
	db     database.Database
	server *http.Server
	logger *zerolog.Logger
}

func New(cfg config.Frontend, db database.Database, logger *zerolog.Logger) *HTTP {
	l := logger.With().Str("object", "frontend").Str("type", "http").Str("name", cfg.Name).Logger()
	h := &HTTP{
		cfg:    cfg,
		db:     db,
		logger: &l,
	}

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Post("/test", h.test)

	h.server = &http.Server{
		Handler: r,
	}

	return h
}

func (h *HTTP) Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer cancel()

	ln, errListen := net.Listen("tcp", fmt.Sprintf("%s:%d", h.cfg.Host, h.cfg.Port))
	if errListen != nil {
		h.logger.Error().Err(errListen).Str("address", fmt.Sprintf("%s:%d", h.cfg.Host, h.cfg.Port)).Msg("error listen address")
		return
	}
	defer func() {
		err := ln.Close()
		if err != nil {
			h.logger.Error().Err(err).Msg("error close listener")
		}
	}()

	go func() {
		<-ctx.Done()
		errShutdown := h.server.Shutdown(context.Background())
		if errShutdown != nil {
			h.logger.Error().Err(errShutdown).Msg("shutdown error")
		}
	}()

	h.logger.Info().Str("address", ln.Addr().String()).Msg("start")

	errServe := h.server.Serve(ln)
	if errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
		h.logger.Error().Err(errServe).Msg("serve error")
	}

	h.logger.Info().Msg("stop")
}

func (h *HTTP) test(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

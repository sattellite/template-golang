package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type Service struct {
	logger *zerolog.Logger
	server *http.Server
}

func New(logger *zerolog.Logger) *Service {
	s := &Service{
		logger: logger,
		server: &http.Server{},
	}

	router := chi.NewRouter()
	router.Route("/debug/pprof", func(r chi.Router) {
		r.Get("/profile", pprof.Profile)
		r.Get("/trace", pprof.Trace)
		r.Get("/heap", pprof.Handler("heap").ServeHTTP)
		r.Get("/goroutine", pprof.Handler("goroutine").ServeHTTP)
		r.Get("/allocs", pprof.Handler("allocs").ServeHTTP)
	})

	s.server.Handler = router

	return s
}

func (s *Service) Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, address string) {
	defer wg.Done()

	ln, errLn := net.Listen("tcp", address)
	if errLn != nil {
		s.logger.Error().Err(errLn).Str("address", address).Msg("failed to listen service address")
		cancel()
		return
	}
	defer ln.Close()

	s.logger.Info().Str("address", ln.Addr().String()).Msg("service started")

	go func() {
		s.logger.Info().Str("address", ln.Addr().String()).Msg("serve service server")
		err := s.server.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Err(err).Msg("error serve service server")
			cancel()
		}
	}()

	<-ctx.Done()

	s.logger.Info().Msg("shutdown service server")

	err := s.server.Shutdown(context.Background())
	if err != nil {
		s.logger.Error().Err(err).Msg("error shutdown service server")
	}
}

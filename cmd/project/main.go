package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"template/internal/config"
	"template/internal/database/postgres"
	frontendHTTP "template/internal/frontend/http"
	"template/internal/service"

	"github.com/rs/zerolog"
	"golang.org/x/term"
)

func main() {
	// load config
	cfg, errCfg := config.Load()
	if errCfg != nil {
		panic(errCfg)
	}

	// create global logger
	// default logger with full timestamp
	var writer io.Writer = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC3339
	})
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		// default logger with full timestamp and no color
		writer = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = time.RFC3339
			w.NoColor = true
		})
	}
	if cfg.Format == "json" {
		// json logger
		writer = os.Stdout
	}

	l := zerolog.New(writer).With().Timestamp().Logger()

	if cfg.Trace {
		l = l.Level(zerolog.TraceLevel)
	} else if cfg.Debug {
		l = l.Level(zerolog.DebugLevel)
	} else {
		l = l.Level(zerolog.InfoLevel)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	if errRun := run(ctx, cancel, cfg, &l); errRun != nil {
		log.Printf("error run application, %v", errRun)
		cancel()
		os.Exit(1)
	}

	l.Info().Msg("done")
}

func run(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, logger *zerolog.Logger) error {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	// Services
	srv := service.New(logger)
	wg.Add(1)
	go srv.Run(ctx, cancel, wg, fmt.Sprintf("%s:%d", cfg.Service.Host, cfg.Service.Port))

	// Databases
	db := postgres.New(cfg.Database, logger)

	// Frontend
	front := frontendHTTP.New(cfg.Frontend, db, logger)

	// Run
	wg.Add(1)
	go front.Run(ctx, cancel, wg)

	<-ctx.Done()

	return nil
}

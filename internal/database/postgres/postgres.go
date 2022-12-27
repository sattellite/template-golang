package postgres

import (
	"context"
	"fmt"
	"sync"
	"template/internal/config"
	"time"

	"github.com/rs/zerolog"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	defaultAttempts = 5
)

type dbi interface {
	Close()
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Postgres struct {
	cfg    config.Database
	db     dbi
	logger *zerolog.Logger
}

func New(cfg config.Database, logger *zerolog.Logger) *Postgres {
	l := logger.With().Str("object", "backend").Str("type", "postgres").Str("name", cfg.Name).Logger()
	p := &Postgres{
		cfg:    cfg,
		logger: &l,
	}

	return p
}

func (back *Postgres) Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer cancel()

	db, errDB := pgConnect(ctx, buildPGConnectionString(back.cfg), defaultAttempts, back.logger)
	if errDB != nil {
		back.logger.Error().Err(errDB).Msg("error connect to postgres")
		return
	}
	back.db = db
	defer back.db.Close()

	back.logger.Info().Msg("start")

	<-ctx.Done()

	back.logger.Info().Msg("stop")
}

func pgConnect(ctx context.Context, connString string, attempts int, logger *zerolog.Logger) (*pgxpool.Pool, error) {
	interval := time.Second
	for {
		db, err := pgxpool.Connect(ctx, connString)
		if err == nil {
			return db, nil
		}

		attempts--
		if attempts <= 0 {
			return nil, fmt.Errorf("no connection, 0 attempts")
		}

		logger.Error().Err(err).Dur("wait interval", interval).Int("attempts", attempts).Msg("error connect to postgres, try again")

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
			interval *= 2
		}
	}
}

func buildPGConnectionString(cfg config.Database) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&sslrootcert=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.SSLCertPath)
}

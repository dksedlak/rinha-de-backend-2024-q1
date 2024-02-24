package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	"github.com/dksedlak/rinha-de-backend-2024-q1/internal/httpd"
	"github.com/dksedlak/rinha-de-backend-2024-q1/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Logger.Level(zerolog.TraceLevel).With().Int("pid", os.Getpid()).Logger()
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"

	ctx, ctxCancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGQUIT,
	)
	defer ctxCancel()

	ctx = log.Logger.WithContext(ctx)

	log.Ctx(ctx).Info().Msg("starting service ...")
	log.Ctx(ctx).Info().Msg("loading configurations ...")

	cfg, err := internal.LoadConfig()
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("error trying to load env variables")
		return
	}

	log.Info().Msg("connection to the database (PostgreSQL) ...")

	//nolint:gomnd
	pgRepository, err := repository.NewPostgreSQL(ctx, repository.PgConfig{
		ConnMaxLifetime: 2 * time.Second,
		MaxIdleTime:     2 * time.Second,
		MaxOpenConns:    100,
		DSN:             cfg.PostgresDSN,
	})
	if err != nil {
		log.Error().AnErr("error", err).Msg("failed to create postgres client")
	}

	server := httpd.NewServer(ctx, cfg.HTTPAddr, pgRepository)

	// starts HTTP server
	server.Run(ctxCancel)

	// waiting until receive any signal
	<-ctx.Done()

	// shutdown gracefully
	server.Close()
}

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	_ "github.com/jackc/pgx/v5/stdlib" // adds the pgx driver
	"github.com/jmoiron/sqlx"
)

type PgConfig struct {
	DSN             string
	RetryMaxTries   uint64
	MaxOpenConns    int
	MaxIdleConns    int
	RetryInterval   time.Duration
	ConnMaxLifetime time.Duration
}

type PostgreSQL struct {
	db *sqlx.DB
}
type Transaction struct {
	Value       uint64
	Type        string
	Description string
	CreatedAt   time.Time
}

type BankStatement struct {
	Total            uint64
	Date             time.Time
	Limit            uint64
	LastTransactions []byte
}

func NewPostgreSQL(ctx context.Context, cfg PgConfig) (*PostgreSQL, error) {
	db, err := sqlx.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database DSN: %w", err)
	}

	err = backoff.Retry(func() error {
		return db.PingContext(ctx)
	}, backoff.WithMaxRetries(
		backoff.WithContext(backoff.NewConstantBackOff(cfg.RetryInterval), ctx),
		cfg.RetryMaxTries,
	))

	if err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return &PostgreSQL{
		db: db,
	}, nil
}

func (pg *PostgreSQL) GetDBInstance() *sqlx.DB {
	return pg.db
}

func (pg *PostgreSQL) AddNewTransaction(ctx context.Context, clientID string, transaction internal.Transaction) error {
	return nil
}

func (pg *PostgreSQL) GetBankStatements(ctx context.Context, clientId string) (*internal.BankStatement, error) {
	return nil, nil
}

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	pgType "github.com/jackc/pgx/pgtype/ext/gofrs-uuid"
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

type Balance struct {
	ClientID   int32       `db:"client_id"`
	Limit      uint64      `db:"balance_limit"`
	Amount     uint64      `db:"amount"`
	LastCommit pgType.UUID `db:"last_commit"`
}

type TransactionObject struct {
	TransactionType string    `json:"transaction_type"`
	Description     string    `json:"description"`
	Value           uint64    `json:"value"`
	CreatedAt       time.Time `json:"created_at"`
}

type TransactionRow struct {
	ClientID         int32       `db:"client_id"`
	LastTransactions []byte      `db:"last_transactions"`
	LastCommit       pgType.UUID `db:"last_commit"`
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

func (pg *PostgreSQL) CreateTransaction(ctx context.Context, clientID int, transaction internal.Transaction) error {

	return nil
}

func (pg *PostgreSQL) GetBankStatements(ctx context.Context, clientID int) (*internal.BankStatement, error) {
	return nil, nil
}

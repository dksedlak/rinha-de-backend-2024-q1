package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	_ "github.com/jackc/pgx/v5/stdlib" // adds the pgx driver
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

var (
	ErrConcurrencyRequest = errors.New("concurrency updating the row")
	ErrInsufficientLimit  = errors.New("insufficient limit")
	ErrNotFound           = errors.New("not found")
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

type balance struct {
	ClientID         int
	Limit            int64
	Amount           int64
	LastTransactions []TransactionObject
}

type TransactionObject struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Value       int64     `json:"value"`
	CreatedAt   time.Time `json:"created_at"`
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

func (pg *PostgreSQL) getClientBalance(ctx context.Context, clientID int) (*balance, error) {
	query := fmt.Sprintf(
		`SELECT client_id, client_limit, amount, last_transactions FROM balances WHERE client_id = %d;`,
		clientID,
	)

	var record balance
	var rawTransactions []byte

	if err := pg.db.QueryRowContext(ctx, query).Scan(
		&record.ClientID,
		&record.Limit,
		&record.Amount,
		&rawTransactions,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to select balance account: %w", err)
	}

	var transactions []TransactionObject
	if err := json.Unmarshal(rawTransactions, &transactions); err != nil {
		log.Err(err).Msg("could not parse the transactions from database")
		return nil, err
	}

	record.LastTransactions = transactions
	return &record, nil
}

func (pg *PostgreSQL) AddNewTransaction(ctx context.Context, clientID int, transaction internal.Transaction) (*internal.Resume, error) {
	tx, err := pg.db.BeginTx(ctx, &sql.TxOptions{
		//Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		log.Err(err).Msg("failed to begin tx")
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	var rawTransactions []byte
	var limit int64
	var amount int64

	query := fmt.Sprintf(`SELECT client_limit, amount, last_transactions FROM balances WHERE client_id = %d;`, clientID)

	if err := tx.QueryRowContext(ctx, query).Scan(
		&limit,
		&amount,
		&rawTransactions,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query the row: %w", err)
	}

	var lastTransactions []TransactionObject
	if err := json.Unmarshal(rawTransactions, &lastTransactions); err != nil {
		log.Err(err).Msg("could not parse the transactions from database")
		return nil, err
	}

	if transaction.Type == internal.TransactionCredit {
		amount += transaction.Value
	} else {
		amount -= transaction.Value
	}

	if (-1 * amount) > limit {
		return nil, ErrInsufficientLimit
	}

	if len(lastTransactions) >= 10 {
		tmp := lastTransactions[1:]
		lastTransactions = append(tmp, TransactionObject{
			Type:        string(transaction.Type),
			Description: transaction.Description,
			Value:       transaction.Value,
			CreatedAt:   transaction.CreatedAt,
		})
	} else {
		lastTransactions = append(lastTransactions, TransactionObject{
			Type:        string(transaction.Type),
			Description: transaction.Description,
			Value:       transaction.Value,
			CreatedAt:   transaction.CreatedAt,
		})
	}

	query = `UPDATE balances SET amount = $1, last_transactions = $2 WHERE client_id = $3`
	if _, err := tx.ExecContext(ctx, query,
		amount,
		lastTransactions,
		clientID,
	); err != nil {
		return nil, ErrConcurrencyRequest
	}

	return &internal.Resume{
		Amount: amount,
		Limit:  limit,
	}, tx.Commit()
}

func (pg *PostgreSQL) GetBankStatements(ctx context.Context, clientID int) (*internal.BankStatement, error) {
	balance, err := pg.getClientBalance(ctx, clientID)
	if err != nil {
		return nil, err
	}

	transactions := make([]internal.Transaction, 0, len(balance.LastTransactions))

	for _, tmp := range balance.LastTransactions {
		transactions = append(transactions, internal.Transaction{
			Value:       tmp.Value,
			Type:        internal.TransactionType(tmp.Type),
			Description: tmp.Description,
			CreatedAt:   tmp.CreatedAt,
		})
	}

	return &internal.BankStatement{
		Amount:           balance.Amount,
		Date:             time.Now(),
		Limit:            balance.Limit,
		LastTransactions: transactions,
	}, nil
}

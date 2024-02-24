package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	_ "github.com/jackc/pgx/v5/stdlib" // adds the pgx driver
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrConcurrencyRequest = errors.New("concurrency the last_commit value is different")
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

type Account struct {
	ID               int
	CreditLimit      int64
	Balance          int64
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

func (pg *PostgreSQL) getAccount(ctx context.Context, clientID int) (*Account, error) {
	query := "SELECT id, credit_limit, balance, last_transactions FROM accounts WHERE id = $1"

	var record Account
	var rawTransactions pq.StringArray

	if err := pg.db.QueryRowContext(ctx, query, clientID).Scan(
		&record.ID,
		&record.CreditLimit,
		&record.Balance,
		&rawTransactions,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to select balance account: %w", err)
	}

	transactions := make([]TransactionObject, 0, len(rawTransactions))

	var tmp TransactionObject
	for _, current := range rawTransactions {
		if err := json.Unmarshal([]byte(current), &tmp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal LastTransactions: %w", err)
		}
		transactions = append(transactions, TransactionObject{
			Type:        tmp.Type,
			Description: tmp.Description,
			Value:       tmp.Value,
			CreatedAt:   tmp.CreatedAt,
		})
	}

	record.LastTransactions = transactions
	return &record, nil
}

func (pg *PostgreSQL) AddNewTransaction(ctx context.Context, clientID int, transaction internal.Transaction) (*internal.Resume, error) {
	var operator string
	if transaction.Type == internal.TransactionCredit {
		operator = "+"
	} else {
		operator = "-"
	}

	var creditLimit, balance int64

	tx, err := pg.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
	})
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		WITH updated_data AS (
			SELECT id, credit_limit, balance, last_transactions
			FROM accounts
			WHERE id = $1
			FOR UPDATE
		)
		UPDATE accounts
		SET balance = updated_data.balance %s $2,
		last_transactions =
			CASE WHEN array_length(updated_data.last_transactions, 1) > 10
				THEN COALESCE(updated_data.last_transactions[2:array_length(updated_data.last_transactions, 1)], ARRAY[]::JSON[]) || ARRAY[$3]::JSON[]
				ELSE COALESCE(updated_data.last_transactions, ARRAY[]::JSON[]) || ARRAY[$3]::JSON[]
			END
		FROM updated_data
		WHERE accounts.id = $1
		RETURNING accounts.credit_limit, accounts.balance;
	`, operator)

	row := tx.QueryRowContext(ctx, query, clientID, transaction.Value, TransactionObject{
		Type:        string(transaction.Type),
		Description: transaction.Description,
		Value:       transaction.Value,
		CreatedAt:   transaction.CreatedAt,
	})
	if row.Err() != nil {
		tx.Rollback()
		return nil, ErrInsufficientLimit
	}

	if err := row.Scan(
		&creditLimit,
		&balance,
	); err != nil {
		return nil, ErrNotFound
	}

	tx.Commit()

	return &internal.Resume{
		Amount: balance,
		Limit:  creditLimit,
	}, nil
}

func (pg *PostgreSQL) GetBankStatements(ctx context.Context, clientID int) (*internal.BankStatement, error) {
	account, err := pg.getAccount(ctx, clientID)
	if err != nil {
		return nil, err
	}

	transactions := make([]internal.Transaction, 0, len(account.LastTransactions))

	size := len(account.LastTransactions)
	for idx := size - 1; idx >= 0; idx-- {
		transactions = append(transactions, internal.Transaction{
			Value:       account.LastTransactions[idx].Value,
			Type:        internal.TransactionType(account.LastTransactions[idx].Type),
			Description: account.LastTransactions[idx].Description,
			CreatedAt:   account.LastTransactions[idx].CreatedAt,
		})
	}

	return &internal.BankStatement{
		Amount:           account.Balance,
		Date:             time.Now(),
		Limit:            account.CreditLimit,
		LastTransactions: transactions,
	}, nil
}

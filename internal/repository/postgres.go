package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficientLimit = errors.New("insufficient limit")
	ErrNotFound          = errors.New("not found")
)

type PostgreSQL struct {
	db *pgxpool.Pool
}

type Account struct {
	ID               int                 `json:"id"`
	CreditLimit      int64               `json:"credit_limit"`
	Balance          int64               `json:"balance"`
	LastTransactions []TransactionObject `json:"last_transactions"`
}

type TransactionObject struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Value       int64     `json:"value"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewPostgreSQL(ctx context.Context, PostgresDSN string) (*PostgreSQL, error) {
	db, err := pgxpool.New(ctx, PostgresDSN)
	if err != nil {
		return nil, err
	}

	return &PostgreSQL{
		db: db,
	}, nil
}

func (pg *PostgreSQL) GetDBInstance() *pgxpool.Pool {
	return pg.db
}

func (pg *PostgreSQL) getAccount(ctx context.Context, accountID int) (*Account, error) {
	row := pg.db.QueryRow(ctx, `
        SELECT
            id,
            credit_limit,
            balance,
            COALESCE(json_agg(last_transactions), '[]'::json) as last_transactions
        FROM
            accounts
        WHERE
            id = $1
        GROUP BY
            id, credit_limit, balance
    `, accountID)

	var account Account
	var lastTransactionsJSON string

	if err := row.Scan(&account.ID, &account.CreditLimit, &account.Balance, &lastTransactionsJSON); err != nil {
		return nil, ErrNotFound
	}

	// Unmarshal the JSON array into a slice of slices
	var nestedTransactions [][]TransactionObject
	if err := json.Unmarshal([]byte(lastTransactionsJSON), &nestedTransactions); err != nil {
		return nil, ErrNotFound
	}

	// Flatten the LastTransactions slice
	for _, transactions := range nestedTransactions {
		account.LastTransactions = append(account.LastTransactions, transactions...)
	}

	return &account, nil
}

func (pg *PostgreSQL) AddNewTransaction(ctx context.Context, accountID int, transaction internal.Transaction) (*internal.Resume, error) {
	var operator string
	if transaction.Type == internal.TransactionCredit {
		operator = "+"
	} else {
		operator = "-"
	}

	var creditLimit, balance int64

	tx, err := pg.db.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	// Update the account
	row := tx.QueryRow(context.Background(),
		fmt.Sprintf(`UPDATE accounts
		SET balance = balance %s $2,
			last_transactions =
				CASE WHEN array_length(last_transactions, 1) > 10
					THEN COALESCE(last_transactions[2:array_length(last_transactions, 1)], ARRAY[]::JSON[]) || ARRAY[$3]::JSON[]
					ELSE COALESCE(last_transactions, ARRAY[]::JSON[]) || ARRAY[$3]::JSON[]
				END
		WHERE id = $1
		RETURNING credit_limit, balance`, operator), accountID, transaction.Value, TransactionObject{
			Type:        string(transaction.Type),
			Description: transaction.Description,
			Value:       transaction.Value,
			CreatedAt:   transaction.CreatedAt,
		})

	if err := row.Scan(&creditLimit, &balance); err != nil {
		return nil, err
	}

	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}

	return &internal.Resume{
		Amount: balance,
		Limit:  creditLimit,
	}, nil
}

func (pg *PostgreSQL) GetBankStatements(ctx context.Context, accountID int) (*internal.BankStatement, error) {
	account, err := pg.getAccount(ctx, accountID)
	if err != nil {
		return nil, ErrNotFound
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

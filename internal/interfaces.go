package internal

import (
	"context"
	"net/http"
)

type Handler interface {
	AddNewTransaction(w http.ResponseWriter, r *http.Request)
	GetStatements(w http.ResponseWriter, r *http.Request)
}

type Repository interface {
	AddNewTransaction(ctx context.Context, clientID int, transaction Transaction) (*Resume, error)
	GetBankStatements(ctx context.Context, clientID int) (*BankStatement, error)
}

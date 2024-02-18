package internal

import (
	"context"
	"net/http"
)

type Handler interface {
	CreateTransaction(w http.ResponseWriter, r *http.Request)
	GetStatements(w http.ResponseWriter, r *http.Request)
}

type Repository interface {
	CreateTransaction(ctx context.Context, clientID int, transaction Transaction) error
	GetBankStatements(ctx context.Context, clientID int) (*BankStatement, error)
}

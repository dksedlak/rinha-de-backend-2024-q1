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
	AddNewTransaction(ctx context.Context, clientID string, transaction Transaction) error
	GetBankStatements(ctx context.Context, clientId string) (*BankStatement, error)
}

package handler

import (
	"context"
	"net/http"

	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
)

type handler struct {
	repository internal.Repository
}

func NewHandler(ctx context.Context, repository internal.Repository) *handler {
	return &handler{
		repository: repository,
	}
}

func (h *handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {}

func (h *handler) GetStatements(w http.ResponseWriter, r *http.Request) {}

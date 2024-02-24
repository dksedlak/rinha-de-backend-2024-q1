package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	"github.com/dksedlak/rinha-de-backend-2024-q1/internal/repository"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type handler struct {
	repository internal.Repository
}

func NewHandler(ctx context.Context, repository internal.Repository) *handler {
	return &handler{
		repository: repository,
	}
}

type AddNewTransactionRequest struct {
	Value       int64  `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
}

type AddNewTransactionResponse struct {
	Limit  int64 `json:"limite"`
	Amount int64 `json:"saldo"`
}

type TransactionsResponse struct {
	Value       int64  `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
	CreatedAt   string `json:"realizada_em"`
}

type AmountResponse struct {
	Total       int64  `json:"total"`
	CurrentDate string `json:"data_extrato"`
	Limit       int64  `json:"limite"`
}

type GetStatementsResponse struct {
	Amount           AmountResponse         `json:"saldo"`
	LastTransactions []TransactionsResponse `json:"ultimas_transacoes"`
}

func (h *handler) AddNewTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString, ok := mux.Vars(r)["id"]

	if !ok {
		log.Error().Msg("wrong URI")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	clientID, err := strconv.Atoi(idString)
	if err != nil {
		log.Error().Msg("the client id needs to be an integer")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var payload AddNewTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Error().Msg("cannot parse the request body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// validate the payload
	if err := validatePayload(payload); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	resume, err := h.repository.AddNewTransaction(ctx, clientID, internal.Transaction{
		Value:       payload.Value,
		Type:        internal.TransactionType(payload.Type),
		Description: payload.Description,
		CreatedAt:   time.Now(),
	})

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(AddNewTransactionResponse{
		Limit:  resume.Limit,
		Amount: resume.Amount,
	}); err != nil {
		log.Err(err).Msg("failed to encode the response body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) GetStatements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString, ok := mux.Vars(r)["id"]

	if !ok {
		log.Error().Msg("wrong URI")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	clientID, err := strconv.Atoi(idString)
	if err != nil {
		log.Error().Msg("the client id needs to be an integer")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	bankStatements, err := h.repository.GetBankStatements(ctx, clientID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(GetStatementsResponse{
		Amount: AmountResponse{
			Total:       bankStatements.Amount,
			Limit:       bankStatements.Limit,
			CurrentDate: bankStatements.Date.Format(time.RFC3339Nano),
		},
		LastTransactions: mapTransactionsToResponse(bankStatements.LastTransactions),
	}); err != nil {
		log.Err(err).Msg("failed to encode the response body")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

}

func validatePayload(payload AddNewTransactionRequest) error {
	if payload.Value < 0 {
		return fmt.Errorf("invalid value: %d", payload.Value)
	}

	if payload.Type != "c" && payload.Type != "d" {
		return fmt.Errorf("invalid type: %s", payload.Type)
	}

	if len(payload.Description) < 1 || len(payload.Description) > 10 {
		return fmt.Errorf("invalid description: %s | length: %d", payload.Description, len(payload.Description))
	}

	return nil
}

func mapTransactionsToResponse(value []internal.Transaction) []TransactionsResponse {
	result := make([]TransactionsResponse, 0, len(value))
	for _, current := range value {
		result = append(result, TransactionsResponse{
			Value:       current.Value,
			Type:        string(current.Type),
			Description: current.Description,
			CreatedAt:   current.CreatedAt.Format(time.RFC3339Nano),
		})
	}
	return result
}

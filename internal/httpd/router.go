package httpd

import (
	"net/http"

	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	"github.com/gorilla/mux"
)

type route struct {
	Name    string
	Method  string
	Path    string
	Handler http.HandlerFunc
}

func mapEndpoints(handler internal.Handler) []route {
	return []route{
		{
			Name:    "CreateTransaction",
			Method:  http.MethodPost,
			Path:    "/clientes/{id}/transacoes",
			Handler: handler.CreateTransaction,
		},
		{
			Name:    "GetStatements",
			Method:  http.MethodGet,
			Path:    "/clientes/{id}/extrato",
			Handler: handler.GetStatements,
		},
	}
}

func initRoutes(router *mux.Router, handler internal.Handler) {
	for _, route := range mapEndpoints(handler) {
		router.
			Name(route.Name).
			Methods(route.Method).
			Path(route.Path).
			Handler(route.Handler)
	}
}

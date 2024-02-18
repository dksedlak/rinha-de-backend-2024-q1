package httpd

import (
	"context"
	"net/http"
	"time"

	"github.com/dksedlak/rinha-de-backend-2024-q1/internal"
	"github.com/dksedlak/rinha-de-backend-2024-q1/internal/handler"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type server struct {
	port       string
	router     *mux.Router
	httpServer *http.Server
	handler    internal.Handler
}

const (
	maxRequestSize    = 5 * 1024 // 5 KB
	maxHeaderSize     = 1024     // 1 KB
	readHeaderTimeout = 2 * time.Second
	shutdownTimeout   = 2 * time.Second
	readTimeout       = 2 * time.Second
	writeTimeout      = 2 * time.Second
)

func NewServer(ctx context.Context, port string, database internal.Repository) *server {
	router := mux.NewRouter()

	// create HTTP handlers
	handler := handler.NewHandler(ctx, database)

	// init routes
	initRoutes(router, handler)

	return &server{
		port:    port,
		router:  router,
		handler: handler,
		httpServer: &http.Server{
			Addr:              ":" + port,
			MaxHeaderBytes:    maxHeaderSize,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			Handler:           http.MaxBytesHandler(router, maxRequestSize),
		},
	}
}

func (s *server) Run(ctxCancel context.CancelFunc) {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			log.Err(err).Msg("http server has been stopped")
		}

		ctxCancel()
	}()
}

func (s *server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Err(err).Msg("failed to shutdown the HTTP server")
	}

	log.Info().Msg("http server was gracefully shutdown")
}

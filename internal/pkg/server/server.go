package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jirevwe/cascade/internal/pkg/util"
	"github.com/jirevwe/cascade/internal/pkg/version"

	log "github.com/sirupsen/logrus"
)

var (
	httpTimeOut = time.Second * 10
)

type Server struct {
	srv *http.Server
}

func NewServer(port uint32) *Server {
	srv := &Server{
		srv: &http.Server{
			ReadTimeout:  httpTimeOut,
			WriteTimeout: httpTimeOut,
			Addr:         fmt.Sprintf(":%d", port),
		},
	}

	return srv
}

func (s *Server) SetHandler(handler http.Handler) {
	router := chi.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = render.Render(w, r, util.NewServerResponse(fmt.Sprintf("cascade %v", version.GetVersion()), nil, http.StatusOK))
	})

	router.Handle("/*", handler)
	s.srv.Handler = router
}

func (s *Server) Listen() {
	go func() {
		//service connections
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("failed to listen")
		}
	}()

	s.gracefulShutdown()
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) {
	go func() {
		//service connections
		if err := s.srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("failed to listen")
		}
	}()

	s.gracefulShutdown()
}

func (s *Server) gracefulShutdown() {
	//Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Info("Stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeOut)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("Server Shutdown")
	}

	log.Info("Server exiting")

	time.Sleep(2 * time.Second) // allow all pending connections close themselves
}

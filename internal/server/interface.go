package server

import (
	"context"
	"net/http"
)

type IServer interface {
	authMiddleware(next http.HandlerFunc) http.HandlerFunc
	createTeamHandler(w http.ResponseWriter, r *http.Request)
	getTeamHandler(w http.ResponseWriter, r *http.Request)
	loginHandler(w http.ResponseWriter, r *http.Request)
	setupRoutes()
	RunServer(ctx context.Context) error
	gracefulShutdown(server *http.Server) error
	writeError(w http.ResponseWriter, code, msg string, status int)
}

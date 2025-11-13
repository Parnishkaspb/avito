package server

import (
	"context"
	"net/http"
)

type IServer interface {
	Run(ctx context.Context) error
	createTeamHandler(w http.ResponseWriter, r *http.Request)
}

package utilapi

import (
	"log/slog"
	"net/http"
	"time"
)

type Router struct {
	mux *http.ServeMux
	log *slog.Logger
	sli time.Duration
}

func NewRouter(log *slog.Logger, sli time.Duration) *Router {
	return &Router{
		mux: http.NewServeMux(),
		log: log,
		sli: sli,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

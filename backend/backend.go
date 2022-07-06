package backend

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/ringtail/lucas/backend/handlers"
	"github.com/ringtail/lucas/backend/types"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type LucasServer struct {
	Handler *mux.Router
}

func (ls *LucasServer) Start(opts *types.Opts) {
	log.Info("LucasServer starting")
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.IndexHandler)
	mux.HandleFunc("/store", handlers.StoreHandler)
	contextMux := ls.Middleware(opts, mux)
	port := opts.Port
	log.Info("port", port)
	log.Fatal(http.ListenAndServe(":"+port, contextMux))
}

func (ls *LucasServer) Middleware(opts *types.Opts, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "opts", opts)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

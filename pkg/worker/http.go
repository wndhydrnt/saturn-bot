package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"go.uber.org/zap"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	const up = "UP"
	_, _ = fmt.Fprint(w, up)
}

type httpServer struct {
	addr   string
	mux    *http.ServeMux
	server *http.Server
}

func (h *httpServer) handle(pattern string, handler http.Handler) {
	if h.mux == nil {
		h.mux = http.NewServeMux()
	}

	h.mux.Handle(pattern, handler)
}

func (h *httpServer) start() {
	if h.addr == "" {
		h.addr = ":3036"
	}

	if h.mux == nil {
		h.mux = http.NewServeMux()
	}

	h.server = &http.Server{
		Addr:    h.addr,
		Handler: h.mux,
	}

	if err := h.server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Log().Errorw("HTTP server failed", zap.Error(err))
		}
	}
}

func (h *httpServer) stop() {
	if h.server == nil {
		return
	}

	err := h.server.Shutdown(context.Background())
	if err != nil {
		log.Log().Errorw("Shutdown of HTTP server failed", zap.Error(err))
	}
}

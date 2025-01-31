package worker

import (
	"fmt"
	"net/http"
	"time"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	const up = "UP"
	_, _ = fmt.Fprint(w, up)
}

func newHttpServer(addr string, mux http.Handler) *http.Server {
	if addr == "" {
		addr = ":3036"
	}

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Millisecond,
	}
}

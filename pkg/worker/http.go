package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/client"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/version"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	const up = "UP"
	_, _ = fmt.Fprint(w, up)
}

// infoResponse defines useful information about the worker.
type infoResponse struct {
	// The list of tasks loaded by the worker.
	Tasks   []infoResponseTask  `json:"tasks"`
	Version version.VersionInfo `json:"version"`
}

type infoResponseTask struct {
	Checksum string `json:"checksum"`
	Path     string `json:"path"`
	Task     string `json:"task"`
}

// infoHandler returns information about the worker as JSON via HTTP.
func infoHandler(worker *Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := infoResponse{
			Version: version.Info,
		}
		for _, workerTask := range worker.tasks {
			resp.Tasks = append(resp.Tasks, infoResponseTask{Path: workerTask.Path(), Checksum: workerTask.Checksum(), Task: workerTask.Name})
		}

		err := json.NewEncoder(w).Encode(&resp)
		if err != nil {
			log.Log().Errorf("Write task info to writer: %v", err)
		}
	}
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

func newApiKeyAddFunc(key, value string) client.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set(key, value)
		return nil
	}
}

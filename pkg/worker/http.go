package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	Path   string `json:"path"`
	Sha256 string `json:"sha256"`
	Task   string `json:"task"`
}

// infoHandler returns information about the worker as JSON via HTTP.
func infoHandler(worker *Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := infoResponse{
			Version: version.Info,
		}
		for _, workerTask := range worker.tasks {
			resp.Tasks = append(resp.Tasks, infoResponseTask{Path: workerTask.Path, Sha256: workerTask.Sha256, Task: workerTask.Task.Name})
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

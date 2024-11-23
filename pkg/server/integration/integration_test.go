package integration_test

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	_ "github.com/ncruces/go-sqlite3/vfs/memdb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"gopkg.in/yaml.v3"
)

var (
	defaultServerConfig = config.Configuration{
		GithubToken:               ptr.To("unittest"),
		ServerGithubWebhookSecret: "secret",
		ServerGitlabWebhookSecret: "secret",
	}
	defaultTask              = schema.Task{Name: "unittest"}
	defaultTaskContentBase64 = "bmFtZTogdW5pdHRlc3QK"
	defaultTaskHash          = "e42a6e186f31b860f22f07ed468b99c6dc75318542fc9ac8383358fae1b5ab8b"
)

type apiCall struct {
	// HTTP method.
	method string
	// Path of the request.
	path string
	// Request headers, if any.
	requestHeaders map[string]string
	// Request body, if any.
	requestBody any
	// Expected status code of the response.
	statusCode int
	// Expected body of the response.
	responseBody any
	// Amount of time to sleep before sending the request.
	// Makes it easier to test race conditions between goroutines.
	sleep time.Duration
}

type testCase struct {
	name     string
	config   *config.Configuration
	tasks    []schema.Task
	apiCalls []apiCall
}

func executeTestCase(t *testing.T, tc testCase) {
	cfg := defaultServerConfig
	if tc.config != nil {
		cfg = ptr.From(tc.config)
	}
	port := randomPort()
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	cfg.ServerAddr = addr
	cfg.ServerBaseUrl = "http://" + addr

	dbDir, err := os.MkdirTemp("", "")
	require.NoError(t, err, "Creates database directory")
	cfg.ServerDatabasePath = filepath.Join(dbDir, "saturn-bot.db")
	t.Logf("Path to database: %s", cfg.ServerDatabasePath)

	opts, err := options.ToOptions(cfg)
	require.NoError(t, err, "Parses options")
	// Always use a new registry to avoid a panic caused by attempts to register the same metrics twice.
	promReg := prometheus.NewRegistry()
	opts.SetPrometheusRegistry(promReg)

	taskFiles := bootstrapTaskFiles(t, tc.tasks)

	server := &server.Server{}
	err = server.Start(opts, taskFiles)
	require.NoError(t, err, "Server starts up")
	defer func() {
		err := server.Stop()
		require.NoError(t, err, "Server shuts down")
	}()

	// Give the HTTP server time to start up - works around flaky tests
	time.Sleep(1 * time.Millisecond)
	e := httpexpect.Default(t, cfg.ServerBaseUrl)
	for _, call := range tc.apiCalls {
		time.Sleep(call.sleep)
		req := e.Request(call.method, call.path).
			WithHeaders(call.requestHeaders)
		if call.requestBody != nil {
			req = req.WithJSON(call.requestBody)
		}

		resp := req.Expect().Status(call.statusCode)
		if call.responseBody == nil {
			resp.NoContent()
		} else {
			resp.JSON().Object().IsEqual(call.responseBody)
		}
	}
}

func randomPort() int {
	for {
		r := rand.New(rand.NewSource(time.Now().UnixMicro())) //nolint:all // It's okay to use weak randomization in unit tests
		port := r.Intn(25000)
		if port > 1024 {
			return port
		}
	}
}

func bootstrapTaskFiles(t *testing.T, tasks []schema.Task) []string {
	var filePaths []string
	for _, task := range tasks {
		f, err := os.CreateTemp("", "*.yaml")
		require.NoErrorf(t, err, "Creates temporary task file for task %s", task.Name)
		defer f.Close()
		enc := yaml.NewEncoder(f)
		enc.SetIndent(2)
		err = enc.Encode(task)
		require.NoErrorf(t, err, "Writes task %s to temporary file", task.Name)
		filePaths = append(filePaths, f.Name())
	}

	return filePaths
}

package integration_test

import (
	"errors"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func Test_Shutdown_WaitForRunningRuns(t *testing.T) {
	task1 := schema.Task{Name: "Task1"}
	task2 := schema.Task{Name: "Task2"}
	taskFiles := bootstrapTaskFiles(t, task1, task2)
	opts := setupOptions(t, nil, nil)
	opts.ServerShutdownTimeout = 10 * time.Second
	svr := &server.Server{}
	err := svr.Start(opts, taskFiles)
	require.NoError(t, err, "server starts")

	httpExpect := httpexpect.Default(t, opts.Config.ServerBaseUrl)
	// Schedule a first run
	assertApiCall(httpExpect, apiCall{
		sleep:  5 * time.Millisecond, // Wait for HTTP server to start up
		method: "POST",
		path:   "/api/v1/runs",
		requestBody: openapi.ScheduleRunV1Request{
			TaskName: task1.Name,
		},
		statusCode: http.StatusOK,
		responseBody: openapi.ScheduleRunV1Response{
			RunID: 1,
		},
	})

	// Schedule a second run
	assertApiCall(httpExpect, apiCall{
		method: "POST",
		path:   "/api/v1/runs",
		requestBody: openapi.ScheduleRunV1Request{
			TaskName: task2.Name,
		},
		statusCode: http.StatusOK,
		responseBody: openapi.ScheduleRunV1Response{
			RunID: 2,
		},
	})

	// Start work on the first run.
	// Puts it into the "running" state.
	assertApiCall(httpExpect, apiCall{
		method:     "GET",
		path:       "/api/v1/worker/work",
		statusCode: http.StatusOK,
		responseBody: openapi.GetWorkV1Response{
			RunID: 1,
			Task:  openapi.WorkTaskV1{Name: task1.Name, Hash: "8e6b6f1b27681d3bb6a30bbb92c82c9cea6cf1acdd52ca483b7e670dfff7ffab"},
		},
	})

	// Shut down the server
	go func() {
		err := svr.Stop()
		require.NoError(t, err, "server stops")
	}()

	// Try to get the second run.
	// Should not return the next run because the server is shutting down.
	assertApiCall(httpExpect, apiCall{
		sleep:        5 * time.Millisecond,
		method:       "GET",
		path:         "/api/v1/worker/work",
		statusCode:   http.StatusOK,
		responseBody: openapi.GetWorkV1Response{},
	})

	// Report the result of the active run.
	assertApiCall(httpExpect, apiCall{
		method: "POST",
		path:   "/api/v1/worker/work",
		requestBody: openapi.ReportWorkV1Request{
			RunID: 1,
			Task: openapi.WorkTaskV1{
				Name: task1.Name,
			},
			TaskResults: []openapi.ReportWorkV1TaskResult{},
		},
		statusCode: http.StatusCreated,
		responseBody: openapi.ReportWorkV1Response{
			Result: "ok",
		},
	})

	require.Eventually(
		t,
		func() bool {
			req, err := http.NewRequest(http.MethodGet, opts.Config.ServerBaseUrl+"/api/v1/worker/work", nil)
			require.NoError(t, err)
			req.Header.Set(openapi.HeaderApiKey, testApiKey)
			_, err = http.DefaultClient.Do(req)
			return errors.Is(err, syscall.ECONNREFUSED)
		},
		20*time.Second, // API server shutdown timeout + http server shutdown timeout
		200*time.Millisecond,
		"Server stops eventually and connect fails",
	)
}

func Test_Shutdown_MarkLateRunsAsFailed(t *testing.T) {
	task := schema.Task{Name: "Task1"}
	taskFiles := bootstrapTaskFiles(t, task)
	opts := setupOptions(t, nil, nil)
	opts.ServerShutdownCheckInterval = 1 * time.Nanosecond
	svrFirst := &server.Server{}
	err := svrFirst.Start(opts, taskFiles)
	require.NoError(t, err, "server starts for the first time")

	httpExpect := httpexpect.Default(t, opts.Config.ServerBaseUrl)
	// Schedule a first run
	assertApiCall(httpExpect, apiCall{
		sleep:  5 * time.Millisecond, // Wait for HTTP server to start up
		method: "POST",
		path:   "/api/v1/runs",
		requestBody: openapi.ScheduleRunV1Request{
			TaskName: task.Name,
		},
		statusCode: http.StatusOK,
		responseBody: openapi.ScheduleRunV1Response{
			RunID: 1,
		},
	})

	// Start work on the run.
	// Puts it into the "running" state.
	assertApiCall(httpExpect, apiCall{
		method:     "GET",
		path:       "/api/v1/worker/work",
		statusCode: http.StatusOK,
		responseBody: openapi.GetWorkV1Response{
			RunID: 1,
			Task:  openapi.WorkTaskV1{Name: task.Name, Hash: "8e6b6f1b27681d3bb6a30bbb92c82c9cea6cf1acdd52ca483b7e670dfff7ffab"},
		},
	})

	// Shut down the server
	err = svrFirst.Stop()
	require.NoError(t, err, "server stops for the first time")

	promReg := prometheus.NewRegistry()
	opts.SetPrometheusRegistry(promReg)
	// Start the server again to check the result
	svrSecond := &server.Server{}
	err = svrSecond.Start(opts, taskFiles)
	require.NoError(t, err, "server starts for the second time")

	// Get and compare the status of the run.
	assertApiCall(httpExpect, apiCall{
		sleep:      5 * time.Millisecond,
		method:     "GET",
		path:       "/api/v1/runs/1",
		statusCode: http.StatusOK,
		responseBody: openapi.GetRunV1Response{
			Run: openapi.RunV1{
				Error:         ptr.To("Run failed to report before shutdown"),
				FinishedAt:    ptr.To(testDate(1, 0, 0, 4)),
				Id:            1,
				Reason:        openapi.Manual,
				ScheduleAfter: testDate(1, 0, 0, 1),
				StartedAt:     ptr.To(testDate(1, 0, 0, 3)),
				Status:        openapi.Failed,
				Task:          task.Name,
			},
		},
	})

	// Shut down the server
	err = svrSecond.Stop()
	require.NoError(t, err, "server stops for the second time")
}

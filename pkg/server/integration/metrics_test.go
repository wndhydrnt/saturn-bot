package integration_test

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func Test_Metrics(t *testing.T) {
	opts := setupOptions(t, nil, nil)

	taskWithMetricLabels := schema.Task{Name: "metric-labels", MetricLabels: map[string]string{"unit": "test"}}
	taskToFail := schema.Task{Name: "fail", MetricLabels: map[string]string{"unit": "test"}}
	taskFiles := bootstrapTaskFiles(t, taskWithMetricLabels, taskToFail)

	svr := &server.Server{}
	err := svr.Start(opts, taskFiles)
	require.NoError(t, err, "sever starts up")

	e := httpexpect.Default(t, opts.Config.ServerBaseUrl)

	// Schedule a new run of the first task.
	assertApiCall(e, apiCall{
		method: "POST",
		path:   "/api/v1/runs",
		requestBody: openapi.ScheduleRunV1Request{
			TaskName: taskWithMetricLabels.Name,
		},
		statusCode: http.StatusOK,
		responseBody: openapi.ScheduleRunV1Response{
			RunID: 1,
		},
	})

	// Process the run.
	assertApiCall(e, apiCall{
		method:     "GET",
		path:       "/api/v1/worker/work",
		statusCode: http.StatusOK,
		responseBody: openapi.GetWorkV1Response{
			RunID: 1,
			Task: openapi.WorkTaskV1{
				Hash: "353c7bd9ab71b13d31e30326209ea5abf093a5d96fb480bfc89251462617ac1f",
				Name: taskWithMetricLabels.Name,
			},
		},
	})

	// Report the result of the run.
	assertApiCall(e, apiCall{
		method: "POST",
		path:   "/api/v1/worker/work",
		requestBody: openapi.ReportWorkV1Request{
			RunID: 1,
			Task: openapi.WorkTaskV1{
				Name: taskWithMetricLabels.Name,
			},
			TaskResults: []openapi.ReportWorkV1TaskResult{},
		},
		statusCode: http.StatusCreated,
		responseBody: openapi.ReportWorkV1Response{
			Result: "ok",
		},
	})

	// Schedule a new run of the second task.
	assertApiCall(e, apiCall{
		method: "POST",
		path:   "/api/v1/runs",
		requestBody: openapi.ScheduleRunV1Request{
			TaskName: taskToFail.Name,
		},
		statusCode: http.StatusOK,
		responseBody: openapi.ScheduleRunV1Response{
			RunID: 2,
		},
	})

	// Process the run.
	assertApiCall(e, apiCall{
		method:     "GET",
		path:       "/api/v1/worker/work",
		statusCode: http.StatusOK,
		responseBody: openapi.GetWorkV1Response{
			RunID: 2,
			Task: openapi.WorkTaskV1{
				Hash: "8422e31abe2cce6e92df46cb5235cbe9b34463ffa54697b11dd364f5650ebb97",
				Name: taskToFail.Name,
			},
		},
	})

	// Report the result of the run.
	assertApiCall(e, apiCall{
		method: "POST",
		path:   "/api/v1/worker/work",
		requestBody: openapi.ReportWorkV1Request{
			Error: ptr.To("failed"),
			RunID: 2,
			Task: openapi.WorkTaskV1{
				Name: taskWithMetricLabels.Name,
			},
			TaskResults: []openapi.ReportWorkV1TaskResult{},
		},
		statusCode: http.StatusCreated,
		responseBody: openapi.ReportWorkV1Response{
			Result: "ok",
		},
	})

	// Verify the metrics.
	e.Request("GET", "/metrics").
		Expect().
		Status(http.StatusOK).
		Body().
		Contains(`sb_server_task_run_success{task="fail",unit="test"} 0`).
		Contains(`sb_server_task_run_success{task="metric-labels",unit="test"} 1`).
		Contains("sb_collector_success 1")
}

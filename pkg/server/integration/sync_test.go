package integration_test

import (
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

func Test_Sync(t *testing.T) {
	opts := setupOptions(t, nil, nil)
	taskNoTrigger := schema.Task{Name: "no-trigger"}
	taskCronTrigger := schema.Task{
		Name:    "cron-trigger",
		Trigger: &schema.TaskTrigger{Cron: ptr.To("3 6 * * *")},
	}
	taskToDelete := schema.Task{Name: "to-delete"}
	taskFilesFirst := bootstrapTaskFiles(t, []schema.Task{taskNoTrigger, taskCronTrigger, taskToDelete})

	serverFirst := &server.Server{}
	err := serverFirst.Start(opts, taskFilesFirst)
	require.NoError(t, err, "Server starts up the first time")
	time.Sleep(1 * time.Millisecond)
	err = serverFirst.Stop()
	require.NoError(t, err, "Server stops the first time")

	// Change the tasks to trigger sync
	taskNoTrigger.BranchName = "test/no-trigger"
	taskCronTrigger.BranchName = "test/cron-trigger"

	taskFilesSecond := bootstrapTaskFiles(t, []schema.Task{taskNoTrigger, taskCronTrigger})
	serverSecond := &server.Server{}
	promReg := prometheus.NewRegistry()
	opts.SetPrometheusRegistry(promReg)
	err = serverSecond.Start(opts, taskFilesSecond)
	require.NoError(t, err, "Server starts up the second time")
	time.Sleep(1 * time.Millisecond)
	defer func() {
		err := serverSecond.Stop()
		require.NoError(t, err, "Server stops the second time")
	}()

	e := httpexpect.Default(t, opts.Config.ServerBaseUrl)
	assertApiCall(e, apiCall{
		method:     "GET",
		path:       "/api/v1/tasks",
		statusCode: 200,
		responseBody: openapi.ListTasksV1Response{
			Tasks: []string{"no-trigger", "cron-trigger"},
		},
	})
	assertApiCall(e, apiCall{
		method:     "GET",
		path:       "/api/v1/worker/runs",
		statusCode: 200,
		responseBody: openapi.ListRunsV1Response{
			Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 2, TotalPages: 1},
			Result: []openapi.RunV1{
				{
					Task:          "cron-trigger",
					Id:            2,
					Reason:        openapi.New,
					ScheduleAfter: testDate(1, 6, 3, 0),
					Status:        openapi.Pending,
				},
				{
					Task:          "no-trigger",
					Id:            1,
					Reason:        openapi.New,
					ScheduleAfter: testDate(1, 0, 0, 0),
					Status:        openapi.Pending,
				},
			},
		},
	})
}

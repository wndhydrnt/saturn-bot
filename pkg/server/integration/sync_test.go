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
	taskToDelete := schema.Task{
		Name:    "to-delete",
		Trigger: &schema.TaskTrigger{Cron: ptr.To("0 4 * * *")},
	}
	taskCronTriggerLater := schema.Task{
		Name: "cron-trigger-later",
	}
	taskWithInput := schema.Task{
		Name: "with-input",
		Inputs: []schema.Input{
			{Name: "input-name"},
		},
	}
	taskFilesFirst := bootstrapTaskFiles(t, []schema.Task{taskNoTrigger, taskCronTrigger, taskToDelete, taskCronTriggerLater, taskWithInput})

	serverFirst := &server.Server{}
	err := serverFirst.Start(opts, taskFilesFirst)
	require.NoError(t, err, "Server starts up the first time")
	time.Sleep(1 * time.Millisecond)
	err = serverFirst.Stop()
	require.NoError(t, err, "Server stops the first time")

	// Change the tasks to trigger sync
	taskNoTrigger.BranchName = "test/no-trigger"
	taskCronTrigger.BranchName = "test/cron-trigger"
	taskCronTriggerLater.Trigger = &schema.TaskTrigger{Cron: ptr.To("16 15 * * *")}

	taskFilesSecond := bootstrapTaskFiles(t, []schema.Task{taskNoTrigger, taskCronTrigger, taskCronTriggerLater, taskWithInput})
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
		query:      "active=true",
		statusCode: 200,
		responseBody: openapi.ListTasksV1Response{
			Page: openapi.Page{
				CurrentPage:  1,
				ItemsPerPage: 20,
				TotalItems:   4,
				TotalPages:   1,
				NextPage:     0,
			},
			Results: []openapi.ListTasksV1ResponseTask{
				{Active: true, Name: "cron-trigger", Checksum: "33fee8a94207840f08247764d90f030cd641884975fbea5270fc4c03e6c17bce"},
				{Active: true, Name: "cron-trigger-later", Checksum: "3781388c6b2ab635138216950c99a8abc10b8cfc75ff73a848a3438ae51d90a9"},
				{Active: true, Name: "no-trigger", Checksum: "f4bd2a5e07bd61f8ac25edfd017821dac5a93d35dabf9862a1e1d623e3b448a3"},
				{Active: true, Name: "with-input", Checksum: "6dfad7cb419141aa5621d29bf3febc9c3a20992a282442edd0fb9822ba8bb8cc"},
			},
		},
	})
	assertApiCall(e, apiCall{
		method:     "GET",
		path:       "/api/v1/runs",
		statusCode: 200,
		responseBody: openapi.ListRunsV1Response{
			Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 2, TotalPages: 1},
			Result: []openapi.RunV1{
				{
					Task:          "cron-trigger-later",
					Id:            3,
					Reason:        openapi.Cron,
					ScheduleAfter: testDate(1, 15, 16, 0),
					Status:        openapi.Pending,
				},
				{
					Task:          "cron-trigger",
					Id:            1,
					Reason:        openapi.Cron,
					ScheduleAfter: testDate(1, 6, 3, 0),
					Status:        openapi.Pending,
				},
			},
		},
	})
}

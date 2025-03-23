package command_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/client"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

const (
	testServerUrl = "http://server.local"
)

func setupClient() *http.Client {
	httpClient := &http.Client{}
	gock.InterceptClient(httpClient)
	return httpClient
}

func TestScheduleRunner_Run_ScheduleNoWait(t *testing.T) {
	defer gock.Off()
	gock.New("http://server.local").
		Post("/api/v1/runs").
		MatchType("json").
		JSON(client.ScheduleRunV1Request{TaskName: "unittest"}).
		Reply(200).
		JSON(client.ScheduleRunV1Response{RunID: 1})
	logOut := &bytes.Buffer{}
	reportOut := &bytes.Buffer{}

	runner, err := command.NewScheduleRunner(command.NewScheduleRunnerOptions{
		HttpClient: setupClient(),
		ServerUrl:  testServerUrl,
	})
	require.NoError(t, err)
	err = runner.Run(command.ScheduleRunnerRunOptions{
		OutLog:          logOut,
		OutReport:       reportOut,
		ScheduleRequest: client.ScheduleRunV1Request{TaskName: "unittest"},
	})
	require.NoError(t, err)

	require.Equal(t, "▶️ Run 1 has been scheduled\n", logOut.String())
	require.True(t, gock.IsDone())
}

func TestScheduleRunner_Run_ScheduleWait(t *testing.T) {
	defer gock.Off()
	gock.New("http://server.local").
		Post("/api/v1/runs").
		MatchType("json").
		JSON(client.ScheduleRunV1Request{TaskName: "unittest"}).
		Reply(200).
		JSON(client.ScheduleRunV1Response{RunID: 1})
	gock.New("http://server.local").
		Get("/api/v1/runs/1").
		Reply(200).
		JSON(client.GetRunV1Response{Run: client.RunV1{
			Status: client.Running,
		}})
	gock.New("http://server.local").
		Get("/api/v1/runs/1").
		ReplyError(fmt.Errorf("timeout"))
	gock.New("http://server.local").
		Get("/api/v1/runs/1").
		Reply(500).
		SetHeader("Content-Type", "text/plain").
		BodyString("internal server error")
	gock.New("http://server.local").
		Get("/api/v1/runs/1").
		Reply(200).
		JSON(client.GetRunV1Response{Run: client.RunV1{
			Status: client.Finished,
		}})
	gock.New("http://server.local").
		Get("/api/v1/taskResults").
		MatchParams(map[string]string{
			"limit": "10",
			"page":  "1",
			"runId": "1",
		}).
		Reply(200).
		JSON(client.ListTaskResultsV1Response{
			TaskResults: []client.TaskResultV1{
				{PullRequestUrl: ptr.To("http://git.local/pr/1"), Status: client.TaskResultStateV1Open},
			},
		})

	runner, err := command.NewScheduleRunner(command.NewScheduleRunnerOptions{
		HttpClient: setupClient(),
		ServerUrl:  testServerUrl,
	})
	require.NoError(t, err)
	logOut := &bytes.Buffer{}
	reportOut := &bytes.Buffer{}
	err = runner.Run(command.ScheduleRunnerRunOptions{
		OutLog:          logOut,
		OutReport:       reportOut,
		OutputFormat:    "json",
		ScheduleRequest: client.ScheduleRunV1Request{TaskName: "unittest"},
		WaitFor:         5 * time.Millisecond,
		WaitInterval:    1 * time.Millisecond,
	})
	require.NoError(t, err)

	expectedLogOut := `▶️ Run 1 has been scheduled
⏰ Waiting 5ms for run 1 to finish
🔁 Run 1 running - 1ms until next check
⚠️ Failed to get run (Get "http://server.local/api/v1/runs/1": timeout) - 1ms until next check
🔁 Got unexpected status code 500 - 1ms until next check
✅ Run 1 finished
`
	require.Equal(t, expectedLogOut, logOut.String())
	expectedReportOut := `{
  "results": [
    {
      "pullRequestUrl": "http://git.local/pr/1",
      "status": "open"
    }
  ]
}
`
	require.Equal(t, expectedReportOut, reportOut.String())
	require.True(t, gock.IsDone())
}

func TestScheduleRunner_Run_WaitExceeded(t *testing.T) {
	defer gock.Off()
	gock.New("http://server.local").
		Post("/api/v1/runs").
		MatchType("json").
		JSON(client.ScheduleRunV1Request{TaskName: "unittest"}).
		Reply(200).
		JSON(client.ScheduleRunV1Response{RunID: 1})
	gock.New("http://server.local").
		Get("/api/v1/runs/1").
		Reply(200).
		JSON(client.GetRunV1Response{Run: client.RunV1{
			Status: client.Running,
		}})
	gock.New("http://server.local").
		Get("/api/v1/runs/1").
		Reply(200).
		JSON(client.GetRunV1Response{Run: client.RunV1{
			Status: client.Running,
		}})

	runner, err := command.NewScheduleRunner(command.NewScheduleRunnerOptions{
		HttpClient: setupClient(),
		ServerUrl:  testServerUrl,
	})
	require.NoError(t, err)
	logOut := &bytes.Buffer{}
	reportOut := &bytes.Buffer{}
	err = runner.Run(command.ScheduleRunnerRunOptions{
		OutLog:          logOut,
		OutReport:       reportOut,
		ScheduleRequest: client.ScheduleRunV1Request{TaskName: "unittest"},
		WaitFor:         2 * time.Millisecond,
		WaitInterval:    1 * time.Millisecond,
	})
	require.Error(t, err)

	expectedOut := `▶️ Run 1 has been scheduled
⏰ Waiting 2ms for run 1 to finish
🔁 Run 1 running - 1ms until next check
🔁 Run 1 running - 1ms until next check
❌ Run failed to finish after 2ms
`
	require.Equal(t, expectedOut, logOut.String())
	require.True(t, gock.IsDone())
}

func TestScheduleRunner_Run_Fails(t *testing.T) {
	defer gock.Off()
	gock.New("http://server.local").
		Post("/api/v1/runs").
		MatchType("json").
		JSON(client.ScheduleRunV1Request{TaskName: "unittest"}).
		Reply(200).
		JSON(client.ScheduleRunV1Response{RunID: 1})
	gock.New("http://server.local").
		Get("/api/v1/runs/1").
		Reply(200).
		JSON(client.GetRunV1Response{Run: client.RunV1{
			Status: client.Running,
		}})
	gock.New("http://server.local").
		Get("/api/v1/runs/1").
		Reply(200).
		JSON(client.GetRunV1Response{Run: client.RunV1{
			Status: client.Failed,
		}})

	runner, err := command.NewScheduleRunner(command.NewScheduleRunnerOptions{
		HttpClient: setupClient(),
		ServerUrl:  testServerUrl,
	})
	require.NoError(t, err)
	logOut := &bytes.Buffer{}
	reportOut := &bytes.Buffer{}
	err = runner.Run(command.ScheduleRunnerRunOptions{
		OutLog:          logOut,
		OutReport:       reportOut,
		ScheduleRequest: client.ScheduleRunV1Request{TaskName: "unittest"},
		WaitFor:         2 * time.Millisecond,
		WaitInterval:    1 * time.Millisecond,
	})
	require.Error(t, err)

	expectedOut := `▶️ Run 1 has been scheduled
⏰ Waiting 2ms for run 1 to finish
🔁 Run 1 running - 1ms until next check
❌ Run failed
`
	require.Equal(t, expectedOut, logOut.String())
	require.True(t, gock.IsDone())
}

func TestScheduleRunner_Run_BadRequest(t *testing.T) {
	defer gock.Off()
	gock.New("http://server.local").
		Post("/api/v1/runs").
		MatchType("json").
		JSON(client.ScheduleRunV1Request{TaskName: "unittest"}).
		Reply(400).
		JSON(client.Error{Errors: []client.ErrorDetail{
			{Message: "bad request"},
		}})
	logOut := &bytes.Buffer{}
	reportOut := &bytes.Buffer{}

	runner, err := command.NewScheduleRunner(command.NewScheduleRunnerOptions{
		HttpClient: setupClient(),
		ServerUrl:  testServerUrl,
	})
	require.NoError(t, err)
	err = runner.Run(command.ScheduleRunnerRunOptions{
		OutLog:          logOut,
		OutReport:       reportOut,
		ScheduleRequest: client.ScheduleRunV1Request{TaskName: "unittest"},
	})
	require.Error(t, err)

	require.Equal(t, "❌ Failed to schedule run:\n  Error: bad request\n", logOut.String())
	require.True(t, gock.IsDone())
}

package command_test

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/client"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
)

const (
	testServerUrl = "http://server.local"
)

func TestScheduleRunner_Run_ScheduleNoWait(t *testing.T) {
	defer gock.Off()
	gock.New("http://server.local").
		Post("/api/v1/runs").
		MatchType("json").
		JSON(client.ScheduleRunV1Request{TaskName: "unittest"}).
		Reply(200).
		JSON(client.ScheduleRunV1Response{RunID: 1})
	out := &bytes.Buffer{}

	runner, err := command.NewScheduleRunner(command.ScheduleOptions{
		HttpClient: setupClient(),
		ServerUrl:  testServerUrl,
	})
	require.NoError(t, err)
	err = runner.Run(out, client.ScheduleRunV1Request{TaskName: "unittest"})
	require.NoError(t, err)

	require.Equal(t, "▶️ Run 1 has been scheduled\n", out.String())
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
		Reply(200).
		JSON(client.GetRunV1Response{Run: client.RunV1{
			Status: client.Finished,
		}})

	runner, err := command.NewScheduleRunner(command.ScheduleOptions{
		HttpClient:   setupClient(),
		WaitFor:      3 * time.Millisecond,
		WaitInterval: 1 * time.Millisecond,
		ServerUrl:    testServerUrl,
	})
	require.NoError(t, err)
	out := &bytes.Buffer{}
	err = runner.Run(out, client.ScheduleRunV1Request{TaskName: "unittest"})
	require.NoError(t, err)

	expectedOut := `▶️ Run 1 has been scheduled
⏰ Waiting 3ms for run 1 to finish
🔁 Run 1 running - 1ms until next check
✅ Run 1 finished
`
	require.Equal(t, expectedOut, out.String())
	require.True(t, gock.IsDone())
}

func setupClient() *http.Client {
	httpClient := &http.Client{}
	gock.InterceptClient(httpClient)
	return httpClient
}

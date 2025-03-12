package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/client"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

var (
	ctx                 = context.Background()
	ErrSchedule         = errors.New("schedule")
	errSchedulingFailed = errors.New("failed to schedule the run")
)

type ScheduleOptions struct {
	HttpClient   *http.Client
	WaitFor      time.Duration
	WaitInterval time.Duration
	ServerApiKey string
	ServerUrl    string
}

func NewScheduleRunner(opts ScheduleOptions) (*ScheduleRunner, error) {

	client, err := client.NewCustomClientWithResponses(client.CustomClientWithResponsesOptions{
		ApiKey:     opts.ServerApiKey,
		BaseUrl:    opts.ServerUrl,
		HttpClient: opts.HttpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("NewScheduleRunner: %w", err)
	}

	return &ScheduleRunner{
		client:       client,
		waitFor:      opts.WaitFor,
		waitInterval: opts.WaitInterval,
	}, nil
}

type ScheduleRunner struct {
	client       client.ClientWithResponsesInterface
	waitFor      time.Duration
	waitInterval time.Duration
}

func (s *ScheduleRunner) Run(out io.Writer, payload client.ScheduleRunV1Request) error {
	response, err := s.client.ScheduleRunV1WithResponse(ctx, payload)
	if err != nil {
		return fmt.Errorf("schedule run: %w", err)
	}

	switch {
	case response.JSON200 != nil:
		_, _ = fmt.Fprintf(out, "▶️ Run %d has been scheduled\n", response.JSON200.RunID)
		if s.waitFor > 0 {
			return s.waitForRun(out, response.JSON200.RunID)
		}
	case response.JSON400 != nil:
		handleApiError(out, response.JSON400)
		return errSchedulingFailed
	case response.JSON401 != nil:
		handleApiError(out, response.JSON401)
		return errSchedulingFailed
	default:
		return fmt.Errorf("unexpected HTTP response with code %d", response.HTTPResponse.StatusCode)
	}

	return nil
}

func (s *ScheduleRunner) waitForRun(out io.Writer, id int) error {
	_, _ = fmt.Fprintf(out, "⏰ Waiting %s for run %d to finish\n", s.waitFor, id)
	var tries int64
	maxTries := s.waitFor / s.waitInterval
	ticker := time.NewTicker(s.waitInterval)
	for range ticker.C {
		tries++
		response, err := s.client.GetRunV1WithResponse(ctx, id)
		if err != nil {
			_, _ = fmt.Fprintf(out, "⚠️ Failed to get run (%s) - %s until next check\n", err, s.waitInterval)
		}

		if response != nil {
			switch {
			case response.JSON200 != nil:
				switch response.JSON200.Run.Status {
				case client.Failed:
					_, _ = fmt.Fprintf(out, "❌ Run failed\n")
					return ErrSchedule
				case client.Finished:
					_, _ = fmt.Fprintf(out, "✅ Run %d finished\n", id)
					return nil
				default:
					_, _ = fmt.Fprintf(out, "🔁 Run %d %s - %s until next check\n", id, response.JSON200.Run.Status, s.waitInterval)
				}
			case response.JSON401 != nil:
				return fmt.Errorf("failed to authenticate")
			case response.JSON404 != nil:
				return fmt.Errorf("run %d not found", id)
			default:
				if response.HTTPResponse != nil {
					_, _ = fmt.Fprintf(out, "🔁 Got unexpected status code %d - %s until next check\n", response.HTTPResponse.StatusCode, s.waitInterval)
				}
			}
		}

		if tries == int64(maxTries) {
			_, _ = fmt.Fprintf(out, "❌ Run failed to finish after %s\n", s.waitFor)
			return ErrSchedule
		}
	}

	return nil
}

func handleApiError(out io.Writer, apiErr *client.Error) {
	_, _ = fmt.Fprintf(out, "❌ Failed to schedule run:\n")
	for _, errDetail := range apiErr.Errors {
		_, _ = fmt.Fprintf(out, "  Error: %s\n", errDetail.Message)
		if errDetail.Detail != nil {
			_, _ = fmt.Fprintf(out, "    Detail: %s\n", ptr.From(errDetail.Detail))
		}
	}
}

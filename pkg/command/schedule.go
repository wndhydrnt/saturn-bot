package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/client"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

var (
	ctx                 = context.Background()
	ErrSchedule         = errors.New("schedule")
	errSchedulingFailed = errors.New("failed to schedule the run")
)

type result struct {
	PullRequestUrl string `json:"pullRequestUrl"`
	Status         string `json:"status"`
}

type resultList struct {
	Results []result `json:"results"`
}

type ScheduleOptions struct {
	HttpClient   *http.Client
	OutputFormat string
	WaitFor      time.Duration
	WaitInterval time.Duration
	ServerApiKey string
	ServerUrl    string
}

func NewScheduleRunner(opts ScheduleOptions) (*ScheduleRunner, error) {
	runner := &ScheduleRunner{
		waitFor:      opts.WaitFor,
		waitInterval: opts.WaitInterval,
	}

	if strings.ToLower(opts.OutputFormat) == "json" {
		runner.reporter = reportResultsAsJson
	}

	client, err := client.NewCustomClientWithResponses(client.CustomClientWithResponsesOptions{
		ApiKey:     opts.ServerApiKey,
		BaseUrl:    opts.ServerUrl,
		HttpClient: opts.HttpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("NewScheduleRunner: %w", err)
	}

	runner.client = client
	return runner, nil
}

type ScheduleRunner struct {
	client       client.ClientWithResponsesInterface
	reporter     reporter
	waitFor      time.Duration
	waitInterval time.Duration
}

func (s *ScheduleRunner) Run(logOut io.Writer, reportOut io.Writer, payload client.ScheduleRunV1Request) error {
	response, err := s.client.ScheduleRunV1WithResponse(ctx, payload)
	if err != nil {
		return fmt.Errorf("schedule run: %w", err)
	}

	switch {
	case response.JSON200 != nil:
		_, _ = fmt.Fprintf(logOut, "▶️ Run %d has been scheduled\n", response.JSON200.RunID)
		if s.waitFor > 0 {
			return s.waitForRun(logOut, reportOut, response.JSON200.RunID)
		}
	case response.JSON400 != nil:
		handleApiError(logOut, response.JSON400)
		return errSchedulingFailed
	case response.JSON401 != nil:
		handleApiError(logOut, response.JSON401)
		return errSchedulingFailed
	default:
		return fmt.Errorf("unexpected HTTP response with code %d", response.HTTPResponse.StatusCode)
	}

	return nil
}

func (s *ScheduleRunner) waitForRun(logOut io.Writer, reportOut io.Writer, id int) error {
	_, _ = fmt.Fprintf(logOut, "⏰ Waiting %s for run %d to finish\n", s.waitFor, id)
	var tries int64
	maxTries := s.waitFor / s.waitInterval
	ticker := time.NewTicker(s.waitInterval)
	for range ticker.C {
		tries++
		response, err := s.client.GetRunV1WithResponse(ctx, id)
		if err != nil {
			_, _ = fmt.Fprintf(logOut, "⚠️ Failed to get run (%s) - %s until next check\n", err, s.waitInterval)
		}

		if response != nil {
			switch {
			case response.JSON200 != nil:
				switch response.JSON200.Run.Status {
				case client.Failed:
					_, _ = fmt.Fprintf(logOut, "❌ Run failed\n")
					return ErrSchedule
				case client.Finished:
					_, _ = fmt.Fprintf(logOut, "✅ Run %d finished\n", id)
					return s.reportResults(reportOut, id)
				default:
					_, _ = fmt.Fprintf(logOut, "🔁 Run %d %s - %s until next check\n", id, response.JSON200.Run.Status, s.waitInterval)
				}
			case response.JSON401 != nil:
				return fmt.Errorf("failed to authenticate")
			case response.JSON404 != nil:
				return fmt.Errorf("run %d not found", id)
			default:
				if response.HTTPResponse != nil {
					_, _ = fmt.Fprintf(logOut, "🔁 Got unexpected status code %d - %s until next check\n", response.HTTPResponse.StatusCode, s.waitInterval)
				}
			}
		}

		if tries == int64(maxTries) {
			_, _ = fmt.Fprintf(logOut, "❌ Run failed to finish after %s\n", s.waitFor)
			return ErrSchedule
		}
	}

	return nil
}

func (s *ScheduleRunner) reportResults(out io.Writer, runId int) error {
	if s.reporter == nil {
		return nil
	}

	var results []result
	listOpts := client.ListOptions{
		Limit: 10,
		Page:  1,
	}
	for {
		resp, err := s.client.ListTaskResultsV1WithResponse(ctx, &client.ListTaskResultsV1Params{
			RunId:       ptr.To(runId),
			ListOptions: ptr.To(listOpts),
		})
		if err != nil {
			return fmt.Errorf("list results: %w", err)
		}

		if resp.JSON200 != nil {
			for _, tr := range resp.JSON200.TaskResults {
				r := result{Status: string(tr.Status)}
				if tr.PullRequestUrl != nil {
					r.PullRequestUrl = ptr.From(tr.PullRequestUrl)
				}

				results = append(results, r)
			}

			if resp.JSON200.Page.NextPage == 0 {
				break
			}

			listOpts.Page = resp.JSON200.Page.NextPage
		}
	}

	if len(results) == 0 {
		return nil
	}

	return s.reporter(out, resultList{Results: results})
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

type reporter func(io.Writer, resultList) error

func reportResultsAsJson(out io.Writer, results resultList) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

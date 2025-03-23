package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/client"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

var (
	ctx = context.Background()
	// ErrRunFailed indicates that a run failed with an error
	ErrRunFailed        = errors.New("run failed")
	errSchedulingFailed = errors.New("failed to schedule the run")
)

type result struct {
	PullRequestUrl string `json:"pullRequestUrl"`
	Status         string `json:"status"`
}

type resultList struct {
	Results []result `json:"results"`
}

// NewScheduleRunnerOptions defines all options expected by [NewScheduleRunner].
type NewScheduleRunnerOptions struct {
	HttpClient   *http.Client
	ServerApiKey string
	ServerUrl    string
}

// NewScheduleRunner initializes a new [ScheduleRunner] from [NewScheduleRunnerOptions].
func NewScheduleRunner(opts NewScheduleRunnerOptions) (*ScheduleRunner, error) {
	client, err := client.NewCustomClientWithResponses(client.CustomClientWithResponsesOptions{
		ApiKey:     opts.ServerApiKey,
		BaseUrl:    opts.ServerUrl,
		HttpClient: opts.HttpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("NewScheduleRunner: %w", err)
	}

	return &ScheduleRunner{client: client}, nil
}

// ScheduleRunner wraps the logic to schedule a new run.
type ScheduleRunner struct {
	client client.ClientWithResponsesInterface
}

// ScheduleRunnerRunOptions are all options expected by [ScheduleRunnerRunOptions].
type ScheduleRunnerRunOptions struct {
	// OutLog is the writer that [ScheduleRunner]
	// writes human-readable log messages to.
	OutLog io.Writer
	// OutReport is the writer that [ScheduleRunner]
	// writes task results to, if the user requests it.
	OutReport io.Writer
	// OutputFormat defines the output format to use.
	OutputFormat string
	// ScheduleRequest is the data to send to saturn-bot server.
	ScheduleRequest client.ScheduleRunV1Request
	// WaitFor is the total amount of time to wait for the run to complete.
	WaitFor time.Duration
	// WaitInterval is the time between requests sent to saturn-bot server
	// to check if the run has finished.
	WaitInterval time.Duration
}

// Run schedules a new run from payload via the API of saturn-bot server.
func (s *ScheduleRunner) Run(opts ScheduleRunnerRunOptions) error {
	response, err := s.client.ScheduleRunV1WithResponse(ctx, opts.ScheduleRequest)
	if err != nil {
		return fmt.Errorf("schedule run: %w", err)
	}

	switch {
	case response.JSON200 != nil:
		_, _ = fmt.Fprintf(opts.OutLog, "▶️ Run %d has been scheduled\n", response.JSON200.RunID)
		if opts.WaitFor > 0 {
			err := s.waitForRun(response.JSON200.RunID, opts)
			if err != nil {
				return err
			}

			return s.reportResults(response.JSON200.RunID, opts)
		}
	case response.JSON400 != nil:
		handleApiError(opts.OutLog, response.JSON400)
		return errSchedulingFailed
	case response.JSON401 != nil:
		handleApiError(opts.OutLog, response.JSON401)
		return errSchedulingFailed
	default:
		return fmt.Errorf("unexpected HTTP response with code %d", response.HTTPResponse.StatusCode)
	}

	return nil
}

func (s *ScheduleRunner) waitForRun(id int, opts ScheduleRunnerRunOptions) error {
	_, _ = fmt.Fprintf(opts.OutLog, "⏰ Waiting %s for run %d to finish\n", opts.WaitFor, id)
	var tries int64
	maxTries := opts.WaitFor / opts.WaitInterval
	ticker := time.NewTicker(opts.WaitInterval)
	for range ticker.C {
		tries++
		response, err := s.client.GetRunV1WithResponse(ctx, id)
		if err != nil {
			_, _ = fmt.Fprintf(opts.OutLog, "⚠️ Failed to get run (%s) - %s until next check\n", err, opts.WaitInterval)
		}

		if response != nil {
			switch {
			case response.JSON200 != nil:
				switch response.JSON200.Run.Status {
				case client.Failed:
					_, _ = fmt.Fprintf(opts.OutLog, "❌ Run failed\n")
					return ErrRunFailed
				case client.Finished:
					_, _ = fmt.Fprintf(opts.OutLog, "✅ Run %d finished\n", id)
					return nil
				default:
					_, _ = fmt.Fprintf(opts.OutLog, "🔁 Run %d %s - %s until next check\n", id, response.JSON200.Run.Status, opts.WaitInterval)
				}
			case response.JSON401 != nil:
				return fmt.Errorf("failed to authenticate")
			case response.JSON404 != nil:
				return fmt.Errorf("run %d not found", id)
			default:
				if response.HTTPResponse != nil {
					_, _ = fmt.Fprintf(opts.OutLog, "🔁 Got unexpected status code %d - %s until next check\n", response.HTTPResponse.StatusCode, opts.WaitInterval)
				}
			}
		}

		if tries == int64(maxTries) {
			_, _ = fmt.Fprintf(opts.OutLog, "❌ Run failed to finish after %s\n", opts.WaitFor)
			return ErrRunFailed
		}
	}

	return nil
}

func (s *ScheduleRunner) reportResults(runId int, opts ScheduleRunnerRunOptions) error {
	if opts.OutputFormat != "json" {
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

	return reportResultsAsJson(opts.OutReport, resultList{Results: results})
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

func reportResultsAsJson(out io.Writer, results resultList) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

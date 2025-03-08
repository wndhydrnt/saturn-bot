package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/itchyny/gojq"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

type webhookType string

const (
	webhookTypeGithub webhookType = "GitHub"
	webhookTypeGitlab webhookType = "GitLab"
)

type cacheEntry struct {
	delay   time.Duration
	event   string
	filters []*gojq.Code
	// Key is the key to set in the run data map.
	// Value is the jq expression that extracts data from the payload of a webhook.
	runDataExtractors map[string]*gojq.Code
}

// WebhookService handles scheduling new runs when a webhook is received.
type WebhookService struct {
	clock              clock.Clock
	githubTriggerCache map[string][]cacheEntry
	gitlabTriggerCache map[string][]cacheEntry
	taskRegistry       *task.Registry
	workerService      *WorkerService
}

// NewWebhookService returns a new instance of [WebhookService].
// It parses the triggers defined by tasks and caches them.
func NewWebhookService(clock clock.Clock, taskRegistry *task.Registry, workerService *WorkerService) (*WebhookService, error) {
	s := &WebhookService{
		clock:         clock,
		taskRegistry:  taskRegistry,
		workerService: workerService,
	}
	err := s.populateCaches()
	if err != nil {
		return nil, fmt.Errorf("populate filter caches: %w", err)
	}

	return s, nil
}

type WebhookEnqueueInput struct {
	Event   string
	ID      string
	Payload any
}

// EnqueueGithub enqueues new runs when a webhook from GitHub is received by the server.
// It visits every task that configures triggers for GitHub webhooks.
func (s *WebhookService) EnqueueGithub(in *WebhookEnqueueInput) error {
	return s.enqueue(in, s.githubTriggerCache, webhookTypeGithub)
}

// EnqueueGitlab enqueues new runs when a webhook from GitLab is received by the server.
// It visits every task that configures triggers for GitLab webhooks.
func (s *WebhookService) EnqueueGitlab(in *WebhookEnqueueInput) error {
	return s.enqueue(in, s.gitlabTriggerCache, webhookTypeGitlab)
}

func (s *WebhookService) enqueue(in *WebhookEnqueueInput, triggerCache map[string][]cacheEntry, wtype webhookType) error {
	var errs []error
	for taskName, triggers := range triggerCache {
		for _, trigger := range triggers {
			if match(in.Event, trigger, in.Payload) {
				log.Log().Debugf("Task %s matches %s webhook %s", taskName, wtype, in.ID)
				runData := extractRunData(in.Payload, trigger.runDataExtractors)
				scheduleAfter := s.clock.Now().Add(trigger.delay)
				_, err := s.workerService.ScheduleRun(db.RunReasonWebhook, nil, scheduleAfter, taskName, runData, nil)
				errs = append(errs, err)
				break
			}
		}
	}

	return errors.Join(errs...)
}

func (s *WebhookService) populateCaches() error {
	s.githubTriggerCache = map[string][]cacheEntry{}
	s.gitlabTriggerCache = map[string][]cacheEntry{}
	for _, t := range s.taskRegistry.GetTasks() {
		if hasGithubWebhookTrigger(t.Trigger) {
			err := s.populateGithubCache(t)
			if err != nil {
				return err
			}
		}

		if hasGitlabWebhookTrigger(t.Trigger) {
			err := s.populateGitlabCache(t)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *WebhookService) populateGithubCache(t *task.Task) error {
	cacheEntries := make([]cacheEntry, len(t.Trigger.Webhook.Github))
	for idxHook, hook := range t.Trigger.Webhook.Github {
		filters, err := parseHookFilters(hook.Filters)
		if err != nil {
			return fmt.Errorf("parse filters of GitHub webhook %d: %w", idxHook, err)
		}

		extractors, err := parseHookExtractors(hook.RunData)
		if err != nil {
			return fmt.Errorf("parse runData extractors of GitHub webhook %d: %w", idxHook, err)
		}

		cacheEntries[idxHook] = cacheEntry{
			delay:             time.Duration(t.Trigger.Webhook.Delay) * time.Second,
			event:             ptr.From(hook.Event),
			filters:           filters,
			runDataExtractors: extractors,
		}
	}

	s.githubTriggerCache[t.Name] = cacheEntries
	return nil
}

func (s *WebhookService) populateGitlabCache(t *task.Task) error {
	cacheEntries := make([]cacheEntry, len(t.Trigger.Webhook.Gitlab))
	for idxHook, hook := range t.Trigger.Webhook.Gitlab {
		filters, err := parseHookFilters(hook.Filters)
		if err != nil {
			return fmt.Errorf("parse filters of GitLab webhook %d: %w", idxHook, err)
		}

		extractors, err := parseHookExtractors(hook.RunData)
		if err != nil {
			return fmt.Errorf("parse runData extractors of GitLab webhook %d: %w", idxHook, err)
		}

		cacheEntries[idxHook] = cacheEntry{
			delay:             time.Duration(t.Trigger.Webhook.Delay) * time.Second,
			event:             ptr.From(hook.Event),
			filters:           filters,
			runDataExtractors: extractors,
		}
	}

	s.gitlabTriggerCache[t.Name] = cacheEntries
	return nil
}

func hasGithubWebhookTrigger(trigger *schema.TaskTrigger) bool {
	if trigger == nil {
		return false
	}

	if trigger.Webhook == nil {
		return false
	}

	if len(trigger.Webhook.Github) == 0 {
		return false
	}

	return true
}

func hasGitlabWebhookTrigger(trigger *schema.TaskTrigger) bool {
	if trigger == nil {
		return false
	}

	if trigger.Webhook == nil {
		return false
	}

	if len(trigger.Webhook.Gitlab) == 0 {
		return false
	}

	return true
}

func match(event string, trigger cacheEntry, payload any) bool {
	if event != trigger.event {
		return false
	}

	for _, f := range trigger.filters {
		valueRaw, hasNext := f.Run(payload).Next()
		if !hasNext {
			return false
		}

		if valueRaw == nil {
			// Query matches nothing.
			return false
		}

		value, ok := valueRaw.(bool)
		if !ok {
			// Query doesn't return a boolean.
			return false
		}

		if !value {
			// Query evaluates to false.
			return false
		}
	}

	// All queries returned true.
	// It's a match!
	return true
}

func parseHookFilters(filters []string) ([]*gojq.Code, error) {
	codes := make([]*gojq.Code, len(filters))
	for idxFilter, filter := range filters {
		query, err := gojq.Parse(filter)
		if err != nil {
			return nil, fmt.Errorf("parse jq expression of item %d: %w", idxFilter, err)
		}

		compiledQuery, err := gojq.Compile(query)
		if err != nil {
			return nil, fmt.Errorf("compile jq expression of item %d: %w", idxFilter, err)
		}

		codes[idxFilter] = compiledQuery
	}

	return codes, nil
}

func extractRunData(payload any, extractors map[string]*gojq.Code) map[string]string {
	data := map[string]string{}
	for key, code := range extractors {
		valueRaw, hasNext := code.Run(payload).Next()
		if !hasNext {
			continue
		}

		if valueRaw == nil {
			// Query matches nothing.
			continue
		}

		data[key] = fmt.Sprintf("%v", valueRaw)
	}

	return data
}

func parseHookExtractors(raw map[string]string) (map[string]*gojq.Code, error) {
	extractors := make(map[string]*gojq.Code, len(raw))
	for key, expr := range raw {
		query, err := gojq.Parse(expr)
		if err != nil {
			return nil, fmt.Errorf("parse jq expression of runData for key %s: %w", key, err)
		}

		compiledQuery, err := gojq.Compile(query)
		if err != nil {
			return nil, fmt.Errorf("compile jq expression of runData for key %s: %w", key, err)
		}

		extractors[key] = compiledQuery
	}

	return extractors, nil
}

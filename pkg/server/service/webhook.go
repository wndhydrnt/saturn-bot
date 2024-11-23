package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/itchyny/gojq"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

type cacheEntry struct {
	event   string
	filters []*gojq.Code
}

type WebhookService struct {
	githubTriggerCache map[string][]cacheEntry
	gitlabTriggerCache map[string][]cacheEntry
	taskRegistry       *task.Registry
	workerService      *WorkerService
}

func NewWebhookService(taskRegistry *task.Registry, workerService *WorkerService) (*WebhookService, error) {
	s := &WebhookService{taskRegistry: taskRegistry, workerService: workerService}
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

func (s *WebhookService) EnqueueGithub(in *WebhookEnqueueInput) error {
	return s.enqueue(in, s.githubTriggerCache)
}

func (s *WebhookService) EnqueueGitlab(in *WebhookEnqueueInput) error {
	return s.enqueue(in, s.gitlabTriggerCache)
}

func (s *WebhookService) enqueue(in *WebhookEnqueueInput, triggerCache map[string][]cacheEntry) error {
	var errs []error
	for taskName, triggers := range triggerCache {
		for _, trigger := range triggers {
			if match(in.Event, trigger, in.Payload) {
				log.Log().Debugf("Task %s matches webhook %s", taskName, in.ID)
				_, err := s.workerService.ScheduleRun(db.RunReasonWebhook, nil, time.Now(), taskName, nil)
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
			return fmt.Errorf("parse GitHub webhook %d: %w", idxHook, err)
		}

		cacheEntries[idxHook] = cacheEntry{event: ptr.From(hook.Event), filters: filters}
	}

	s.githubTriggerCache[t.Name] = cacheEntries
	return nil
}

func (s *WebhookService) populateGitlabCache(t *task.Task) error {
	cacheEntries := make([]cacheEntry, len(t.Trigger.Webhook.Gitlab))
	for idxHook, hook := range t.Trigger.Webhook.Gitlab {
		filters, err := parseHookFilters(hook.Filters)
		if err != nil {
			return fmt.Errorf("parse GitLab webhook %d: %w", idxHook, err)
		}

		cacheEntries[idxHook] = cacheEntry{event: ptr.From(hook.Event), filters: filters}
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
	var codes []*gojq.Code
	for idxFilter, filter := range filters {
		query, err := gojq.Parse(filter)
		if err != nil {
			return nil, fmt.Errorf("parse jq expression of item %d: %w", idxFilter, err)
		}

		compiledQuery, err := gojq.Compile(query)
		if err != nil {
			return nil, fmt.Errorf("compile jq expression of item %d: %w", idxFilter, err)
		}
		codes = append(codes, compiledQuery)
	}

	return codes, nil
}

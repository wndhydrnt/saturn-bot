package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/itchyny/gojq"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

type WebhookType uint

const (
	GithubWebhookType WebhookType = iota
)

type WebhookService struct {
	githubFilterCache map[string]map[int][]*gojq.Code
	tasks             []schema.ReadResult
	workerService     *WorkerService
}

func NewWebhookService(tasks []schema.ReadResult, workerService *WorkerService) (*WebhookService, error) {
	s := &WebhookService{tasks: tasks, workerService: workerService}
	err := s.populateCaches()
	if err != nil {
		return nil, fmt.Errorf("populate filter caches: %w", err)
	}

	return s, nil
}

type EnqueueInput struct {
	Event   string
	ID      string
	Payload any
	Type    WebhookType
}

func (s *WebhookService) Enqueue(in *EnqueueInput) error {
	if s.githubFilterCache == nil {
		err := s.populateCaches()
		if err != nil {
			return err
		}
	}

	var errs []error
	for _, task := range s.tasks {
		if !hasGithubWebhookTrigger(task.Task.Trigger) {
			log.Log().Debugf("Task %s does not define GitHub webhooks", task.Task.Name)
			continue
		}

		for idx, trigger := range task.Task.Trigger.Webhook.Github {
			if s.match(in.Event, idx, task.Task.Name, trigger, in.Payload) {
				log.Log().Debugf("Task %s matches GitHub webhook %s", task.Task.Name, in.ID)
				_, err := s.workerService.ScheduleRun(db.RunReasonWebhook, nil, time.Now(), task.Task.Name, nil)
				errs = append(errs, err)
			} else {
				log.Log().Debugf("Task %s does not match GitHub webhook %s", task.Task.Name, in.ID)
			}
		}
	}

	return errors.Join(errs...)
}

func (s *WebhookService) populateCaches() error {
	m := map[string]map[int][]*gojq.Code{}
	for _, t := range s.tasks {
		if !hasGithubWebhookTrigger(t.Task.Trigger) {
			continue
		}

		mm := map[int][]*gojq.Code{}
		for idxHook, hook := range t.Task.Trigger.Webhook.Github {
			var filters []*gojq.Code
			for idxFilter, filter := range hook.Filters {
				query, err := gojq.Parse(filter)
				if err != nil {
					return fmt.Errorf("parse jq expression of GitHub webhook %d item %d: %w", idxHook, idxFilter, err)
				}

				compiledQuery, err := gojq.Compile(query)
				if err != nil {
					return fmt.Errorf("compile jq expression of GitHub webhook %d item %d: %w", idxHook, idxFilter, err)
				}
				filters = append(filters, compiledQuery)
			}

			mm[idxHook] = filters
		}

		m[t.Task.Name] = mm
	}

	s.githubFilterCache = m
	return nil
}

func (s *WebhookService) match(event string, idx int, taskName string, trigger schema.GithubTrigger, payload any) bool {
	if ptr.From(trigger.Event) != event {
		return false
	}

	filters := s.githubFilterCache[taskName][idx]
	for _, f := range filters {
		valueRaw, hasNext := f.Run(payload).Next()
		if !hasNext {
			return false
		}

		if valueRaw == nil {
			// Query matched nothing.
			return false
		}

		value, ok := valueRaw.(bool)
		if !ok {
			// If valueRaw isn't of type bool,
			// then the query returned an object or the value of a field.
			// Example query: '.repository.full_name' returns a string
			// Consider this a match.
			continue
		}

		if !value {
			return false
		}
	}

	return true
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

package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"go.uber.org/zap"
)

const (
	collectSuccessFqName = "sb_collector_success"
	collectSuccessHelp   = "Status of the last metric collection. 1 = metric collection, 0 = an error occurred."
	taskRunSuccessFqName = "sb_server_task_run_success"
	taskRunSuccessHelp   = "Status of the latest run of a task. 1 = the run succeeded, 0 = the run failed."
)

var (
	collectSuccessDesc = prometheus.NewDesc(collectSuccessFqName, collectSuccessHelp, []string{}, nil)
)

// Collector is a [prometheus.Collector] that exposes custom metrics of the server.
type Collector struct {
	taskService   *service.TaskService
	workerService *service.WorkerService
}

// NewCollector returns a new [Collector].
func NewCollector(taskService *service.TaskService, workerService *service.WorkerService) *Collector {
	return &Collector{
		taskService:   taskService,
		workerService: workerService,
	}
}

// Collect implements [prometheus.Collector].
// It collects the latest status of a run for each known task.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var collectResult float64 = 1
	for _, t := range c.taskService.ListTasks() {
		metric, err := c.createMetricForTask(t)
		if err != nil {
			log.Log().Warnw("Failed to collect task status metric", zap.Error(err))
			collectResult = 0
		}

		if metric != nil {
			ch <- *metric
		}
	}

	ch <- prometheus.MustNewConstMetric(collectSuccessDesc, prometheus.GaugeValue, collectResult)
}

// Describe implements [prometheus.Collector].
// This is a noop, which makes the collector "unchecked".
// An unchecked collector can return metrics with an arbitrary number of labels.
// This makes it possible for users to pass custom labels via the MetricLabels
// field of [github.com/wndhydrnt/saturn-bot/pkg/task/schema.Task].
func (c *Collector) Describe(_ chan<- *prometheus.Desc) {}

func (c *Collector) createMetricForTask(t *task.Task) (*prometheus.Metric, error) {
	listOpts := service.ListOptions{Limit: 1}
	runs, err := c.workerService.ListRuns(
		service.ListRunsOptions{
			Status:   []db.RunStatus{db.RunStatusFailed, db.RunStatusFinished},
			TaskName: t.Name,
		},
		&listOpts,
	)
	if err != nil {
		return nil, fmt.Errorf("list runs to create metric: %w", err)
	}

	if listOpts.TotalItems() == 0 {
		// No reported runs for the task yet.
		return nil, nil
	}

	var metricValue float64
	if runs[0].Status == db.RunStatusFinished {
		metricValue = 1
	}

	labelKeys := []string{"task"}
	labelValues := []string{t.Name}
	for k, v := range t.MetricLabels {
		labelKeys = append(labelKeys, k)
		labelValues = append(labelValues, v)
	}

	desc := prometheus.NewDesc(
		taskRunSuccessFqName,
		taskRunSuccessHelp,
		labelKeys,
		nil,
	)
	m := prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		metricValue,
		labelValues...,
	)
	return &m, nil
}

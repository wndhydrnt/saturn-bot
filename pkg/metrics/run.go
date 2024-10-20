package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RunTaskSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Help: "Status of the last run of a task. 1 indicates a successful run. 0 indicates a failed run.",
			Name: "saturn_bot_run_task_success",
		},
		[]string{"task"},
	)
	RunFinish = prometheus.NewGauge(prometheus.GaugeOpts{
		Help: "Last unix time when the most recent run finished.",
		Name: "saturn_bot_run_finish_time_seconds",
	})
	RunStart = prometheus.NewGauge(prometheus.GaugeOpts{
		Help: "Last unix time when the most recent run started.",
		Name: "saturn_bot_run_start_time_seconds",
	})
)

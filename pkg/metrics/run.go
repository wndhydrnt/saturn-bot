package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RunTaskSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Help: "Status of the last run of a task. 1 indicates success. 0 indicates failure.",
			Name: "run_task_success",
		},
		[]string{"task"},
	)
	RunFinish = prometheus.NewGauge(prometheus.GaugeOpts{
		Help: "Last unix time when the run finished.",
		Name: "run_finish_time_seconds",
	})
	RunStart = prometheus.NewGauge(prometheus.GaugeOpts{
		Help: "Last unix time when the run started.",
		Name: "run_start_time_seconds",
	})
)

package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	GitCommandsDurationSecondsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Help: "Total number of commands executed by git.",
		Name: "git_commands_duration_seconds_count",
	}, []string{"command"})
	GitCommandsDurationSecondsSum = prometheus.NewCounterVec(prometheus.CounterOpts{
		Help: "Total duration it took for git to execute commands.",
		Name: "git_commands_duration_seconds_sum",
	}, []string{"command"})
)

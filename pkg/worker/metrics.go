package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	promversioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	promversion "github.com/prometheus/common/version"
	"github.com/wndhydrnt/saturn-bot/pkg/metrics"
	"github.com/wndhydrnt/saturn-bot/pkg/version"
)

const (
	metricNs                  = "saturn_bot"
	metricSub                 = "worker"
	metricLabelOpGetWorkV1    = "GetWorkV1"
	metricLabelOpReportWorkV1 = "ReportWorkV1"
)

var (
	metricRunsFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_runs_failed_total",
		Help: "Total number of runs processed by this worker that failed.",
	})
	metricRuns = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "worker_runs",
		Help: "Current number of runs being processed in parallel.",
	})
	metricRunsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "worker_runs_total",
		Help: "Total number of runs processed by this worker.",
	})
	metricRunsMax = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "worker_runs_max",
		Help: "Maximum number of runs that can be processed in parallel.",
	})
	metricServerRequestsFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_server_requests_failed_total",
		Help: "A counter that increases on failed requests to the server component. Splits by operation.",
	}, []string{"op"})
)

func initMetrics() {
	promversion.Version = version.Version
	promversion.Revision = version.Hash
	promversion.BuildDate = version.DateTime
	prometheus.DefaultRegisterer.MustRegister(
		promversioncollector.NewCollector("worker"),
		metricRunsFailed,
		metricRuns,
		metricRunsMax,
		metricServerRequestsFailed,
	)
	metrics.Register(prometheus.DefaultRegisterer)
	metricRunsFailed.Add(0)
	metricRunsTotal.Set(0)
	metricRuns.Set(0)
	metricRunsMax.Set(0)
}

package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	promversioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	promversion "github.com/prometheus/common/version"
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
		Namespace: metricNs,
		Subsystem: metricSub,
		Name:      "runs_failed_total",
		Help:      "Number of runs processed by this worker that failed.",
	})
	metricRuns = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: metricNs,
		Subsystem: metricSub,
		Name:      "runs",
		Help:      "Current number of runs being processed in parallel.",
	})
	metricRunsMax = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: metricNs,
		Subsystem: metricSub,
		Name:      "runs_max",
		Help:      "Maximum number of runs that can be processed in parallel.",
	})
	metricServerRequestsFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNs,
		Subsystem: metricSub,
		Name:      "server_requests_failed_total",
		Help:      "A counter that increases on failed requests to the server component. Splits by operation.",
	}, []string{"op"})
)

func initMetrics() {
	promversion.Version = version.Version
	promversion.Revision = version.Hash
	promversion.BuildDate = version.DateTime
	prometheus.DefaultRegisterer.MustRegister(
		promversioncollector.NewCollector("saturn_bot_worker"),
		metricRunsFailed,
		metricRuns,
		metricRunsMax,
		metricServerRequestsFailed,
	)
	metricRunsFailed.Add(0)
	metricRuns.Set(0)
	metricRunsMax.Set(0)
}

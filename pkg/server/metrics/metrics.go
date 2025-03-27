package metrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	promversioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promversion "github.com/prometheus/common/version"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"github.com/wndhydrnt/saturn-bot/pkg/version"
	"go.uber.org/zap"
)

func Init(registry prometheus.Registerer, dbInfo *service.DbInfo) {
	promversion.Version = version.Version
	promversion.Revision = version.Hash
	promversion.BuildDate = version.DateTime
	registry.MustRegister(promversioncollector.NewCollector("server"))

	registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Help: "Size of the sqlite database file in bytes. A value of -1 indicates an error during collection.",
			Name: "db_size_bytes",
		},
		newDbSizeCollectorFunc(dbInfo),
	))
}

// RegisterPrometheusRouteOpts defines all options accepted by [RegisterPrometheusRoute].
type RegisterPrometheusRouteOpts struct {
	PrometheusGatherer   prometheus.Gatherer
	PrometheusRegisterer prometheus.Registerer
	Router               chi.Router
}

// RegisterPrometheusRoute registers the handler that exposes Prometheus metrics.
func RegisterPrometheusRoute(opts RegisterPrometheusRouteOpts) {
	opts.Router.Handle(
		"/metrics",
		promhttp.InstrumentMetricHandler(
			opts.PrometheusRegisterer,
			promhttp.HandlerFor(opts.PrometheusGatherer, promhttp.HandlerOpts{}),
		),
	)
}

func newDbSizeCollectorFunc(dbInfo *service.DbInfo) func() float64 {
	return func() float64 {
		size, err := dbInfo.Size()
		if err != nil {
			log.Log().Errorw("Failed to collect database size", zap.Error(err))
			return -1
		}

		return size
	}
}

package metrics

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	promversioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promversion "github.com/prometheus/common/version"
	"github.com/wndhydrnt/saturn-bot/pkg/version"
)

func Init(registry prometheus.Registerer) {
	promversion.Version = version.Version
	promversion.Revision = version.Hash
	promversion.BuildDate = version.DateTime
	registry.MustRegister(promversioncollector.NewCollector("saturn_bot_server"))
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

package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	hostLabel = "host"
)

var (
	httpClientRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Help: "Total number of requests sent via HTTP clients.",
			Name: "saturn_bot_http_client_requests_total",
		},
		[]string{"code", "method", hostLabel},
	)
)

type hostLabelCtxKey struct{}

func newRoundTripper(next http.RoundTripper) promhttp.RoundTripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		ctx := context.WithValue(req.Context(), hostLabelCtxKey{}, req.Host)
		return next.RoundTrip(req.WithContext(ctx))
	}
}

func InstrumentHttpClient(c *http.Client) {
	opts := promhttp.WithLabelFromCtx(hostLabel,
		func(ctx context.Context) string {
			return ctx.Value(hostLabelCtxKey{}).(string)
		},
	)

	roundTripper := newRoundTripper(
		promhttp.InstrumentRoundTripperCounter(httpClientRequestsTotal, c.Transport, opts),
	)
	c.Transport = roundTripper
}

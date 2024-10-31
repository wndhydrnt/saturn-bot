package metrics

import "github.com/prometheus/client_golang/prometheus"

func Register(reg prometheus.Registerer) {
	reg.MustRegister(
		GitCommandsDurationSecondsCount,
		GitCommandsDurationSecondsSum,
		httpClientRequestsTotal,
		RunTaskSuccess,
		RunFinish,
		RunStart,
	)
}

package prometheus_data

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of http requests",
		},
		[]string{"method", "endpoint", "status"},
	)
)

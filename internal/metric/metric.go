package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace = "grimd"
)

var (
	CustomRecordCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "custom_records",
			Help:      "Amount of custom resource records configured at startup",
		},
	)

	RequestCustomCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "request_custom",
		Help:      "Served custom DNS requests",
	})

	ResponseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "request_total",
			Help:      "Served DNS replies",
		}, []string{"q_type", "remote_ip", "q_name", "rcode"})

	RequestUpstreamResolveCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "upstream_request",
			Help:      "Upstream DNS requests",
		}, []string{"q_type", "q_name", "rcode", "upstream"})

	RequestUpstreamDohRequest = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "upstream_request_doh",
			Help:      "Upstream DoH requests - only works when DoH configured",
		}, []string{"success"})
)

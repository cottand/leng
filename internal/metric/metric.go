package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace = "grimd"
)

var (
	DNSRequestCustomCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "request_custom",
		Help:      "Served custom DNS requests",
	})

	DNSResponseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "request_total",
			Help:      "Served DNS replies",
		}, []string{"q_type", "remote_ip", "q_name", "rcode"})

	DNSRequestUpstreamResolveCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "upstream_request",
			Help:      "Upstream DNS requests",
		}, []string{"q_type", "q_name", "rcode", "upstream"})

	DNSRequestUpstreamDohRequest = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "upstream_request_doh",
			Help:      "Upstream DoH requests - only works when DoH configured",
		}, []string{"success"})
)

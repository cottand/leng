package metric

import (
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net"
	"strconv"
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

	responseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "request_total",
			Help:      "Served DNS replies",
		}, []string{"q_type", "remote_ip", "q_name", "rcode", "blocked"})

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

func init() {
	prometheus.MustRegister(
		responseCounter,
		RequestUpstreamResolveCounter,
		RequestUpstreamDohRequest,
	)
}

func ReportDNSResponse(w dns.ResponseWriter, message *dns.Msg, blocked bool) {
	question := message.Question[0]
	remoteHost, _, _ := net.SplitHostPort(w.RemoteAddr().String())
	responseCounter.With(prometheus.Labels{
		"remote_ip": remoteHost,
		"q_type":    dns.Type(question.Qtype).String(),
		"q_name":    question.Name,
		"rcode":     dns.RcodeToString[message.Rcode],
		"blocked":   strconv.FormatBool(blocked),
	}).Inc()
}

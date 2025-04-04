package metric

import (
	"context"
	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net"
	"strconv"
	"time"
)

const (
	Namespace = "leng"
)

var (
	CustomRecordCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "custom_records",
			Help:      "Amount of custom resource records configured in config",
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

	cachedResponseCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "request_cached",
			Help:      "Cached DNS replies",
		},
	)

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

	RequestUpstreamResolveDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:                      Namespace,
		Name:                           "upstream_request_duration",
		Help:                           "Upstream requests duration in seconds, by request type",
		Buckets:                        []float64{0.0001, 0.0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		NativeHistogramBucketFactor:    1.1,
		NativeHistogramMaxBucketNumber: 32,
	}, []string{"upstream_type"})

	CustomDNSConfigReload = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "config_reload_customdns",
			Help:      "Custom DNS config reloads",
		})

	DohResponseCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "doh_response_count",
			Help:      "Successful DoH responses",
		}, []string{"status"})

	allVecMetrics = []*prometheus.CounterVec{
		responseCounter,
		RequestUpstreamResolveCounter,
		RequestUpstreamDohRequest,
		DohResponseCount,
	}

	configHighCardinality = false
	configHistograms      = false
)

func init() {
	prometheus.MustRegister(
		responseCounter,
		RequestUpstreamResolveCounter,
		RequestUpstreamDohRequest,
		CustomDNSConfigReload,
		DohResponseCount,
		cachedResponseCounter,
	)
}

func Start(resetPeriodMinutes int64, highCardinality bool, histogramsEnabled bool) (closeChan context.CancelFunc) {
	configHighCardinality = highCardinality
	if histogramsEnabled {
		prometheus.MustRegister(RequestUpstreamResolveDuration)
		configHistograms = true
	}
	ctx, cancel := context.WithCancel(context.Background())
	mark := time.Now()

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(resetPeriodMinutes) * time.Minute)
				if time.Now().Sub(mark) > time.Duration(resetPeriodMinutes)*time.Minute {
					for _, m := range allVecMetrics {
						m.Reset()
					}
					mark = time.Now()
				}
			}
			time.Sleep(400 * time.Millisecond)
		}

	}(ctx)

	return cancel
}

func ReportDNSResponse(w dns.ResponseWriter, message *dns.Msg, blocked bool) {
	question := message.Question[0]
	var remoteHost string
	var qName string
	if !configHighCardinality {
		remoteHost = ""
		qName = ""
	} else {
		remoteHost, _, _ = net.SplitHostPort(w.RemoteAddr().String())
		qName = question.Name
	}
	responseCounter.With(prometheus.Labels{
		"remote_ip": remoteHost,
		"q_type":    dns.Type(question.Qtype).String(),
		"q_name":    qName,
		"rcode":     dns.RcodeToString[message.Rcode],
		"blocked":   strconv.FormatBool(blocked),
	}).Inc()
}

func ReportDNSRespond(remote net.IP, message *dns.Msg, blocked bool, cached bool) {
	question := message.Question[0]
	var remoteHost string
	var qName string
	if !configHighCardinality {
		remoteHost = ""
		qName = ""
	} else {
		remoteHost = remote.String()
		qName = question.Name
	}
	responseCounter.With(prometheus.Labels{
		"remote_ip": remoteHost,
		"q_type":    dns.Type(question.Qtype).String(),
		"q_name":    qName,
		"rcode":     dns.RcodeToString[message.Rcode],
		"blocked":   strconv.FormatBool(blocked),
	}).Inc()
	if cached {
		cachedResponseCounter.Inc()
	}
}

func ReportUpstreamResolve(upstreamType string, duration time.Duration) {
	if !configHistograms {
		return
	}
	RequestUpstreamResolveDuration.
		With(prometheus.Labels{"upstream_type": upstreamType}).
		Observe(duration.Seconds())
}

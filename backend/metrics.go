package backend

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metric declarations adapted from promhttp code to approximate the removed
// prometheus.InstrumentHandlerFunc().

type httpMetricsStruct struct {
	inFlight     prometheus.Gauge
	counter      *prometheus.CounterVec
	duration     *prometheus.SummaryVec
	requestSize  *prometheus.SummaryVec
	responseSize *prometheus.SummaryVec
}

func newHttpMetricsStruct() *httpMetricsStruct {
	m := &httpMetricsStruct{}

	m.inFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "http",
		Name:      "in_flight_requests",
		Help:      "Number HTTP requests currently in flight.",
	})

	m.counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests made.",
		},
		[]string{"handler", "code", "method"},
	)

	m.duration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "request_duration_seconds",
			Help:       "HTTP request latencies in seconds.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"handler", "code", "method"},
	)

	m.requestSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "request_size_bytes",
			Help:       "HTTP request sizes in bytes.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"handler", "code", "method"},
	)

	m.responseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "response_size_bytes",
			Help:       "HTTP response sizes in bytes.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"handler", "code", "method"},
	)

	prometheus.MustRegister(m.inFlight, m.counter, m.duration, m.requestSize, m.responseSize)

	return m
}

var httpMetrics *httpMetricsStruct

// not thread-safe (why are you trying to create multiple HTTP servers in parallel?)
func instrumentHttpHandlerFunc(label string, h http.HandlerFunc) http.HandlerFunc {
	if (httpMetrics == nil) {
		httpMetrics = newHttpMetricsStruct()
	}
	m := httpMetrics

	return promhttp.InstrumentHandlerInFlight(m.inFlight,
		promhttp.InstrumentHandlerCounter(m.counter.MustCurryWith(prometheus.Labels{"handler": label}),
			promhttp.InstrumentHandlerDuration(m.duration.MustCurryWith(prometheus.Labels{"handler": label}),
				promhttp.InstrumentHandlerRequestSize(m.requestSize.MustCurryWith(prometheus.Labels{"handler": label}),
					promhttp.InstrumentHandlerResponseSize(m.responseSize.MustCurryWith(prometheus.Labels{"handler": label}), h),
				),
			),
		),
	).ServeHTTP
}

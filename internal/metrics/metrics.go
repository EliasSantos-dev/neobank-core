// Package metrics instrumenta o serviço no padrão Prometheus.
package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	reg        *prometheus.Registry
	requests   *prometheus.CounterVec
	duration   *prometheus.HistogramVec
	operations *prometheus.CounterVec
}

func New() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		reg: reg,
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "neobank_http_requests_total", Help: "Total de requisições HTTP.",
		}, []string{"method", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "neobank_http_request_duration_seconds", Help: "Duração das requisições HTTP.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method"}),
		operations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "neobank_operations_total", Help: "Total de operações de negócio.",
		}, []string{"op"}),
	}
	reg.MustRegister(m.requests, m.duration, m.operations)
	return m
}

func (m *Metrics) IncOperation(op string) { m.operations.WithLabelValues(op).Inc() }

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Middleware mede duração e conta requisições por método+status.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(m.duration.WithLabelValues(r.Method))
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		timer.ObserveDuration()
		m.requests.WithLabelValues(r.Method, strconv.Itoa(rec.status)).Inc()
	})
}

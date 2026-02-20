package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dataease_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dataease_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "dataease_active_connections",
			Help: "Number of active connections",
		},
	)

	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dataease_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dataease_cache_hits_total",
			Help: "Total number of cache hits and misses",
		},
		[]string{"result"},
	)
)

func RecordRequest(method, path, status string, duration float64) {
	HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
	HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

func RecordDbQuery(operation, table string, duration float64) {
	DbQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

func RecordCacheHit(isHit bool) {
	if isHit {
		CacheHits.WithLabelValues("hit").Inc()
	} else {
		CacheHits.WithLabelValues("miss").Inc()
	}
}

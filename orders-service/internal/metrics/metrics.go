package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	CacheHits   prometheus.Counter
	CacheMisses prometheus.Counter

	DBGetDuration  prometheus.Histogram
	DBSaveDuration prometheus.Histogram

	KafkaMessages prometheus.Counter
	KafkaBad      prometheus.Counter
	KafkaErrors   prometheus.Counter
}

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		CacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total cache hits",
		}),
		CacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total cache misses",
		}),
		DBGetDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "db_get_order_duration_seconds",
			Help:    "DB GetByID duration",
			Buckets: prometheus.DefBuckets,
		}),
		DBSaveDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "db_save_order_duration_seconds",
			Help:    "DB Save duration",
			Buckets: prometheus.DefBuckets,
		}),
		KafkaMessages: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "Total Kafka messages successfully processed",
		}),
		KafkaBad: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_bad_messages_total",
			Help: "Total bad Kafka messages (skipped)",
		}),
		KafkaErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_processing_errors_total",
			Help: "Total Kafka processing errors (no commit)",
		}),
	}

	reg.MustRegister(
		m.CacheHits, m.CacheMisses,
		m.DBGetDuration, m.DBSaveDuration,
		m.KafkaMessages, m.KafkaBad, m.KafkaErrors,
	)
	return m
}

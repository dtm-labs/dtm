package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// https://yunlzheng.gitbook.io/prometheus-book/parti-prometheus-ji-chu/promql/what-is-prometheus-metrics-and-labels
// Metric: <metric name>{<label name>=<label value>, ...}
var (
	globalTransactionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dtm_global_transaction_total",
		Help: "Count all transactions processed by dtm with diff model and status",
	},
		[]string{"role", "model", "status"})

	globalTransactionCost = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "dtm_global_transaction_cost",
		Help: "Count all transactions request cost by dtm with diff model and status",
	},
		[]string{"role", "model", "status"})
)
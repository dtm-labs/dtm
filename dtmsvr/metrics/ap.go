package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	appTransactionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_transaction_total",
		Help: "All transaction submitted by app",
	},
		[]string{"role", "name", "model", "status"})

	appTransactionCost = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "app_transaction_cost",
		Help: "All transaction cost by app",
	},
		[]string{"role", "name", "model", "status"})
)

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	branchTransactionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rm_branch_transaction_total",
		Help: "All branch transaction received by rm",
	},
		[]string{"role", "name", "model", "status"})

	branchTransactionCost = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "rm_branch_transaction_cost",
		Help: "All branch transaction cost by rm",
	},
		[]string{"role", "name", "model", "status"})
)

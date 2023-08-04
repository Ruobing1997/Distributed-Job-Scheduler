package task_executor

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	tasksTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tasks_total",
		Help: "The total number of tasks received",
	})

	tasksSuccessTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tasks_success_total",
		Help: "The total number of tasks executed successfully",
	})

	tasksFailedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tasks_failed_total",
		Help: "The total number of tasks that failed execution",
	})

	postgresqlOpsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "postgresql_interactions_total",
		Help: "The total number of interactions with PostgreSQL",
	}, []string{"operation", "table"})

	grpcOpsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "The total number of gRPC requests sent",
	}, []string{"method", "sender"})
)

func PrometheusManagerInit() {
	prometheus.MustRegister(tasksTotal)
	prometheus.MustRegister(tasksSuccessTotal)
	prometheus.MustRegister(tasksFailedTotal)
	prometheus.MustRegister(postgresqlOpsTotal)
	prometheus.MustRegister(grpcOpsTotal)
}

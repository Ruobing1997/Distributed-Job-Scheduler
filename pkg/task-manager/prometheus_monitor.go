package task_manager

import "github.com/prometheus/client_golang/prometheus"

var (
	tasksTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tasks_total",
		Help: "Total number of tasks.",
	})
	succeedTasksTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "succeed_tasks_total",
		Help: "Total number of succeed tasks.",
	})
	failedTasksTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "failed_tasks_total",
		Help: "Total number of failed tasks.",
	})

	dispatchTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dispatch_total",
		Help: "Total number of dispatches.",
	})

	redisThroughput = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "redis_throughput_total",
		Help: "Total throughput of Redis operations.",
	})
	postgresqlOpsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "postgresql_throughput_total",
			Help: "Total throughput of PostgreSQL operations.",
		},
		[]string{"operation", "table"},
	)
	grpcOpsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_throughput_total",
			Help: "Total throughput of gRPC operations.",
		},
		[]string{"operation", "sender"},
	)
)

func PrometheusManagerInit() {
	prometheus.MustRegister(tasksTotal)
	prometheus.MustRegister(dispatchTotal)
	prometheus.MustRegister(failedTasksTotal)
	prometheus.MustRegister(succeedTasksTotal)
	prometheus.MustRegister(redisThroughput)
	prometheus.MustRegister(postgresqlOpsTotal)
	prometheus.MustRegister(grpcOpsTotal)
}

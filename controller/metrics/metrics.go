package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	ReconcilesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consoleapplication_reconcile_total",
			Help: "Number of total reconciliation attempts for ConsoleApplication",
		}, []string{"namespace", "name"},
	)

	GitRepoReachableDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "consoleapplication_git_repo_reachable_duration_seconds",
			Help:    "Duration of Git repository reachability check",
			Buckets: prometheus.DefBuckets,
		}, []string{"namespace", "name"},
	)

	ConsoleApplicationsProcessing = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consoleapplication_crs_processing_gauge",
			Help: "Number of ConsoleApplication objects being processed",
		}, []string{"namespace"},
	)

	ConsoleApplicationsSuccessTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consoleapplication_crs_success_total",
			Help: "Number of successful ConsoleApplication creations",
		}, []string{"namespace"},
	)

	ResourcesCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consoleapplication_resources_created_total",
			Help: "Number of resources created for ConsoleApplication",
		}, []string{"namespace", "name", "kind"},
	)
)

func init() {
	metrics.Registry.MustRegister(ReconcilesTotal)
	metrics.Registry.MustRegister(GitRepoReachableDuration)
}

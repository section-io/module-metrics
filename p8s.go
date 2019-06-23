package metrics

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricsAddress     = "0.0.0.0"
	defaultMetricsPath = "/metrics"
	defaultMetricsPort = "9000"
	promeSubsystem     = "http"
)

var (
	requestsTotal *prometheus.CounterVec
	bytesTotal    *prometheus.CounterVec
	registry      *prometheus.Registry

	p8sLabels = []string{"hostname", "status"}

	// MetricsURI is the address the prometheus server is listening on
	MetricsURI string
)

func addRequest(hostname string, status string, bytes int) {
	requestsTotal.With(prometheus.Labels{"hostname": hostname, "status": status}).Inc()
	bytesTotal.With(prometheus.Labels{"hostname": hostname, "status": status}).Add(float64(bytes))
}

// initMetrics sets up the prometheus registry and creates the metrics. Calling this
// will reset any collected metrics
func initMetrics(promeNamespace string) {
	registry = prometheus.NewRegistry()

	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: promeSubsystem,
		Name:      "request_count_total",
		Help:      "Total count of HTTP requests.",
	}, p8sLabels)

	bytesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: promeSubsystem,
		Name:      "bytes_total",
		Help:      "Total sum of response bytes.",
	}, p8sLabels)

	registry.MustRegister(requestsTotal, bytesTotal)

}

// StartPrometheusServer starts the prometheus HTTP server
func StartPrometheusServer(stderr io.Writer) {

	metricsPath := os.Getenv("P8S_METRICS_PATH")
	if metricsPath == "" {
		metricsPath = defaultMetricsPath
	}

	metricsPort := os.Getenv("P8S_METRICS_PORT")
	if metricsPort == "" {
		metricsPort = defaultMetricsPort
	}

	MetricsURI = fmt.Sprintf("http://%s:%s%s", metricsAddress, metricsPort, metricsPath)

	http.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	go func() {
		_, _ = fmt.Fprintf(stderr, "Listening on %s\n", MetricsURI)
		if err := http.ListenAndServe(metricsAddress+":"+metricsPort, nil); err != nil {
			log.Fatalf("[ERROR] failed to start HTTP server: %v\n", err)
		}
	}()
}

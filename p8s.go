package metrics

import (
	"context"
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
	jsonParseErrorTotal prometheus.Counter
	requestsTotal       *prometheus.CounterVec
	bytesTotal          *prometheus.CounterVec
	registry            *prometheus.Registry
	httpServer          *http.Server

	defaultP8sLabels   = []string{"hostname"}
	logFieldNames      []string
	sanitizedP8sLabels []string

	p8sHTTPServerStarted = false

	// MetricsURI is the address the prometheus server is listening on
	MetricsURI string
)

func addRequest(labels map[string]string, bytes int) {
	requestsTotal.With(labels).Inc()
	bytesTotal.With(labels).Add(float64(bytes))
}

// InitMetrics sets up the prometheus registry and creates the metrics. Calling this
// will reset any collected metrics. Returns the registry so additional metrics can be registered.
func InitMetrics(additionalLabels ...string) *prometheus.Registry {

	logFieldNames = append(defaultP8sLabels, additionalLabels...)

	sanitizedP8sLabels = defaultP8sLabels
	for _, label := range additionalLabels {
		label, _ = sanitizeLabel(label, "dummy")
		sanitizedP8sLabels = append(sanitizedP8sLabels, label)
	}

	const promeNamespace = "section"
	registry = prometheus.NewRegistry()

	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: promeSubsystem,
		Name:      "request_count_total",
		Help:      "Total count of HTTP requests.",
	}, sanitizedP8sLabels)

	bytesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: promeSubsystem,
		Name:      "bytes_total",
		Help:      "Total sum of response bytes.",
	}, sanitizedP8sLabels)

	jsonParseErrorTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: promeSubsystem,
		Name:      "json_parse_errors_total",
		Help:      "Total count of JSON parsing errors.",
	})

	registry.MustRegister(requestsTotal, bytesTotal, jsonParseErrorTotal)

	go startPrometheusServer(os.Stderr)

	return registry
}

func startPrometheusServer(stderr io.Writer) {

	if p8sHTTPServerStarted {
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Fatalf("Failed to shutdown HTTP server: %v\n", err)
		}
	}

	metricsPath := os.Getenv("P8S_METRICS_PATH")
	if metricsPath == "" {
		metricsPath = defaultMetricsPath
	}

	metricsPort := os.Getenv("P8S_METRICS_PORT")
	if metricsPort == "" {
		metricsPort = defaultMetricsPort
	}

	MetricsURI = fmt.Sprintf("http://%s:%s%s", metricsAddress, metricsPort, metricsPath)

	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	httpServer = &http.Server{
		Addr:    metricsAddress + ":" + metricsPort,
		Handler: mux,
	}

	p8sHTTPServerStarted = true
	_, _ = fmt.Fprintf(stderr, "Listening on %s\n", MetricsURI)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[ERROR] failed to start HTTP server: %v\n", err)
	}
}

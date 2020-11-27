package metrics

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	pageViewTotal       prometheus.Counter
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

	// vars related to limiting the number of unique hostname labels
	uniqueHostnameMap  = make(map[string]struct{})
	maxUniqueHostnames = 1000
)

func isPageView(logline map[string]interface{}) bool {
	// Count text/html 2XX requests as page-views
	return strings.HasPrefix(fmt.Sprintf("%v", logline["status"]), "2") && strings.HasPrefix(strings.ToLower(fmt.Sprintf("%v", logline["content_type"])), "text/html")
}

func addRequest(labels map[string]string, logline map[string]interface{}) {
	_, ok := uniqueHostnameMap[labels["hostname"]]
	if !ok {
		if len(uniqueHostnameMap) < maxUniqueHostnames {
			uniqueHostnameMap[labels["hostname"]] = struct{}{}
		} else {
			// Use hard-coded hostname so wildcard domains don't make cardinality explode.
			labels["hostname"] = "max-hostnames-reached"
		}
	}

	bytes := getBytes(logline)

	requestsTotal.With(labels).Inc()
	bytesTotal.With(labels).Add(float64(bytes))

	if isPageView(logline) {
		pageViewTotal.Inc()
	}
}

// InitMetrics sets up the prometheus registry and creates the metrics. Calling this
// will reset any collected metrics. Returns the registry so additional metrics can be registered.
func InitMetrics(additionalLabels ...string) *prometheus.Registry {

	logFieldNames = append(defaultP8sLabels, additionalLabels...)

	sanitizedP8sLabels = defaultP8sLabels
	for _, label := range additionalLabels {
		label = sanitizeLabelName(label)
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

	pageViewTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: promeSubsystem,
		Name:      "page_view_total",
		Help:      "Legacy: Total count of page views.",
	})

	jsonParseErrorTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: promeSubsystem,
		Name:      "json_parse_errors_total",
		Help:      "Total count of JSON parsing errors.",
	})

	registry.MustRegister(requestsTotal, bytesTotal, pageViewTotal, jsonParseErrorTotal)

	maxUniqueHostnamesStr := os.Getenv("MODULE_METRICS_MAX_HOSTNAMES")
	if maxUniqueHostnamesStr != "" {
		maxUniqueHostnamesInt, err := strconv.Atoi(maxUniqueHostnamesStr)
		if err == nil {
			maxUniqueHostnames = maxUniqueHostnamesInt
			log.Printf("[DEBUG] Using %d for maxUniqueHostnames\n", maxUniqueHostnames)
		}
	}

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

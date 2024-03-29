package metrics

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	gojsonq "github.com/thedevsaddam/gojsonq/v2"
	"golang.org/x/exp/slices"
)

const (
	metricsAddress     = "0.0.0.0"
	defaultMetricsPath = "/metrics"
	defaultMetricsPort = "9000"
	promeSubsystem     = "http"
	promeNamespace     = "section"
	hostnameLabel      = "hostname"

	aeeHealthcheckLabel = "section_aee_healthcheck"
)

var (
	jsonParseErrorTotal prometheus.Counter
	pageViewTotal       prometheus.Counter
	requestsTotal       *prometheus.CounterVec
	bytesTotal          *prometheus.CounterVec
	registry            *prometheus.Registry
	httpServer          *http.Server

	requestsByHostnameTotal *prometheus.CounterVec
	bytesByHostnameTotal    *prometheus.CounterVec

	logFieldNames      []string
	sanitizedP8sLabels []string
	withGeoLabel       []string
	requestLabels      []string

	p8sHTTPServerStarted = false

	// MetricsURI is the address the prometheus server is listening on
	MetricsURI string

	// vars related to limiting the number of unique hostname labels
	uniqueHostnameMap  = make(map[string]struct{})
	maxUniqueHostnames = 1000

	includeHostnameMetrics = false

	aeeUserAgentRegex = regexp.MustCompile(`^aee/v.+`)
)

// Logf is a type that can be provided for outputing logs to specifi stream
type Logf func(format string, v ...interface{})

// ShowLabels outputs via the log function the "effective" labels
// which can optionally include "geo_hash".  The intent is to
// reconcile the labels used for initialization vs the ones output
// when extracting metrics during some kind of issue
func ShowLabels(log Logf) {
	log("[INFO] logFieldNames %+v", logFieldNames)
	log("[INFO] sanitizedP8sLabels %+v", sanitizedP8sLabels)
	log("[INFO] withGeoLabel %+v", withGeoLabel)
	log("[INFO] requestLabels %+v", requestLabels)
}

func extractUserAgent(logline map[string]interface{}) string {
	userAgent := gojsonq.New().FromInterface(logline).Find("request.http_user_agent")
	if userAgent != nil {
		if userAgentStr, ok := userAgent.(string); ok {
			return userAgentStr
		}
	}
	return ""
}

func isPageView(logline map[string]interface{}) bool {
	// Count text/html 2XX requests as page-views
	return strings.HasPrefix(fmt.Sprintf("%v", logline["status"]), "2") &&
		strings.HasPrefix(strings.ToLower(fmt.Sprintf("%v", logline["content_type"])), "text/html") &&
		!aeeUserAgentRegex.MatchString(extractUserAgent(logline))
}

func addRequest(labels map[string]string, logline map[string]interface{}) {

	hostname := ""
	ok := false
	if hostname, ok = labels[hostnameLabel]; ok {
		delete(labels, hostnameLabel)
		_, ok := uniqueHostnameMap[hostname]
		if !ok {
			if len(uniqueHostnameMap) < maxUniqueHostnames {
				uniqueHostnameMap[hostname] = struct{}{}
			} else {
				// Use hard-coded hostname so wildcard domains don't make cardinality explode.
				hostname = "max-hostnames-reached"
			}
		}

	}

	bytes := float64(getBytes(logline))

	requestsTotal.With(labels).Inc()

	// remove geo_hash for bytesTotal
	bytePairs := scrubGeoHash(labels)
	bytesTotal.With(bytePairs).Add(bytes)

	if isPageView(logline) {
		pageViewTotal.Inc()
	}

	if includeHostnameMetrics {
		requestsByHostnameTotal.WithLabelValues(hostname).Inc()
		bytesByHostnameTotal.WithLabelValues(hostname).Add(bytes)
	}
}

// InitMetrics sets up the prometheus registry and creates the metrics. Calling this
// will reset any collected metrics. Returns the registry so additional metrics can be registered.
func InitMetrics(additionalLabels ...string) *prometheus.Registry {
	logFieldNames = additionalLabels
	includeHostnameMetrics = false

	// iterate over any additionalLabels passed during metrics initialization & sanitize them (if we have rules defined)
	sanitizedP8sLabels = []string{}
	for _, label := range additionalLabels {
		label = sanitizeLabelName(label)
		sanitizedP8sLabels = append(sanitizedP8sLabels, label)
	}

	// If the hostname label is included, generate the by_hostname metrics and remove the hostname label from the
	// non-by_hostname metrics to reduce cardinality.
	if idx := slices.Index(sanitizedP8sLabels, hostnameLabel); idx > -1 {
		includeHostnameMetrics = true
		sanitizedP8sLabels = slices.Delete(sanitizedP8sLabels, idx, idx+1)
	}

	requestLabels = sanitizedP8sLabels
	requestLabels = append(requestLabels, aeeHealthcheckLabel)
	if isGeoHashing {
		requestLabels = append(requestLabels, geoHash)
	}

	// request labels has geo_hash only for requests counts (not bytes)
	// when geo_hash is used, bytes needs doesn't use that label
	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: promeSubsystem,
		Name:      "request_count_total",
		Help:      "Total count of HTTP requests.",
	}, requestLabels)

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

	registry = prometheus.NewRegistry()
	registry.MustRegister(requestsTotal, bytesTotal, pageViewTotal, jsonParseErrorTotal)

	if includeHostnameMetrics {
		requestsByHostnameTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: promeNamespace,
			Subsystem: promeSubsystem,
			Name:      "request_count_by_hostname_total",
			Help:      "Total count of HTTP requests by hostname.",
		}, []string{hostnameLabel})

		bytesByHostnameTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: promeNamespace,
			Subsystem: promeSubsystem,
			Name:      "bytes_by_hostname_total",
			Help:      "Total sum of response bytes by hostname.",
		}, []string{hostnameLabel})

		registry.MustRegister(requestsByHostnameTotal, bytesByHostnameTotal)
	}

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

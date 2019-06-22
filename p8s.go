package metrics

import (
	"fmt"
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
)

type counter interface {
	Inc()
	Add(float64)
}

var (
	requestsTotal counter
	bytesTotal    counter

	// MetricsURI is the address the prometheus server is listening on
	MetricsURI string
)

func addRequest(hostname string, bytes int) {
	bytesTotal.Add(float64(bytes))
	requestsTotal.Inc()
}

func initMetrics(promeNamespace string) {
	requestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: "total",
		Name:      "request_count_total",
		Help:      "Total count of HTTP requests.",
	})

	bytesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: "total",
		Name:      "bytes_total",
		Help:      "Total sum of response bytes.",
	})
}

func StartPrometheusServer() {

	metricsPath := os.Getenv("P8S_METRICS_PATH")
	if metricsPath == "" {
		metricsPath = defaultMetricsPath
	}

	metricsPort := os.Getenv("P8S_METRICS_PORT")
	if metricsPort == "" {
		metricsPort = defaultMetricsPort
	}

	MetricsURI = fmt.Sprintf("http://%s:%s%s", metricsAddress, metricsPort, metricsPath)

	prometheus.MustRegister(requestsTotal.(prometheus.Counter), bytesTotal.(prometheus.Counter))

	http.Handle(metricsPath, promhttp.Handler())
	go func() {
		log.Printf("Listening on %s\n", MetricsURI)
		if err := http.ListenAndServe(metricsAddress+":"+metricsPort, nil); err != nil {
			log.Fatalf("[ERROR] failed to start HTTP server: %v\n", err)
		}
	}()
}

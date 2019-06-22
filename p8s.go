package metrics

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultMetricsPath = "/metrics"
	defaultMetricsPort = "9000"
)

type counter interface {
	Inc()
	Add(float64)
}

var (
	requestsTotal counter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: "total",
		Name:      "request_count_total",
		Help:      "Total count of HTTP requests.",
	})

	bytesTotal counter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promeNamespace,
		Subsystem: "total",
		Name:      "bytes_total",
		Help:      "Total sum of response bytes.",
	})

	metricsPath    string
	metricsPort    string
	promeNamespace string
)

func addRequest(hostname string, bytes int) {
	bytesTotal.Add(float64(bytes))
	requestsTotal.Inc()
}

func StartPrometheusServer(moduleName string) {

	promeNamespace = moduleName

	metricsPath = os.Getenv("P8S_METRICS_PATH")
	if metricsPath == "" {
		metricsPath = defaultMetricsPath
	}

	metricsPort = os.Getenv("P8S_METRICS_PORT")
	if metricsPort == "" {
		metricsPort = defaultMetricsPort
	}

	prometheus.MustRegister(requestsTotal.(prometheus.Counter), bytesTotal.(prometheus.Counter))

	http.Handle(metricsPath, promhttp.Handler())
	go func() {
		log.Printf("Listening on http://0.0.0.0:%s%s\n", metricsPort, metricsPath)
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			log.Fatalf("[ERROR] failed to start HTTP server: %v\n", err)
		}
	}()
}

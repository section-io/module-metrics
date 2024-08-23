package metrics

import (
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const statusLabel = "status"
const statusBucketLabel = "status_bucket"
const cacheHandlingLabel = "cache_handling"
const varnishHandlingLabel = "varnish_handling"

type requestTimeProcessor struct {
	vec      *prometheus.HistogramVec
	logField LogFieldValueFunc
	logUnit  SecondsMultiplier
}

type noopRequestTimeProcessor struct{}

func (p *noopRequestTimeProcessor) Observe(_ map[string]string, _ map[string]interface{}) {}

func newRequestTimeProcessor(vec *prometheus.HistogramVec, logField LogFieldValueFunc, logUnit SecondsMultiplier) *requestTimeProcessor {
	return &requestTimeProcessor{
		vec:      vec,
		logField: logField,
		logUnit:  logUnit,
	}
}

func (p *requestTimeProcessor) Observe(requestLabels map[string]string, logline map[string]interface{}) {
	subject := p.logField(logline)
	floatValue, err := strconv.ParseFloat(fmt.Sprintf("%v", subject), 64)
	if err != nil {
		return
	}
	floatValue = floatValue * float64(p.logUnit)

	labels := make(map[string]string, len(requestLabels))
	for label, labelValue := range requestLabels {
		histogramLabel := translateRequestLabelToHistogramLabel(label)
		if histogramLabel == "" {
			continue
		}
		switch histogramLabel {
		case statusBucketLabel:
			labels[histogramLabel] = statusBucket(labelValue)
		case cacheHandlingLabel:
			labels[histogramLabel] = translateCacheHandling(labelValue)
		}
	}

	p.vec.With(labels).Observe(floatValue)
}

func translateRequestLabelToHistogramLabel(requestLabel string) string {
	switch requestLabel {

	// translations
	case statusLabel:
		return statusBucketLabel
	case varnishHandlingLabel:
		// Varnish should emit metrics consistent with other caches
		return cacheHandlingLabel

	// allow list
	case statusBucketLabel:
		return requestLabel
	case cacheHandlingLabel:
		return requestLabel

	}
	return ""
}

func translateCacheHandling(handlingLogValue string) string {
	// allow list
	switch handlingLogValue {
	case "hit":
		return handlingLogValue
	case "miss":
		return handlingLogValue
	case "pass":
		return handlingLogValue
	case "synth":
		return handlingLogValue
	case "pipe":
		return handlingLogValue
	}
	return ""
}

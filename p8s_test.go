package metrics

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
)

func TestIsPageView(t *testing.T) {
	var logline = map[string]interface{}{"content_type": "text/html", "status": "200"}
	assert.True(t, isPageView(logline))

	logline = map[string]interface{}{"content_type": "text/html", "status": "201"}
	assert.True(t, isPageView(logline))

	logline = map[string]interface{}{"content_type": "text/css", "status": "200"}
	assert.False(t, isPageView(logline))

	logline = map[string]interface{}{"content_type": "text/html", "status": "404"}
	assert.False(t, isPageView(logline))

	logline = map[string]interface{}{"content_type": "text/html;charset=UTF-8", "status": "200"}
	assert.True(t, isPageView(logline))
}

func getP8sHTTPResponse(t *testing.T) string {
	resp, err := http.Get(MetricsURI)
	if err != nil {
		t.Errorf("Error getting %s: %#v", MetricsURI, err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading body: %#v", err)
	}

	return string(bodyBytes)
}

func gatherP8sResponse(t *testing.T) string {
	gathering, err := registry.Gather()
	if err != nil {
		t.Errorf("Error in prometheus Gather: %#v", err)
	}

	out := &bytes.Buffer{}
	for _, mf := range gathering {
		if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
			t.Errorf("Error in expfmt.MetricFamilyToText: %#v", err)
		}
	}

	return out.String()
}

func testLogsOutputEqualsInput(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip"}`,
	}

	InitMetrics()

	writeLogs(t, logs)

	outputLines := strings.Split(strings.TrimSpace(stdout.String()), "\n")

	assert.Equal(t, logs, outputLines)
}

func testCountersIncrease(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"foo.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"4e189f278375962cd19d380562846296"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"6ef1b5083893627d2426e42206d78f70"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.077","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"1959","body_bytes_sent":"1498","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.077","upstream_connect_time":"0.057","upstream_response_time":"0.077","upstream_response_length":"1498","upstream_bytes_received":"1889","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"b1ea9bc0be7edfc997bc18a9f6b20d68"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.073","hostname":"foo.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.073","upstream_connect_time":"0.055","upstream_response_time":"0.073","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"ff3117bb0ac0307d8d0e78fc8b8ba5c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"85e833ae62745c50492c80b4d7b78016"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.072","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.072","upstream_connect_time":"0.054","upstream_response_time":"0.072","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"a095e3c2c3a0f25b4bbca4c941babd76"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.071","hostname":"bar.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.071","upstream_connect_time":"0.053","upstream_response_time":"0.071","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"8fb3941b35418bdfa1946ef02c90e8c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"2126","body_bytes_sent":"1665","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"1665","upstream_bytes_received":"2056","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"789addb393a18ff1caf5d776b53cf30e"}`,
	}

	InitMetrics()

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_request_count_total{hostname="www.example.com"} 7`)
	assert.Contains(t, actual, `section_http_bytes_total{hostname="www.example.com"} 5875`)
}

func testBytesAndBytesSentAreRead(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","status":"200","bytes":"10","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","bytes_sent":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
	}

	InitMetrics()

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_request_count_total{hostname="www.example.com"} 2`)
	assert.Contains(t, actual, `section_http_bytes_total{hostname="www.example.com"} 30`)
}

func testInvalidBytesAndBytesSent(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","status":"200","bytes":"10","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","status":"200","bytes":"-","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","bytes_sent":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","bytes_sent":"-","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
	}

	InitMetrics()

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_request_count_total{hostname="www.example.com"} 4`)
	assert.Contains(t, actual, `section_http_bytes_total{hostname="www.example.com"} 30`)
}

func testJSONParseErrors(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`Not JSON`,
		`{Neither is is}`,
		`{"Broken: "Property"}`,
	}

	InitMetrics()

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_json_parse_errors_total 3`)
}

func testP8sServer(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"foo.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"4e189f278375962cd19d380562846296"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"6ef1b5083893627d2426e42206d78f70"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.077","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"1959","body_bytes_sent":"1498","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.077","upstream_connect_time":"0.057","upstream_response_time":"0.077","upstream_response_length":"1498","upstream_bytes_received":"1889","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"b1ea9bc0be7edfc997bc18a9f6b20d68"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.073","hostname":"foo.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.073","upstream_connect_time":"0.055","upstream_response_time":"0.073","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"ff3117bb0ac0307d8d0e78fc8b8ba5c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"85e833ae62745c50492c80b4d7b78016"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.072","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.072","upstream_connect_time":"0.054","upstream_response_time":"0.072","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"a095e3c2c3a0f25b4bbca4c941babd76"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.071","hostname":"bar.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.071","upstream_connect_time":"0.053","upstream_response_time":"0.071","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"8fb3941b35418bdfa1946ef02c90e8c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"2126","body_bytes_sent":"1665","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"1665","upstream_bytes_received":"2056","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"789addb393a18ff1caf5d776b53cf30e"}`,
	}

	InitMetrics()

	writeLogs(t, logs)

	actual := getP8sHTTPResponse(t)

	assert.Contains(t, actual, `section_http_request_count_total{hostname="bar.example.com"} 1`)
	assert.Contains(t, actual, `section_http_bytes_total{hostname="www.example.com"} 5875`)
}

func testAdditionalLabelsAreUsed(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","status":"200","bytes":"10","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","bytes_sent":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
	}

	InitMetrics("http_accept_encoding")

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_request_count_total{hostname="www.example.com",http_accept_encoding="gzip"} 2`)
	assert.Contains(t, actual, `section_http_bytes_total{hostname="www.example.com",http_accept_encoding="gzip"} 30`)
}

func testAdditionalLabelsWhenMissingFromLogs(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","status":"200","bytes":"10","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","bytes_sent":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
	}

	InitMetrics("missing_field")

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_request_count_total{hostname="www.example.com",missing_field=""} 2`)
	assert.Contains(t, actual, `section_http_bytes_total{hostname="www.example.com",missing_field=""} 30`)
}

func testNonStringProperties(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","status":"200","int": 12345}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","bool": true}`,
	}

	InitMetrics("int", "bool")

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_request_count_total{bool="",hostname="www.example.com",int="12345"} 1`)
	assert.Contains(t, actual, `section_http_request_count_total{bool="true",hostname="www.example.com",int=""} 1`)
}

func testAdditionalMetricsAfterInit(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","status":"200","bytes":"10","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","bytes_sent":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
	}

	counter := prometheus.NewCounter(prometheus.CounterOpts{Name: "name", Namespace: "namespace"})

	registry := InitMetrics()
	registry.MustRegister(counter)

	counter.Add(5)

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `namespace_name 5`)
}

func testPageViews(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","status":"200","content_type": "text/html", "bytes":"10","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","content_type": "text/html", "bytes":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"404","content_type": "text/html", "bytes":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","content_type": "text/css", "bytes":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
	}

	InitMetrics()

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_page_view_total 2`)
}

func testContentTypeBucket(t *testing.T, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","status":"200","content_type": "text/html", "bytes":"10","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","content_type": "text/html", "bytes":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"404","content_type": "text/html", "bytes":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"www.example.com","status":"200","content_type": "-", "bytes":"20","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
	}

	InitMetrics("content_type")

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_request_count_total{content_type_bucket="html",hostname="www.example.com"} 3`)
	assert.Contains(t, actual, `section_http_request_count_total{content_type_bucket="",hostname="www.example.com"} 1`)
}

func TestReaderRunning(t *testing.T) {
	stdout := setupReader(t)

	t.Run("testLogsOutputEqualsInput", func(t *testing.T) { testLogsOutputEqualsInput(t, stdout) })
	t.Run("testCountersIncrease", func(t *testing.T) { testCountersIncrease(t, stdout) })
	t.Run("testBytesAndBytesSentAreRead", func(t *testing.T) { testBytesAndBytesSentAreRead(t, stdout) })
	t.Run("testInvalidBytesAndBytesSent", func(t *testing.T) { testInvalidBytesAndBytesSent(t, stdout) })
	t.Run("testJSONParseErrors", func(t *testing.T) { testJSONParseErrors(t, stdout) })
	t.Run("testP8sServer", func(t *testing.T) { testP8sServer(t, stdout) })
	t.Run("testAdditionalLabelsAreUsed", func(t *testing.T) { testAdditionalLabelsAreUsed(t, stdout) })
	t.Run("testAdditionalLabelsWhenMissingFromLogs", func(t *testing.T) { testAdditionalLabelsWhenMissingFromLogs(t, stdout) })
	t.Run("testNonStringProperties", func(t *testing.T) { testNonStringProperties(t, stdout) })
	t.Run("testAdditionalMetricsAfterInit", func(t *testing.T) { testAdditionalMetricsAfterInit(t, stdout) })
	t.Run("testPageViews", func(t *testing.T) { testPageViews(t, stdout) })
	t.Run("testContentTypeBucket", func(t *testing.T) { testContentTypeBucket(t, stdout) })
}

func TestSetupModule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	var stdout bytes.Buffer
	err := SetupModule(fifoFilePath, &stdout, os.Stderr, "status", "content_type")
	if err != nil {
		t.Error(err)
	}

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","content_type":"text/html","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"foo.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","content_type":"text/html","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","content_type":"text/html","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"4e189f278375962cd19d380562846296"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"0","upstream_bytes_received":"288","content_type":"text/html","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"6ef1b5083893627d2426e42206d78f70"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.077","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"1959","body_bytes_sent":"1498","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.077","upstream_connect_time":"0.057","upstream_response_time":"0.077","upstream_response_length":"1498","upstream_bytes_received":"1889","content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"b1ea9bc0be7edfc997bc18a9f6b20d68"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.073","hostname":"foo.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.073","upstream_connect_time":"0.055","upstream_response_time":"0.073","upstream_response_length":"0","upstream_bytes_received":"288","content_type":"text/html","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"ff3117bb0ac0307d8d0e78fc8b8ba5c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","content_type":"text/html","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"85e833ae62745c50492c80b4d7b78016"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.072","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.072","upstream_connect_time":"0.054","upstream_response_time":"0.072","upstream_response_length":"0","upstream_bytes_received":"288","content_type":"text/html","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"a095e3c2c3a0f25b4bbca4c941babd76"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.071","hostname":"bar.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.071","upstream_connect_time":"0.053","upstream_response_time":"0.071","upstream_response_length":"0","upstream_bytes_received":"288","content_type":"text/html","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"8fb3941b35418bdfa1946ef02c90e8c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"www.example.com","request":"GET /a/path HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"2126","body_bytes_sent":"1665","upstream_label":"default","upstream_addr":"198.51.100.1:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.example.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"1665","upstream_bytes_received":"2056","content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"789addb393a18ff1caf5d776b53cf30e"}`,
	}

	writeLogs(t, logs)

	actual := gatherP8sResponse(t)

	assert.Contains(t, actual, `section_http_request_count_total{content_type_bucket="javascript",hostname="www.example.com",status="200"} 2`)
	assert.Contains(t, actual, `section_http_bytes_total{content_type_bucket="html",hostname="www.example.com",status="304"} 1790`)
}

func TestAddRequestUniqueHostnames(t *testing.T) {
	InitMetrics()

	// reset the map
	uniqueHostnameMap = make(map[string]struct{})

	maxUniqueHostnames = 2 // keep the test brief

	logline := map[string]interface{}{
		"bytes":        7,
		"content_type": "text/plain",
		"status":       "405",
	}
	labels := map[string]string{}

	// first unique hostname
	labels["hostname"] = "a.foo.com"
	addRequest(labels, logline)
	assert.Equal(t, "a.foo.com", labels["hostname"])
	assert.Contains(t, uniqueHostnameMap, "a.foo.com")

	// second unique hostname
	labels["hostname"] = "b.foo.com"
	addRequest(labels, logline)
	assert.Equal(t, "b.foo.com", labels["hostname"])
	assert.Contains(t, uniqueHostnameMap, "b.foo.com")

	// third unique hostname exceeds the maximum
	labels["hostname"] = "c.foo.com"
	addRequest(labels, logline)
	assert.Equal(t, "max-hostnames-reached", labels["hostname"])
	assert.NotContains(t, uniqueHostnameMap, "c.foo.com")

	// first unique hostname still counted
	labels["hostname"] = "a.foo.com"
	addRequest(labels, logline)
	assert.Equal(t, "a.foo.com", labels["hostname"], "first unique hostname, second request")
	assert.Contains(t, uniqueHostnameMap, "a.foo.com", "first unique hostname, second request")
}

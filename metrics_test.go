package metrics

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
)

const (
	fifoFilePath = "/tmp/section.module.metrics-fifotest"
	moduleName   = "test_module"
)

func TestCreateLogFifo(t *testing.T) {
	path := "/tmp/TestCreateLogFifo-file"
	err := CreateLogFifo(path)
	if err != nil {
		t.Error(err)
	}

	fileinfo, err := os.Stat(path)
	if err != nil {
		t.Error(err)
	}

	mode := fileinfo.Mode()
	if mode&os.ModeNamedPipe != os.ModeNamedPipe {
		t.Errorf("Mode of file %s is %s does not have expected %s", path, mode, os.ModeNamedPipe)
	}

	err = os.Remove(path)
	if err != nil {
		t.Errorf("Error removing %s: %#v", path, err)
	}
}

func TestCreateLogFifoFails(t *testing.T) {
	nonExistantPath := "/i/dont/exist/test-file"

	err := CreateLogFifo(nonExistantPath)

	if err == nil {
		t.Errorf("CreateLogFifo didn't fail when creating %s", nonExistantPath)
	}
}

func setup(t *testing.T) (io.Writer, *bytes.Buffer) {
	err := CreateLogFifo(fifoFilePath)
	if err != nil {
		t.Error(err)
	}

	reader, err := OpenReadFifo(fifoFilePath)
	if err != nil {
		t.Errorf("OpenReadFifo(%s) failed: %#v", fifoFilePath, err)
	}

	writer, err := OpenWriteFifo(fifoFilePath)
	if err != nil {
		t.Errorf("OpenWriteFifo(%s) failed: %#v", fifoFilePath, err)
	}

	var stdout bytes.Buffer
	StartReader(moduleName, reader, &stdout, os.Stderr)

	return writer, &stdout
}

func teardown(t *testing.T) {
	err := os.Remove(fifoFilePath)
	if err != nil {
		t.Errorf("Error removing %s: %#v", fifoFilePath, err)
	}
}

func testLogsOutputEqualsInput(t *testing.T, writer io.Writer, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","request":"GET /c/hotjar-327527.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","request":"GET /c/hotjar-66107.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","request":"GET /c/hotjar-237964.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.50.128.205:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"4e189f278375962cd19d380562846296"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","request":"GET /c/hotjar-9443.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.50.128.205:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"6ef1b5083893627d2426e42206d78f70"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.077","request":"GET /c/hotjar-52354.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"1959","body_bytes_sent":"1498","upstream_label":"default","upstream_addr":"52.49.107.188:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.077","upstream_connect_time":"0.057","upstream_response_time":"0.077","upstream_response_length":"1498","upstream_bytes_received":"1889","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"b1ea9bc0be7edfc997bc18a9f6b20d68"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.073","request":"GET /c/hotjar-637045.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"54.194.227.5:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.073","upstream_connect_time":"0.055","upstream_response_time":"0.073","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"ff3117bb0ac0307d8d0e78fc8b8ba5c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","request":"GET /c/hotjar-770871.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.30.74.76:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"85e833ae62745c50492c80b4d7b78016"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.072","request":"GET /c/hotjar-1145835.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.072","upstream_connect_time":"0.054","upstream_response_time":"0.072","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"a095e3c2c3a0f25b4bbca4c941babd76"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.071","request":"GET /c/hotjar-1286870.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"54.194.227.5:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.071","upstream_connect_time":"0.053","upstream_response_time":"0.071","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"8fb3941b35418bdfa1946ef02c90e8c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","request":"GET /c/hotjar-1281702.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"2126","body_bytes_sent":"1665","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"1665","upstream_bytes_received":"2056","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"789addb393a18ff1caf5d776b53cf30e"}`,
	}

	initMetrics(moduleName)

	for _, line := range logs {
		_, err := writer.Write([]byte(line + "\n"))
		if err != nil {
			t.Errorf("Error writing line: %#v", err)
		}
	}

	//Give the reader loop time to finish
	time.Sleep(time.Second * 1)

	outputLines := strings.Split(stdout.String(), "\n")

	for i := 0; i < len(logs); i++ {
		if logs[i] != outputLines[i] {
			t.Errorf("Logs not equal, %s != %s", outputLines[i], logs[i])
		}
	}
}

func testCountersIncrease(t *testing.T, writer io.Writer, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"vars.hotjar.com","request":"GET /c/hotjar-327527.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"foo.hotjar.com","request":"GET /c/hotjar-66107.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"vars.hotjar.com","request":"GET /c/hotjar-237964.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.50.128.205:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"4e189f278375962cd19d380562846296"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"vars.hotjar.com","request":"GET /c/hotjar-9443.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.50.128.205:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"6ef1b5083893627d2426e42206d78f70"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.077","hostname":"vars.hotjar.com","request":"GET /c/hotjar-52354.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"1959","body_bytes_sent":"1498","upstream_label":"default","upstream_addr":"52.49.107.188:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.077","upstream_connect_time":"0.057","upstream_response_time":"0.077","upstream_response_length":"1498","upstream_bytes_received":"1889","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"b1ea9bc0be7edfc997bc18a9f6b20d68"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.073","hostname":"foo.hotjar.com","request":"GET /c/hotjar-637045.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"54.194.227.5:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.073","upstream_connect_time":"0.055","upstream_response_time":"0.073","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"ff3117bb0ac0307d8d0e78fc8b8ba5c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"vars.hotjar.com","request":"GET /c/hotjar-770871.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.30.74.76:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"85e833ae62745c50492c80b4d7b78016"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.072","hostname":"vars.hotjar.com","request":"GET /c/hotjar-1145835.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.072","upstream_connect_time":"0.054","upstream_response_time":"0.072","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"a095e3c2c3a0f25b4bbca4c941babd76"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.071","hostname":"bar.hotjar.com","request":"GET /c/hotjar-1286870.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"54.194.227.5:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.071","upstream_connect_time":"0.053","upstream_response_time":"0.071","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"8fb3941b35418bdfa1946ef02c90e8c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"vars.hotjar.com","request":"GET /c/hotjar-1281702.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"2126","body_bytes_sent":"1665","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"1665","upstream_bytes_received":"2056","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"789addb393a18ff1caf5d776b53cf30e"}`,
	}

	initMetrics(moduleName)

	for _, line := range logs {
		_, err := writer.Write([]byte(line + "\n"))
		if err != nil {
			t.Errorf("Error writing line: %#v", err)
		}
	}

	//Give the reader loop time to finish
	time.Sleep(time.Second * 1)

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

	actual := out.String()
	expected := `test_module_http_request_count_total{hostname="bar.hotjar.com",status="304"} 1`
	if !strings.Contains(actual, expected) {
		t.Errorf("Output:\n%s\n does not contain expected %s", actual, expected)
	}

	expected = `test_module_http_bytes_total{hostname="vars.hotjar.com",status="304"} 1790`
	if !strings.Contains(actual, expected) {
		t.Errorf("Output:\n%s\n does not contain expected %s", actual, expected)
	}
}

func testP8sServer(t *testing.T, writer io.Writer, stdout *bytes.Buffer) {

	logs := []string{
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"vars.hotjar.com","request":"GET /c/hotjar-327527.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"cf99df8057b93ec96c0ee1253ba4c309"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.069","hostname":"foo.hotjar.com","request":"GET /c/hotjar-66107.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.069","upstream_connect_time":"0.052","upstream_response_time":"0.069","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"451e230222237f722eb49324d47142f6"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"vars.hotjar.com","request":"GET /c/hotjar-237964.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.50.128.205:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"4e189f278375962cd19d380562846296"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"vars.hotjar.com","request":"GET /c/hotjar-9443.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.50.128.205:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"6ef1b5083893627d2426e42206d78f70"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.077","hostname":"vars.hotjar.com","request":"GET /c/hotjar-52354.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"1959","body_bytes_sent":"1498","upstream_label":"default","upstream_addr":"52.49.107.188:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.077","upstream_connect_time":"0.057","upstream_response_time":"0.077","upstream_response_length":"1498","upstream_bytes_received":"1889","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"b1ea9bc0be7edfc997bc18a9f6b20d68"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.073","hostname":"foo.hotjar.com","request":"GET /c/hotjar-637045.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"54.194.227.5:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.073","upstream_connect_time":"0.055","upstream_response_time":"0.073","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"ff3117bb0ac0307d8d0e78fc8b8ba5c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.070","hostname":"vars.hotjar.com","request":"GET /c/hotjar-770871.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.30.74.76:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.070","upstream_connect_time":"0.052","upstream_response_time":"0.070","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"85e833ae62745c50492c80b4d7b78016"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.072","hostname":"vars.hotjar.com","request":"GET /c/hotjar-1145835.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.072","upstream_connect_time":"0.054","upstream_response_time":"0.072","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"a095e3c2c3a0f25b4bbca4c941babd76"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.071","hostname":"bar.hotjar.com","request":"GET /c/hotjar-1286870.js?sv=5 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"304","bytes_sent":"358","body_bytes_sent":"0","upstream_label":"default","upstream_addr":"54.194.227.5:443","upstream_status":"304","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.071","upstream_connect_time":"0.053","upstream_response_time":"0.071","upstream_response_length":"0","upstream_bytes_received":"288","upstream_http_content_type":"-","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"-","upstream_http_transfer_encoding":"-","sent_http_content_length":"-","sent_http_content_encoding":"-","sent_http_transfer_encoding":"-","section-io-id":"8fb3941b35418bdfa1946ef02c90e8c7"}`,
		`{"time":"2019-06-20T01:34:36+00:00","request_time":"0.075","hostname":"vars.hotjar.com","request":"GET /c/hotjar-1281702.js?sv=6 HTTP/1.1","http_accept_encoding":"gzip","http_x_forwarded_proto":"https","http_upgrade":"-","http_connection":"-","status":"200","bytes_sent":"2126","body_bytes_sent":"1665","upstream_label":"default","upstream_addr":"52.48.54.139:443","upstream_status":"200","upstream_request_connection":"","upstream_request_host":"in.hotjar.com","upstream_header_time":"0.075","upstream_connect_time":"0.056","upstream_response_time":"0.075","upstream_response_length":"1665","upstream_bytes_received":"2056","upstream_http_content_type":"application/javascript","upstream_http_cache_control":"max-age=60","upstream_http_content_length":"-","upstream_http_content_encoding":"gzip","upstream_http_transfer_encoding":"chunked","sent_http_content_length":"-","sent_http_content_encoding":"gzip","sent_http_transfer_encoding":"chunked","section-io-id":"789addb393a18ff1caf5d776b53cf30e"}`,
	}

	initMetrics(moduleName)

	StartPrometheusServer(os.Stderr)

	for _, line := range logs {
		_, err := writer.Write([]byte(line + "\n"))
		if err != nil {
			t.Errorf("Error writing line: %#v", err)
		}
	}

	//Give the reader loop time to finish
	time.Sleep(time.Second * 1)

	resp, err := http.Get(MetricsURI)
	if err != nil {
		t.Errorf("Error getting %s: %#v", MetricsURI, err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading body: %#v", err)
	}

	body := string(bodyBytes)
	expected := `test_module_http_request_count_total{hostname="bar.hotjar.com",status="304"} 1`
	if !strings.Contains(body, expected) {
		t.Errorf("HTTP response:\n%s\n does not contain expected %s", body, expected)
	}

	expected = `test_module_http_bytes_total{hostname="vars.hotjar.com",status="304"} 1790`
	if !strings.Contains(body, expected) {
		t.Errorf("HTTP response:\n%s\n does not contain expected %s", body, expected)
	}
}

func TestReaderRunning(t *testing.T) {
	writer, stdout := setup(t)

	t.Run("testLogsOutputEqualsInput", func(t *testing.T) { testLogsOutputEqualsInput(t, writer, stdout) })
	t.Run("testCountersIncrease", func(t *testing.T) { testCountersIncrease(t, writer, stdout) })
	t.Run("testP8sServer", func(t *testing.T) { testP8sServer(t, writer, stdout) })

	teardown(t)
}

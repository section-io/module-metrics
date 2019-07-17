# module-metrics

Access logs for Section modules are written to the container `STDOUT` for collection by the Filebeat daemonset.

This module will intercept logs by reading from a fifo file, parse them to collect metrics and then write to a `io.Writer` to continue the existing log processing. The `io.Writer` will normally be `os.Stdout`.

## Expected Log Format

This expectes a JSON log format with one line per HTTP request logged with the following properties:

* `hostname` - The value of the `Host` HTTP request header.
* `status` - The value of the response status code.
* `bytes` or `bytes_sent` - The total bytes (header + body) sent downstream by this module .

`metrics_test.go` has examples of valid log lines.

## Metrics Collection

The metrics collected are:

* `section_http_request_count_total{ section_io_module_name="module name", hostname="www.example.com", status="200" }` - Counter of number of HTTP requests by hostname & status.
* `section_http_bytes_total{ section_io_module_name="module name", hostname="www.example.com", status="200" }` - Counter of sum of bytes sent downstream by hostname & status.
* `section_http_json_parse_errors_total{ section_io_module_name="module name" }` - Counter of the number of times it has been unable to JSON parse a log line.

The `section_io_module_name` is configured as a target label on the service monitor for the module using this module.

The metrics are published as a Prometheus exporter by default on port `9000` with the path `/metrics`.

## Additional Labels

Metrics can have additional labels added based on fields in the log lines. These can be added as a string array when setting up the module (see code below). The valus will pass through the `sanitizeValue` function in `metrics.go` so that fields that 
have known sanitization issues (like stripping the additional fields from `content_type`.) Additional sanitization can be added to this function as needed. The additional lables need to be valid Prometheus labels (see https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels).  This is required, but not currently enforced by the code. The additional lables also need to match the name of a field in the log lines. Lines that do not have the field will create a metric with a blank label value.

## Usage

```
import metrics "github.com/section-io/module-metrics"

...

err := metrics.SetupModule(pathToLogFile, os.Stdout, os.Stderr, []string{"content_type"})
```
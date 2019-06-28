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

* `http_request_count_total{ section_io_module_name="module name", hostname="www.example.com", status="200" }` - Counter of number of HTTP requests by hostname & status.
* `http_bytes_total{ section_io_module_name="module name", hostname="www.example.com", status="200" }` - Counter of sum of bytes sent downstream by hostname & status.
* `http_json_parse_errors_total{ section_io_module_name="module name" }` - Counter of the number of times it has been unable to JSON parse a log line.

The `section_io_module_name` is configured as a target label on the service monitor for the module using this module.

The metrics are published as a Prometheus exporter by default on port `9000` with the path `/metrics`.

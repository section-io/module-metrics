# module-metrics

Access logs for Section modules are written to the container `STDOUT`
for collection by the Filebeat daemonset.

This module will intercept logs by reading from a fifo file provided
during a `SetupYYYY` method.  It then parses log lines from which it
makes prometheues metrics.  The logs unmodified are also written to an
`io.Writer` which is generally stdout.

## Expected Log Format

This expects a JSON log format with one line per HTTP request logged
with the following properties:

* `hostname` - The value of the `Host` HTTP request header.
* `status` - The value of the response status code.
* `bytes` or `bytes_sent` - The total bytes (header + body) sent downstream by this module .

`metrics_test.go` has examples of valid log lines.

## Metrics Collection

The metrics collected are:

* `section_http_request_count_total{ section_io_module_name="module name", status="200" }` - Counter of number of HTTP requests by status.
* `section_http_bytes_total{ section_io_module_name="module name", status="200" }` - Counter of sum of bytes sent downstream by status.
* `section_http_json_parse_errors_total{ section_io_module_name="module name" }` - Counter of the number of times it has been unable to JSON parse a log line.
* `section_http_request_count_by_hostname_total{ hostname="www.example.com" }` - Counter of the number of HTTP requests by hostname.
* `section_http_bytes_by_hostname_total{ hostname="www.example.com" }` - Counter of sum of bytes sent downstream by hostname.

The `by_hostname` metrics will only be generated if `hostname` is included in the additional labels parameter.

The `section_io_module_name` is configured as a target label on the
service monitor for the module using this module.

The metrics are published as a Prometheus exporter by default on port
`9000` with the path `/metrics`.

## Additional Labels

Metrics can have additional labels added based on fields in the log
lines. These can be added as a string array when setting up the module
(see code below). The values will pass through the `sanitizeValue`
function in `metrics.go` so that fields that have known sanitization
issues (like stripping the additional fields from `content_type`.)
Additional sanitization can be added to this function as needed. The
additional labels need to be valid Prometheus labels (see
https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels).
This is required, but not currently enforced by the code. The
additional labels also need to match the name of a field in the log
lines. Lines that do not have the field will create a metric with a
blank label value.

## Usage

There are two ways to use the module.

### Using a FIFO file for the logs to be written to

This will setup the FIFO file at `pathToLogFile`, start the metrics
collection and redirect all content written the log file to
`os.Stdout`.  Any errors will be written to `os.Stderr` and the
metrics will have the additional label of `content_type`.

    ```
    import metrics "github.com/section-io/module-metrics"

    ...

    err := metrics.SetupModule(pathToLogFile, os.Stdout, os.Stderr, "content_type")
    ```

### Using a reader

If the logs are already in an `io.Reader` you can setup the metrics
and start the reader explicitly.

    ```
    metrics.InitMetrics("content_type")
    metrics.StartReader(logReader, os.Stdout, os.Stderr)
    ```

### Turning GeoIP latitude/longitude to GeoIP hashes

The setup for GeoIP hashes uses the method `SetupWithGeoHash` which is
very similar to `SetupModule`.  The difference is that the Geo Hash
precision must be provided.  The precision represent an output of 1-12
Geo Hash characters.  Since these hash characters will result in
labels on request metrics it's important to choose a number that won't
adversely effect your prometheus scraping.  A reasonable starting
point is 2 characters of precision and then slowly grow from there
keeping in mind an exponential growth.


## Tagging and Releasing

Once we merge changes from feature branch to master after code review & approval,
we need to manually tag it

    # checkout master and fetch latest merged changes
    git checkout master
    git pull
    # create and push tag ( version in semver format )
    git tag -a vX.Y.Z -m'Short message describing change'
    git push origin --tags

To create a new release for this module, click on the [new release link](https://github.com/section-io/module-metrics/releases/new) and provide relevant release details and publish.

package metrics //import github.com/section-io/module-metrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

const maxLabelValueLength = 80

var (
	filepath          string
	isValidHostHeader = regexp.MustCompile(`^[a-z0-9.-]+$`).MatchString
)

func sanitizeValue(label string, value interface{}) string {

	if value == nil || value == "" || value == "-" {
		return ""
	}

	// Convert to a string, no matter what underlying type it is
	labelValue := fmt.Sprintf("%v", value)
	labelValue = strings.TrimSpace(labelValue)

	switch label {
	case "content_type":
		labelValue = strings.ToLower(labelValue)
		if strings.HasPrefix(labelValue, "image") {
			labelValue = "image"
		} else if strings.HasPrefix(labelValue, "text/html") {
			labelValue = "html"
		} else if strings.HasPrefix(labelValue, "text/css") {
			labelValue = "css"
		} else if strings.Contains(labelValue, "javascript") {
			labelValue = "javascript"
		} else {
			labelValue = "other"
		}
	case "hostname":
		labelValue = strings.Split(labelValue, ":")[0]
		labelValue = strings.ToLower(labelValue)
		if !isValidHostHeader(labelValue) {
			labelValue = ""
		}

	case "status":
		statusInt, _ := strconv.Atoi(labelValue)
		switch {
		case statusInt >= 100 && statusInt <= 103:
		case statusInt >= 200 && statusInt <= 208:
		case statusInt >= 300 && statusInt <= 308:
		case statusInt >= 400 && statusInt <= 431:
		case statusInt == 499:
		case statusInt >= 500 && statusInt <= 411:
		default:
			// If it matches any of the above cases, do nothing (leave labelValue as is)
			// otherwise set to blank
			labelValue = ""
		}
	}

	if len(labelValue) > maxLabelValueLength {
		labelValue = labelValue[0:maxLabelValueLength]
	}

	return labelValue
}

func getBytes(l map[string]interface{}) int {

	var bytes interface{}
	var ok bool

	if bytes, ok = l["bytes"]; !ok {
		bytes, _ = l["bytes_sent"]
	}

	// Force convert to a string then to int, simpler than trying to figure out what the
	// underlying type is. Atoi will return a 0 if the string can't be converted to an int.
	bytes, _ = strconv.Atoi(fmt.Sprintf("%v", bytes))

	return bytes.(int)
}

// CreateLogFifo creates the log pipe, will remove the file first if it already exists.
func CreateLogFifo(path string) error {

	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "Remove %s failed: %v", path, err)
	}

	err = syscall.Mkfifo(path, 0666)
	if err != nil {
		return errors.Wrapf(err, "Mkfifo %s failed: %v", path, err)
	}

	// Really make sure the file is 0666 in case of umask
	err = os.Chmod(path, 0666)
	if err != nil {
		return errors.Wrapf(err, "Chmod %s failed: %v", path, err)
	}

	filepath = path

	return nil
}

// OpenReadFifo opens the fifo file for reading, returning the reader
func OpenReadFifo(path string) (io.ReadCloser, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, os.ModeNamedPipe)
	if err != nil {
		return nil, errors.Wrapf(err, "OpenReadFifo %s failed: %v", path, err)
	}

	return file, nil
}

// StartReader starts a loop in a goroutine that reads from the fifo file and writes out to the
// output file. Any errors regarding parsing the log line are written to the errorWriter (eg os.Stderr)
// but do not panic.
func StartReader(file io.Reader, output io.Writer, errorWriter io.Writer) {

	go func() {

		reader := bufio.NewReader(file)
		line, err := reader.ReadBytes('\n')
		for err == nil {

			_, writeErr := output.Write(line)
			if writeErr != nil {
				panic(errors.Wrapf(writeErr, "Writing to output failed"))
			}

			var logline map[string]interface{}
			jsonErr := json.Unmarshal(line, &logline)
			if jsonErr != nil {
				_, _ = fmt.Fprintf(errorWriter, "json.Unmarshal failed: %v", jsonErr)
				jsonParseErrorTotal.Inc()
			} else {
				labelValues := map[string]string{}

				for _, label := range p8sLabels {
					labelValues[label] = sanitizeValue(label, logline[label])
				}
				addRequest(labelValues, getBytes(logline))
			}

			line, err = reader.ReadBytes('\n')
		}

		// If EOF is reached the writer program closed the file, so reopen it
		if err == io.EOF {
			file, err = OpenReadFifo(filepath)
			if err != nil {
				panic(err)
			}
			StartReader(file, output, errorWriter)
			return
		}

		panic(errors.Wrapf(err, "ReadBytes failed"))
	}()
}

// SetupModule does the default setup scenario: creating & opening the FIFO file,
// starting the Prometheus server and starting the reader.
func SetupModule(path string, stdout io.Writer, stderr io.Writer, additionalLabels ...string) error {
	err := CreateLogFifo(path)
	if err != nil {
		return err
	}

	reader, err := OpenReadFifo(path)
	if err != nil {
		return err
	}

	InitMetrics(additionalLabels...)

	StartReader(reader, stdout, stderr)

	return nil
}

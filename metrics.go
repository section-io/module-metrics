package metrics //import github.com/section-io/module-metrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
)

var (
	filepath string
)

func getBytes(l map[string]string) int {
	bytes, _ := strconv.Atoi(l["bytes"])
	if bytes <= 0 {
		bytes, _ = strconv.Atoi(l["bytes_sent"])
	}

	return bytes
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

			logline := map[string]string{}
			jsonErr := json.Unmarshal(line, &logline)
			if jsonErr != nil {
				_, _ = fmt.Fprintf(errorWriter, "json.Unmarshal failed: %v", jsonErr)
				jsonParseErrorTotal.Inc()
			} else {
				labelValues := map[string]string{}

				for _, label := range p8sLabels {
					labelValues[label] = logline[label]
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
func SetupModule(path string, stdout io.Writer, stderr io.Writer) error {
	err := CreateLogFifo(path)
	if err != nil {
		return err
	}

	reader, err := OpenReadFifo(path)
	if err != nil {
		return err
	}

	InitMetrics([]string{})

	StartReader(reader, stdout, stderr)

	StartPrometheusServer(stderr)

	return nil
}

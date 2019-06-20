package metrics //import section.io/module-metrics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/pkg/errors"
)

type logLine struct {
	Status string `json:"status"`
	Bytes  int    `json:"bytes,string"`
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

	return nil
}

// OpenReadFifo opens the fifo file for reading, returning the reader
func OpenReadFifo(path string) (io.Reader, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, os.ModeNamedPipe)
	if err != nil {
		return nil, errors.Wrapf(err, "OpenReadFifo %s failed: %v", path, err)
	}

	return file, nil
}

// OpenWriteFifo opens the fifo file for writing, returning the writer
func OpenWriteFifo(path string) (io.Writer, error) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, errors.Wrapf(err, "OpenWriteFifo %s failed: %v", path, err)
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

			var logline logLine
			jsonErr := json.Unmarshal(line, &logline)
			if jsonErr != nil {
				_, _ = fmt.Fprintf(errorWriter, "json.Unmarshal failed: %v", jsonErr)
			}

			_, _ = fmt.Fprintf(errorWriter, "Bytes: %d, Status: %s", logline.Bytes, logline.Status)

			line, err = reader.ReadBytes('\n')
		}

		if err != nil {
			panic(errors.Wrapf(err, "ReadBytes failed"))
		}
	}()
}

package metrics //import section.io/module-metrics

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

const (
	DefaultFifoFilePath = "/run/section-logs.pipe"
)

type LogLine struct {
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

//OpenReadFifo opens the fifo file for reading, returning the reader
func OpenReadFifo(path string) (io.Reader, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, errors.Wrapf(err, "OpenReadFifo %s failed: %v", path, err)
	}

	return file, nil
}

//OpenWriteFifo opens the fifo file for writing, returning the writer
func OpenWriteFifo(path string) (io.Reader, error) {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, errors.Wrapf(err, "OpenWriteFifo %s failed: %v", path, err)
	}

	return file, nil
}

func StartReader(file io.Reader, output io.Writer) {
	go func() {

		reader := bufio.NewReader(file)
		line, err := reader.ReadBytes('\n')
		for err == nil {

			_, writeErr := output.Write(line)
			if writeErr != nil {
				log.Printf("Writing to output failed: %v", writeErr)
			}

			var logline LogLine
			jsonErr := json.Unmarshal(line, &logline)
			if jsonErr != nil {
				log.Printf("json.Unmarshal failed: %v", jsonErr)
			}

			log.Printf("Bytes: %d, Status: %s", logline.Bytes, logline.Status)

			time.Sleep(time.Second * 1)

			line, err = reader.ReadBytes('\n')
		}

		if err != nil {
			log.Fatalf("ReadBytes failed: %v", err)
		}
	}()
}

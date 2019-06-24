package metrics

import (
	"bytes"
	"os"
	"testing"
	"time"
)

const (
	fifoFilePath = "/tmp/section.module.metrics-fifotest"
)

func setupReader(t *testing.T) *bytes.Buffer {
	err := CreateLogFifo(fifoFilePath)
	if err != nil {
		t.Error(err)
	}

	reader, err := OpenReadFifo(fifoFilePath)
	if err != nil {
		t.Errorf("OpenReadFifo(%s) failed: %#v", fifoFilePath, err)
	}

	var stdout bytes.Buffer
	StartReader(reader, &stdout, os.Stderr)

	return &stdout
}

func writeLogs(t *testing.T, logs []string) {
	writer, err := OpenWriteFifo(fifoFilePath)
	defer func() { _ = writer.Close() }()
	if err != nil {
		t.Errorf("OpenWriteFifo(%s) failed: %#v", fifoFilePath, err)
	}

	for _, line := range logs {
		_, err := writer.Write([]byte(line + "\n"))
		if err != nil {
			t.Errorf("Error writing line: %#v", err)
		}
	}

	//Give the reader loop time to finish
	time.Sleep(time.Second * 1)
}

package metrics

import (
	"os"
	"testing"
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

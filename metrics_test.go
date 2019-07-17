package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestSanitizeContentType(t *testing.T) {
	const expected = "text/html"
	actual := sanitizeValue("content_type", "text/html; charset=iso-8859-1")

	assert.Equal(t, expected, actual)
}

func TestSanitizeContentTypeNoSemiColon(t *testing.T) {
	const expected = "text/html"
	actual := sanitizeValue("content_type", "text/html")

	assert.Equal(t, expected, actual)
}

func TestUnsanitizedLabel(t *testing.T) {
	const expected = " fooooo3iwac "
	actual := sanitizeValue("some_unknown_type", expected)

	assert.Equal(t, expected, actual)
}

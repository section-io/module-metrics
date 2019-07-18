package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateLogFifo(t *testing.T) {
	path := "/tmp/TestCreateLogFifo-file"
	err := CreateLogFifo(path)
	assert.NoError(t, err)

	fileinfo, err := os.Stat(path)
	assert.NoError(t, err)

	mode := fileinfo.Mode()
	assert.Equal(t, os.ModeNamedPipe, mode&os.ModeNamedPipe)

	err = os.Remove(path)
	assert.NoError(t, err)
}

func TestCreateLogFifoFails(t *testing.T) {
	nonExistantPath := "/i/dont/exist/test-file"

	err := CreateLogFifo(nonExistantPath)

	assert.Error(t, err)
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

func TestSanitizeContentTypeEmpty(t *testing.T) {
	const expected = ""
	actual := sanitizeValue("content_type", "")

	assert.Equal(t, expected, actual)
}

func TestUnsanitizedLabel(t *testing.T) {
	const expected = "fooooo3iwac"
	actual := sanitizeValue("some_unknown_type", " fooooo3iwac ")

	assert.Equal(t, expected, actual)
}

func TestSanitizeNil(t *testing.T) {
	const expected = ""
	actual := sanitizeValue("foo", nil)

	assert.Equal(t, expected, actual)
}

func TestSanitizeHostname(t *testing.T) {
	const expected = "www.foo.com"
	actual := sanitizeValue("hostname", "www.foo.com")

	assert.Equal(t, expected, actual)
}

func TestSanitizeHostnameWithPort(t *testing.T) {
	const expected = "www.foo.com"
	actual := sanitizeValue("hostname", "www.foo.com:80")

	assert.Equal(t, expected, actual)
}

func TestSanitizeHostnameWithSpaces(t *testing.T) {
	const expected = "www.foo.com"
	actual := sanitizeValue("hostname", "   www.foo.com   ")

	assert.Equal(t, expected, actual)
}

func TestSanitizeHostnameMissing(t *testing.T) {
	const expected = ""
	actual := sanitizeValue("hostname", nil)

	assert.Equal(t, expected, actual)
}

func TestGetBytes(t *testing.T) {
	const expected = 5
	actual := getBytes(map[string]interface{}{"bytes": "5"})

	assert.Equal(t, expected, actual)
}

func TestGetBytesSent(t *testing.T) {
	const expected = 5
	actual := getBytes(map[string]interface{}{"bytes_sent": "5"})

	assert.Equal(t, expected, actual)
}

func TestGetBytesInt(t *testing.T) {
	const expected = 5
	actual := getBytes(map[string]interface{}{"bytes": 5})

	assert.Equal(t, expected, actual)
}

func TestGetBytesSentInt(t *testing.T) {
	const expected = 5
	actual := getBytes(map[string]interface{}{"bytes_sent": 5})

	assert.Equal(t, expected, actual)
}

func TestGetBytesDash(t *testing.T) {
	const expected = 0
	actual := getBytes(map[string]interface{}{"bytes": "-"})

	assert.Equal(t, expected, actual)
}

func TestGetBytesMissing(t *testing.T) {
	const expected = 0
	actual := getBytes(map[string]interface{}{"somthing": "foo"})

	assert.Equal(t, expected, actual)
}

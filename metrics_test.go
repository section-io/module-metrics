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

func TestUnknownLabelSanitize(t *testing.T) {
	const expectedLabel = "some_other_field"
	const expectedValue = "foobar"
	actual, _ := sanitizeLabel("some_other_field", "foobar")
	assert.Equal(t, expectedLabel, actual)
}

func TestSanitizeContentTypeLabelName(t *testing.T) {
	const expectedLabel = "content_type_bucket"
	actual, _ := sanitizeLabel("content_type", "text/html")
	assert.Equal(t, expectedLabel, actual)
}

func TestSanitizeContentTypeHTML(t *testing.T) {
	const expected = "html"
	_, actual := sanitizeLabel("content_type", "text/html; charset=iso-8859-1")
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("content_type", "text/html")
	assert.Equal(t, expected, actual)
}

func TestSanitizeContentTypeImage(t *testing.T) {
	const expected = "image"
	_, actual := sanitizeLabel("content_type", "image/jpg")
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("content_type", "image/gif")
	assert.Equal(t, expected, actual)
}

func TestSanitizeContentTypeCSS(t *testing.T) {
	const expected = "css"
	_, actual := sanitizeLabel("content_type", "text/css")
	assert.Equal(t, expected, actual)
}

func TestSanitizeContentTypeJavascript(t *testing.T) {
	const expected = "javascript"
	_, actual := sanitizeLabel("content_type", "text/javascript")
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("content_type", "application/javascript")
	assert.Equal(t, expected, actual)
}

func TestSanitizeContentTypeEmpty(t *testing.T) {
	const expected = ""
	_, actual := sanitizeLabel("content_type", "")
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("content_type", "-")
	assert.Equal(t, expected, actual)
}

func TestSanitizeContentTypeOther(t *testing.T) {
	const expected = "other"
	_, actual := sanitizeLabel("content_type", "foobar")
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("content_type", "text/rtf")
	assert.Equal(t, expected, actual)
}

func TestUnsanitizedLabel(t *testing.T) {
	const expected = "fooooo3iwac"
	_, actual := sanitizeLabel("some_unknown_type", " fooooo3iwac ")

	assert.Equal(t, expected, actual)
}

func TestSanitizeNil(t *testing.T) {
	const expected = ""
	_, actual := sanitizeLabel("foo", nil)

	assert.Equal(t, expected, actual)
}

func TestSanitizeHostname(t *testing.T) {
	const expected = "www.foo.com"
	_, actual := sanitizeLabel("hostname", "www.foo.com")

	assert.Equal(t, expected, actual)
}

func TestSanitizeHostnameWithPort(t *testing.T) {
	const expected = "www.foo.com"
	_, actual := sanitizeLabel("hostname", "www.foo.com:80")

	assert.Equal(t, expected, actual)
}

func TestSanitizeHostnameWithSpaces(t *testing.T) {
	const expected = "www.foo.com"
	_, actual := sanitizeLabel("hostname", "   www.foo.com   ")

	assert.Equal(t, expected, actual)
}

func TestSanitizeHostnameMissing(t *testing.T) {
	const expected = ""

	_, actual := sanitizeLabel("hostname", nil)
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("hostname", "-")
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("hostname", "    ")
	assert.Equal(t, expected, actual)
}

func TestSanitizeHostnameCasing(t *testing.T) {
	const expected = "www.foo.com"
	_, actual := sanitizeLabel("hostname", "WWw.FOo.COm")

	assert.Equal(t, expected, actual)
}

func TestSanitizeHostnameInvalidChars(t *testing.T) {
	const expected = ""

	_, actual := sanitizeLabel("hostname", "www.fi$h.com")
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("hostname", "%(+ ")
	assert.Equal(t, expected, actual)
}

func TestSanitizeStatus(t *testing.T) {
	const expected = "200"
	_, actual := sanitizeLabel("status", "200")

	assert.Equal(t, expected, actual)
}

func TestSanitizeStatusInvalid(t *testing.T) {
	const expected = ""

	_, actual := sanitizeLabel("status", "220")
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("status", "foobar")
	assert.Equal(t, expected, actual)

	_, actual = sanitizeLabel("status", "220foo")
	assert.Equal(t, expected, actual)

}

func TestSanitizeMaxLength(t *testing.T) {
	const expected = "01234567890123456789012345678901234567890123456789012345678901234567890123456789"
	_, actual := sanitizeLabel("hostname", "012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")

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

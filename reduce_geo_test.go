package metrics

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errorExtractingExpected = errors.New("error while extracting lat/lon from log line")
var errorConvertingExpected = errors.New("error while converting found lat/lon from log line")

func TestExtractGeoIp(t *testing.T) {
	cases := []struct {
		logline  map[string]interface{}
		isZero   bool
		expected coords
		message  string
	}{
		{
			message: "empty logline",
			logline: map[string]interface{}{},
			isZero:  true,
			expected: coords{
				missingLatLon: true,
				missingGeo:    true,
			},
		},
		{
			message: "no geo JSON object",
			logline: map[string]interface{}{
				"geo": map[string]interface{}{},
			},
			expected: coords{
				missingLatLon: true,
				missingGeo:    false,
			},
		},
		{
			message: "only lat provided",
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": "-33.86010",
				},
			},
			expected: coords{
				extractError: errorExtractingExpected,
			},
		},
		{
			message: "only lon provided",
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": ",-33.86010",
				},
			},
			expected: coords{
				extractError: errorExtractingExpected,
			},
		},
		{
			message: "non empty, but still missing lat/lon",
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": " , ",
				},
			},
			expected: coords{
				extractError: errorExtractingExpected,
			},
		},
		{
			message: "non-floats provided",
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": "-,-",
				},
			},
			expected: coords{
				convertError: errorConvertingExpected,
			},
		},
		{
			message: "strange/non-floats provided",
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": "-k,+a",
				},
			},
			expected: coords{
				convertError: errorConvertingExpected,
			},
		},
		{
			message: "neither lat or lon, just a comma",
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": ",",
				},
			},
			expected: coords{
				extractError: errorExtractingExpected,
			},
		},
		{
			message: "bad object found for 'geo' key",
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": "-33.86010,151.21010",
				},
			},
			expected: coords{
				rawLat: "-33.86010",
				rawLon: "151.21010",
			},
		},
		{
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": "1.1,2.2",
				},
			},
			expected: coords{
				rawLat: "1.1",
				rawLon: "2.2",
			},
		},
	}
	for _, c := range cases {
		latlon := extractGeoip(c.logline)
		if c.isZero {
			assert.True(t, latlon.isZero(), "coord: %+v", latlon)
			continue
		}
		if c.expected.extractError != nil {
			assert.NotNil(t, latlon.extractError,
				"message: %s, actual: %+v", c.message, latlon)
			assert.False(t, latlon.isValid())
			continue
		}
		if c.expected.convertError != nil {
			assert.NotNil(t, latlon.convertError,
				"message: %s, actual: %+v", c.message, latlon)
			assert.False(t, latlon.isValid())
			continue
		}
		if c.expected.missingGeo {
			assert.True(t,
				latlon.missingGeo && latlon.missingLatLon,
				"message: %s, actual: %+v", c.message, latlon)
			assert.False(t, latlon.isValid())
			continue
		}
		if c.expected.missingLatLon {
			assert.True(t, latlon.missingLatLon, "message: %s", c.message)
			continue
		}
		assert.False(
			t, latlon.isZero(),
			"message: %s, log: %+v, coord: %+v",
			c.message, c.logline, latlon)
		assert.Equal(
			t, c.expected.rawLat, latlon.rawLat,
			"message: %s, actual: %+v", c.message, latlon)
		assert.Equal(
			t, c.expected.rawLon, latlon.rawLon,
			"message: %s, actual: %+v", c.message, latlon)

		labels := convertLatLonToHash(nil, c.logline)
		assert.NotNil(t, labels)
		_, hasHash := labels[geoHash]
		assert.True(t, hasHash,
			"expecting resulting labels to include '%s', labels: %+v, from log line: %+v",
			geoHash, labels, c.logline)
	}
}

func TestConvertLatLonToHash_HandlesNilLabels(t *testing.T) {
	var currentLabels map[string]string
	var logline map[string]interface{}
	labels := convertLatLonToHash(currentLabels, logline)
	assert.NotNil(t, labels)
}

func TestConvertLatLonToHash_HandlesNilLogLine(t *testing.T) {
	var logline map[string]interface{}
	currentLabels := map[string]string{}
	labels := convertLatLonToHash(currentLabels, logline)
	assert.NotNil(t, labels)
}

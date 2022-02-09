package metrics

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errorParsingExpected = errors.New("error while parsing expected")

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
				parseError: errorParsingExpected,
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
				parseError: errorParsingExpected,
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
				parseError: errorParsingExpected,
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
				parseError: errorParsingExpected,
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
				parseError: errorParsingExpected,
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
				lat: "-33.86010",
				lon: "151.21010",
			},
		},
		{
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": "1.1,2.2",
				},
			},
			expected: coords{
				lat: "1.1",
				lon: "2.2",
			},
		},
	}
	for _, c := range cases {
		latlon := extractGeoip(c.logline)
		if c.isZero {
			assert.True(t, latlon.isZero(), "coord: %+v", latlon)
			continue
		}
		if c.expected.parseError != nil {
			assert.NotNil(t, latlon.parseError,
				"message: %s, actual: %+v", c.message, latlon)
			continue
		}
		if c.expected.missingGeo {
			assert.True(t,
				latlon.missingGeo && latlon.missingLatLon,
				"message: %s, actual: %+v", c.message, latlon)
			continue
		}
		if c.expected.missingLatLon {
			assert.True(t, latlon.missingLatLon, "message: %s", c.message)
			continue
		}
		assert.False(t, latlon.isZero(),
			"message: %s, log: %+v, coord: %+v",
			c.message, c.logline, latlon)
		assert.Equal(t, c.expected.lat, latlon.lat)
		assert.Equal(t, c.expected.lon, latlon.lon)
	}
}

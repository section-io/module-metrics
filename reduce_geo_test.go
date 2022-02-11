package metrics

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errorExtractingExpected = errors.New("error while extracting lat/lon from log line")
var errorConvertingExpected = errors.New("error while converting found lat/lon from log line")

func mockLogLineWithGeo(latlon string) map[string]interface{} {
	return map[string]interface{}{
		"geo": map[string]interface{}{
			geoLatLon: latlon,
		},
	}
}

func TestConvertLatLon_ProducesValidCoords(t *testing.T) {
	emptyLabels := map[string]string{}
	cases := []struct {
		message string
		logline map[string]interface{}
	}{
		{message: "with floats", logline: mockLogLineWithGeo("1.1,2.2")},
		{message: "with ints", logline: mockLogLineWithGeo("1.1,2.2")},
		{message: "with no precision", logline: mockLogLineWithGeo("1.,2.")},
		{message: "only no precision", logline: mockLogLineWithGeo(".1,.2")},
	}
	for _, c := range cases {
		labels, c := convertLatLonToHash(emptyLabels, c.logline)
		assert.True(t, c.isValid(), "expected an valid coord %+v", c)
		_, ok := labels[geoHash]
		assert.True(t, ok)
	}
}

func TestConvertLatLon_ProducesInvalidCoords(t *testing.T) {
	emptyLabels := map[string]string{}
	cases := []struct {
		message string
		logline map[string]interface{}
	}{
		{message: "empty latlon", logline: mockLogLineWithGeo("")},
		{message: "only lat", logline: mockLogLineWithGeo("1.1,")},
		{message: "only lon", logline: mockLogLineWithGeo(",1.1")},
		{message: "too many parts", logline: mockLogLineWithGeo("1.1,1.1,")},
		{message: "no geo object", logline: map[string]interface{}{}},
		{message: "no geo latlon", logline: map[string]interface{}{
			"geo": map[string]interface{}{},
		}},
	}
	for _, c := range cases {
		_, coord := convertLatLonToHash(emptyLabels, c.logline)
		assert.False(t, coord.isValid(), "expected an invalid coord %+v", c)
		coord.logErrors(c.logline, t.Logf)
	}
}

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
			message: "empty lat provided",
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
			message: "empty lon provided",
			logline: map[string]interface{}{
				"geo": map[string]interface{}{
					"latlon": "-33.86010,",
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
		{
			message: "geo is not a map",
			logline: map[string]interface{}{
				"geo": 0,
			},
			expected: coords{
				missingGeo: true,
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

		labels, _ := convertLatLonToHash(nil, c.logline)
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
	labels, _ := convertLatLonToHash(currentLabels, logline)
	assert.NotNil(t, labels)
}

func TestConvertLatLonToHash_HandlesNilLogLine(t *testing.T) {
	currentLabels := map[string]string{}
	var logline map[string]interface{}
	labels, _ := convertLatLonToHash(currentLabels, logline)
	assert.NotNil(t, labels)
}

func TestScrubLatLon(t *testing.T) {
	cases := []struct {
		message string
		labels  map[string]string
	}{
		{
			message: "doesn't contain geoLat or geoLon to begin with",
			labels: map[string]string{
				geoHash:        "abc",
				"cluster_name": "do-sfo-k9",
			},
		},
		{
			message: "doesn't contains lat",
			labels: map[string]string{
				geoHash:        "abc",
				"cluster_name": "do-sfo-k9",
				"lat":          "1.1",
			},
		},
		{
			message: "doesn't contains lon",
			labels: map[string]string{
				geoHash:        "abc",
				"cluster_name": "do-sfo-k9",
				"lon":          "1.1",
			},
		},
	}
	for _, c := range cases {
		actual := scrubGeoHash(c.labels)
		_, hasGeoHash := actual[geoHash]
		assert.False(t, hasGeoHash)
	}
}

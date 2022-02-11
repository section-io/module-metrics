package metrics

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mmcloughlin/geohash"
)

var missingLogFields = coords{
	missingLatLon: true,
	missingGeo:    true,
}

type coords struct {
	lat           float64
	lon           float64
	rawLat        string
	rawLon        string
	missingLatLon bool
	missingGeo    bool
	extractError  error
	convertError  error
}

func (c coords) isZero() bool {
	return strings.TrimSpace(c.rawLat) == "" &&
		strings.TrimSpace(c.rawLon) == "" &&
		c.lat == 0.0 &&
		c.lon == 0.0 &&
		c.rawLat == "" &&
		c.rawLon == "" &&
		c.extractError == nil &&
		c.convertError == nil
}

func (c coords) isValid() bool {
	return !c.missingGeo && !c.missingLatLon &&
		c.extractError == nil &&
		c.convertError == nil
}

func (c coords) logErrors(logline map[string]interface{}, logf func(f string, args ...interface{})) {
	if c.missingGeo {
		logf("missing geo object (geo: %+v)", logline[geoHash])
	}
	if c.missingLatLon {
		logf("missing lat/lon (geo: %+v)", logline[geoHash])
	}
	if c.extractError != nil {
		logf("parse error (geo: %+v)(error: %+v)", logline[geoHash], c.extractError)
	}
	if c.convertError != nil {
		logf("converting raw lat/lon error: (error: %+v)", c.convertError)
	}
}

func convertLatLon(rawLat, rawLon string) (float64, float64, error) {
	lat, latErr := strconv.ParseFloat(rawLat, 64)
	lon, lonErr := strconv.ParseFloat(rawLon, 64)
	if latErr != nil || lonErr != nil {
		err := fmt.Errorf("%+v and %+v", latErr, lonErr)
		return lat, lon, err
	}
	return lat, lon, nil
}

func extractGeoip(logline map[string]interface{}) coords {
	rawGeo, ok := logline["geo"]
	if !ok {
		return missingLogFields
	}
	geo, isMap := rawGeo.(map[string]interface{})
	if !isMap {
		return missingLogFields
	}
	rawLatLon, hasLatLon := geo["latlon"]
	latlon, isLatLonString := rawLatLon.(string)
	rawLat, rawLon, extractErr := extractLatLon(latlon)
	lat, lon, convertErr := convertLatLon(rawLat, rawLon)
	return coords{
		lat:           lat,
		lon:           lon,
		rawLat:        rawLat,
		rawLon:        rawLon,
		missingLatLon: !isLatLonString && !hasLatLon,
		missingGeo:    !isMap,
		extractError:  extractErr,
		convertError:  convertErr,
	}
}

func extractLatLon(latlon string) (string, string, error) {
	parts := strings.Split(latlon, ",")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("found more extrat lat/lon info: %s", latlon)
	}
	lat, lon := parts[0], parts[1]
	if strings.TrimSpace(lat) == "" {
		return lat, lon, fmt.Errorf("did not parse a value for lat")
	}
	if strings.TrimSpace(lon) == "" {
		return lat, lon, fmt.Errorf("did not parse a value for lon")
	}
	return lat, lon, nil
}

func convertLatLonToHash(labels map[string]string, logline map[string]interface{}) (map[string]string, coords) {
	if labels == nil {
		return map[string]string{geoHash: geoMissing}, coords{}
	}
	c := extractGeoip(logline)
	if !c.isValid() {
		labels[geoHash] = geoMissing
		return labels, c
	}
	hash := geohash.EncodeWithPrecision(c.lat, c.lon, effectiveHashPrecision)
	labels[geoHash] = hash
	return labels, c
}

// scrubGeoHash must create a new map since the values of the
// provided map could be used at some future time after having been
// previously provided to other calls since passed by pointer
func scrubGeoHash(labels map[string]string) map[string]string {
	rv := map[string]string{}
	for k, v := range labels {
		switch k {
		case geoHash: // remove/scrub these keys
			continue
		default:
			rv[k] = v
		}
	}
	return rv
}

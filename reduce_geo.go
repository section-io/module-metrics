package metrics

import (
	"errors"
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
	lat           string
	lon           string
	missingLatLon bool
	missingGeo    bool
	parseError    error
}

func (c coords) isZero() bool {
	return strings.TrimSpace(c.lat) == "" && strings.TrimSpace(c.lon) == ""
	//&&        !c.missingLat && !c.missingLon && !c.missingGeo
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
	lat, lon, err := extractLatLon(latlon)
	return coords{
		lat:           lat,
		lon:           lon,
		missingLatLon: !hasLatLon && !isLatLonString,
		missingGeo:    !isMap,
		parseError:    err,
	}
	panic(errors.New("not implemented"))
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

func convertLatLon(rawLat, rawLon string) (float64, float64, error) {
	lat, latErr := strconv.ParseFloat(rawLat, 64)
	lon, lonErr := strconv.ParseFloat(rawLon, 64)
	if latErr != nil || lonErr != nil {
		err := errors.New(
			fmt.Sprintf("%+v and %+v", latErr, lonErr),
		)
		return lat, lon, err
	}
	return lat, lon, nil
}

func reduceGeoHashLabels(labels map[string]string) map[string]string {
	rawLat, hasLat := labels[geoLat]
	rawLon, hasLon := labels[geoLon]
	if !hasLat || !hasLon {
		//TODO: actually log some kind of error instead of making this
		//obviously bad label
		labels[geoHash] = fmt.Sprintf("%s, %s", rawLat, rawLon)
	}
	lat, lon, err := convertLatLon(rawLat, rawLon)
	if err != nil {
		//TODO: actually log some kind of error instead of making this
		//obviously bad label
		labels[geoHash] = fmt.Sprintf("%s, %s", rawLat, rawLon)
		return labels
	}
	hash := geohash.EncodeWithPrecision(lat, lon, geoHashPrecision)
	labels[geoHash] = hash
	return labels
}

// scrubGeoHashAndLatLon must create a new map since the values of the
// provided map could be used at some future time for after having
// been previously provided to other calls since passed by pointer
func scrubGeoHashAndLatLon(labels map[string]string) map[string]string {
	rv := map[string]string{}
	for k, v := range labels {
		switch k {
		case geoHash, geoLat, geoLon: // remove/scrub these keys
			continue
		default:
			rv[k] = v
		}
	}
	return rv
}

// scrubLatLon must create a new map since the values of the
// provided map could be used at some future time for after having
// been previously provided to other calls since passed by pointer
func scrubLatLon(labels map[string]string) map[string]string {
	rv := map[string]string{}
	for k, v := range labels {
		switch k {
		case geoLat, geoLon: // remove/scrub these keys
			continue
		default:
			rv[k] = v
		}
	}
	return rv
}

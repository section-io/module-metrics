package metrics

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/mmcloughlin/geohash"
)

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

func reduceGeoHashLabels(labels map[string]string, stderr io.Writer) map[string]string {
	rawLat, hasLat := labels["lat"]
	rawLon, hasLon := labels["lon"]
	if !hasLat || !hasLon {
		labels[geo_hash] = fmt.Sprintf("%s, %s", rawLat, rawLon)
	}
	lat, lon, err := convertLatLon(rawLat, rawLon)
	if err != nil {
		labels[geo_hash] = fmt.Sprintf("%s, %s", rawLat, rawLon)
		return labels
	}
	hash := geohash.EncodeWithPrecision(lat, lon, precision)
	labels[geo_hash] = hash
	return labels
}

// scrubGeoHashAndLatLon must create a new map since the values of the
// provided map could be used at some future time for after having
// been previously provided to other calls since passed by pointer
func scrubGeoHashAndLatLon(labels map[string]string) map[string]string {
	rv := map[string]string{}
	for k, v := range labels {
		switch k {
		case geo_hash, geo_lat, geo_lon: // remove/scrub these keys
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
		case geo_lat, geo_lon: // remove/scrub these keys
			continue
		default:
			rv[k] = v
		}
	}
	return rv
}

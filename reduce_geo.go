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

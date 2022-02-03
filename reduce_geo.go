package metrics

import (
	"strconv"

	"github.com/mmcloughlin/geohash"
)

func convertLatLon(rawLat, rawLon string) (float64, float64, error) {
	lat, err := strconv.ParseFloat(rawLat, 64)
	if err != nil {
		return 0.0, 0.0, err
	}
	lon, err := strconv.ParseFloat(rawLon, 64)
	if err != nil {
		return 0.0, 0.0, err
	}
	return lat, lon, nil
}

func reduceGeoHashLabels(labels map[string]string) map[string]string {
	rawLat, ok := labels["lat"]
	if !ok {
		return labels
	}
	rawLon, ok := labels["lon"]
	if !ok {
		return labels
	}
	lat, lon, err := convertLatLon(rawLat, rawLon)
	if err != nil {
		return labels
	}
	hash := geohash.EncodeWithPrecision(lat, lon, precision)
	labels[geo_hash] = hash
	return labels
}

func scrubGeoHashLabels(labels map[string]string) map[string]string {
	delete(labels, geo_lat)
	delete(labels, geo_lon)
	delete(labels, geo_hash)
	return labels
}

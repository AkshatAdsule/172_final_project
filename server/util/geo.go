package util

import "math"

// haversineDistance calculates the distance between two GPS coordinates
// (lat1, lon1) and (lat2, lon2) in meters.
// Latitude and longitude are expected in degrees.
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371e3 // Earth radius in meters

	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	return distance
}

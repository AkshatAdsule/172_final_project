package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	pstOffsetSeconds = -7 * 60 * 60 // PST is UTC-7
)

var pstLocation *time.Location

func init() {
	pstLocation = time.FixedZone("PST", pstOffsetSeconds)
}

// CombineDateTime takes a baseTime (primarily for its date part in UTC)
// and a timeStr in "HHMMSS.SS" format (assumed UTC),
// and returns a new time.Time object in UTC.
func CombineDateTime(baseTime time.Time, timeStr string) (time.Time, error) {
	parts := strings.Split(timeStr, ".")
	if len(parts) == 0 {
		return time.Time{}, fmt.Errorf("invalid time string format: %s", timeStr)
	}
	hmsStr := parts[0]
	if len(hmsStr) != 6 {
		return time.Time{}, fmt.Errorf("invalid HHMSS part of time string: %s", hmsStr)
	}

	hour, err := strconv.Atoi(hmsStr[0:2])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid hour: %s", hmsStr[0:2])
	}
	minute, err := strconv.Atoi(hmsStr[2:4])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid minute: %s", hmsStr[2:4])
	}
	second, err := strconv.Atoi(hmsStr[4:6])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid second: %s", hmsStr[4:6])
	}

	var millisecond int
	if len(parts) > 1 && len(parts[1]) > 0 {
		centiSeconds, err := strconv.Atoi(parts[1])
		if err != nil {
			if len(parts[1]) == 3 {
				millisecond, err = strconv.Atoi(parts[1])
				if err != nil {
					return time.Time{}, fmt.Errorf("invalid fractional second: %s", parts[1])
				}
			} else {
				return time.Time{}, fmt.Errorf("invalid fractional second: %s", parts[1])
			}
		} else {
			millisecond = centiSeconds * 10 // convert centiseconds to milliseconds
		}
	}

	return time.Date(
		baseTime.Year(), baseTime.Month(), baseTime.Day(),
		hour, minute, second, millisecond*int(time.Millisecond),
		time.UTC), nil
}

// GetRideName determines the ride name based on the start time in PST.
func GetRideName(startTimeUTC time.Time) string {
	// Convert start time to PST for naming
	startTimePST := startTimeUTC.In(pstLocation)
	hour := startTimePST.Hour()

	switch {
	case hour >= 6 && hour < 12:
		return "Morning Ride"
	case hour >= 12 && hour < 18:
		return "Afternoon Ride"
	case hour >= 18 && hour < 22:
		return "Evening Ride"
	default:
		return "Night Ride"
	}
}

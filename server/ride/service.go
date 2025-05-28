package ride

import (
	"b3/server/config"
	"b3/server/models"
	"b3/server/util"
	"log"
	"time"
)

// RideState represents the current state of ride tracking.
type RideState int

const (
	StateIdle     RideState = iota // Not currently in a ride
	StateTracking                  // Currently tracking a ride
	StatePaused                    // Ride ongoing, but temporarily static (potential end)
)

// ProcessGPSUpdate determines if a new GPS point triggers a ride state change.
// It returns the new ride state and a boolean indicating if a ride started or ended.
// This function is stateless itself but operates based on provided current state.
func ProcessGPSUpdate(
	currentState RideState,
	lastPosition *models.Position, // Can be nil if no previous point or in Idle state
	currentPosition models.Position,
	lastMoveTime time.Time, // Time of the last significant movement or ride start
	cfg config.Config,
) (newState RideState, rideEventOccurred bool, eventType string) {

	if lastPosition == nil { // First point ever, or after a long pause
		log.Println("ProcessGPSUpdate: No last position, storing first point and remaining IDLE.")
		return StateIdle, false, ""
	}

	distance := util.HaversineDistance(lastPosition.Latitude, lastPosition.Longitude, currentPosition.Latitude, currentPosition.Longitude)
	log.Printf("ProcessGPSUpdate: Distance from last position: %.2fm", distance)

	switch currentState {
	case StateIdle:
		if distance >= cfg.RideStartDistance {
			log.Printf("ProcessGPSUpdate: IDLE -> TRACKING. Distance: %.2fm (>= threshold %.2fm)", distance, cfg.RideStartDistance)
			return StateTracking, true, "ride_started"
		}
		log.Printf("ProcessGPSUpdate: IDLE -> IDLE. Distance: %.2fm (< threshold %.2fm)", distance, cfg.RideStartDistance)
		return StateIdle, false, ""

	case StateTracking:
		if distance < cfg.RideEndStaticDist { // Not much movement
			log.Printf("ProcessGPSUpdate: TRACKING -> PAUSED. Distance: %.2fm (below static threshold %.2fm)", distance, cfg.RideEndStaticDist)
			return StatePaused, false, ""
		}
		log.Printf("ProcessGPSUpdate: TRACKING -> TRACKING. Distance: %.2fm", distance)
		return StateTracking, false, ""

	case StatePaused:
		if distance >= cfg.RideStartDistance { // Movement resumed
			log.Printf("ProcessGPSUpdate: PAUSED -> TRACKING. Distance: %.2fm", distance)
			return StateTracking, false, ""
		}
		log.Printf("ProcessGPSUpdate: PAUSED -> PAUSED. Distance: %.2fm", distance)
		return StatePaused, false, ""
	}
	log.Printf("ProcessGPSUpdate: Unhandled state %v or fallthrough", currentState)
	return currentState, false, ""
}

// ShouldEndRideDueToInactivity checks if a ride should end based on inactivity timeout.
func ShouldEndRideDueToInactivity(currentState RideState, lastUpdateTime, pausedSince time.Time, cfg config.Config) bool {
	now := time.Now().UTC()

	if now.Sub(lastUpdateTime) > time.Duration(cfg.RideEndInactivity)*time.Second {
		log.Printf("Ride should end: General inactivity. Last update: %v, Current time: %v, Timeout: %ds",
			lastUpdateTime, now, cfg.RideEndInactivity)
		return true
	}

	if currentState == StatePaused && !pausedSince.IsZero() {
		if now.Sub(pausedSince) > time.Duration(cfg.RideEndStaticSecs)*time.Second {
			log.Printf("Ride should end: Static for too long. Paused since: %v, Current time: %v, Static timeout: %ds",
				pausedSince, now, cfg.RideEndStaticSecs)
			return true
		}
	}
	return false
}

// DetermineRideName assigns a name to the ride based on its start time (in UTC).
func DetermineRideName(startTimeUTC time.Time, cfg config.Config) string {
	startTimeLocalized := startTimeUTC.In(cfg.PSTLocation)
	hour := startTimeLocalized.Hour()

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

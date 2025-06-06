package ride

import (
	"b3/server/config"
	"b3/server/database"
	"b3/server/models"
	"b3/server/ws"

	"database/sql"
	"log"
	"sync"
	"time"
)

// RideManager handles the business logic of ride tracking.
type RideManager struct {
	mu             sync.Mutex
	currentState   RideState
	currentRideID  int64
	lastPosition   *models.Position
	rideStartTime  time.Time
	pausedSince    time.Time // When the ride entered PAUSED state
	lastUpdateTime time.Time // Timestamp of the last processed GPS point
	db             *sql.DB
	cfg            config.Config
	hub            *ws.Hub                                     // WebSocket hub for broadcasting
	lockStatus     string                                      // Current lock status: "LOCKED" or "UNLOCKED"
	theftAlertFunc func(lat, lon float64, timestamp time.Time) // Function to call for theft alerts
}

// NewRideManager creates a new RideManager.
func NewRideManager(db *sql.DB, appConfig config.Config, hub *ws.Hub) *RideManager { // Added hub parameter
	return &RideManager{
		currentState:   StateIdle,
		db:             db,
		cfg:            appConfig,
		hub:            hub,        // Assign hub
		lockStatus:     "UNLOCKED", // Initialize to unlocked
		theftAlertFunc: nil,        // Will be set separately if needed
	}
}

// HandleGPSData processes an incoming GPS data point.
// This is the main entry point for new data from MQTT.
func (rm *RideManager) HandleGPSData(point models.Position) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	cfg := rm.cfg // Use the stored config

	// 1. Check for ride end due to general inactivity before processing the new point
	// This handles cases where GPS data stops entirely for a while.
	if rm.currentState != StateIdle && !rm.lastUpdateTime.IsZero() {
		if ShouldEndRideDueToInactivity(rm.currentState, rm.lastUpdateTime, rm.pausedSince, cfg) {
			log.Printf("RideManager: Ride %d ending due to inactivity.", rm.currentRideID)
			rm.endCurrentRide(rm.lastUpdateTime)                           // End ride with the timestamp of the last known point
			rm.hub.BroadcastRideEnded(rm.currentRideID, rm.lastUpdateTime) // Uncommented
			rm.resetRideState()
			// After resetting, we might still process the current point if it's a new start
		}
	}

	// Update lastUpdateTime with the current point's timestamp
	rm.lastUpdateTime = point.Timestamp

	// Broadcast current location to all WebSocket clients
	rm.hub.BroadcastCurrentLocation(point)

	// 2. Process the current GPS point using the stateless service logic
	previousState := rm.currentState
	newState, eventOccurred, eventType := ProcessGPSUpdate(
		rm.currentState,
		rm.lastPosition,
		point,
		rm.rideStartTime, // Or time of last significant move, if we track that more granularly
		cfg,
	)
	rm.currentState = newState

	// 3. Handle state transitions and events
	if eventOccurred {
		if eventType == "ride_started" {
			// Check if bike is locked - if so, this is potential theft
			if rm.lockStatus == "LOCKED" {
				log.Printf("THEFT DETECTION: Movement detected while bike is locked! Location: lat %f, lon %f",
					point.Latitude, point.Longitude)
				// Send theft alert if alert function is set
				if rm.theftAlertFunc != nil {
					rm.theftAlertFunc(point.Latitude, point.Longitude, point.Timestamp)
				}
				return // Exit early, don't start a ride in lock mode
			}
			rm.startNewRide(point)
			// Notify via WebSocket (TODO)
			// rm.hub.BroadcastRideStarted(rm.currentRideID, rm.currentRideName, rm.rideStartTime, point)
		}
		// "ride_ended" from ProcessGPSUpdate is not expected here, as it's handled by ShouldEndRideDueToInactivity
		// or by the PAUSED state timeout logic below.
	}

	// If we are tracking or paused, add the point to the current ride
	if rm.currentState == StateTracking || rm.currentState == StatePaused {
		if rm.currentRideID != 0 { // Ensure ride has been created
			err := database.AddPositionToRide(rm.db, rm.currentRideID, point)
			if err != nil {
				log.Printf("Error adding position to ride %d: %v", rm.currentRideID, err)
				// Decide on error handling: continue, try to rollback, etc. For now, just log.
			} else {
				rm.hub.BroadcastRidePositionAdded(rm.currentRideID, point) // Uncommented
				log.Printf("Added position (%f, %f) to ride %d", point.Latitude, point.Longitude, rm.currentRideID)
			}
		} else {
			log.Println("Warning: In tracking/paused state but no currentRideID. GPS point not saved.")
		}
	}

	// 4. Specific logic for PAUSED state
	if previousState != StatePaused && rm.currentState == StatePaused {
		// Just entered paused state
		rm.pausedSince = time.Now().UTC() // Record when pause began
		log.Printf("RideManager: Ride %d entered PAUSED state at %v.", rm.currentRideID, rm.pausedSince)
	} else if rm.currentState == StatePaused {
		// Still in paused state, check if static timeout is exceeded
		if !rm.pausedSince.IsZero() && time.Now().UTC().Sub(rm.pausedSince) > time.Duration(cfg.RideEndStaticSecs)*time.Second {
			log.Printf("RideManager: Ride %d ending due to being static for too long (paused). Paused since: %v", rm.currentRideID, rm.pausedSince)
			rm.endCurrentRide(point.Timestamp)                           // End with current point's timestamp
			rm.hub.BroadcastRideEnded(rm.currentRideID, point.Timestamp) // Uncommented
			rm.resetRideState()
		}
	}

	// If we transitioned from PAUSED to TRACKING, clear pausedSince
	if previousState == StatePaused && rm.currentState == StateTracking {
		rm.pausedSince = time.Time{} // Zero value for time
		log.Printf("RideManager: Ride %d resumed from PAUSED to TRACKING.", rm.currentRideID)
	}

	// Update last position
	rm.lastPosition = &point
}

// CheckInactivityLoop is intended to be run as a goroutine to periodically
// check for ride endings due to prolonged inactivity, even if no new GPS points are coming in.
func (rm *RideManager) CheckInactivityLoop(tickerDuration time.Duration) {
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()

	for range ticker.C {
		rm.mu.Lock()
		if rm.currentState != StateIdle && !rm.lastUpdateTime.IsZero() {
			if ShouldEndRideDueToInactivity(rm.currentState, rm.lastUpdateTime, rm.pausedSince, rm.cfg) {
				log.Printf("RideManager (InactivityLoop): Ride %d ending due to inactivity.", rm.currentRideID)

				// Determine end time: if paused, use pausedSince + static duration, else lastUpdateTime + inactivity duration
				// Or simply, the last effective point's time before timeout. For now, use lastUpdateTime.
				endTime := rm.lastUpdateTime
				if rm.currentState == StatePaused && !rm.pausedSince.IsZero() {
					// If it ended due to static timeout, the effective end time is when that timeout was breached.
					// This might be slightly different from just lastUpdateTime if no new points came in.
					// However, ShouldEndRideDueToInactivity uses time.Now(), so using lastUpdateTime is simpler.
				}

				rm.endCurrentRide(endTime)
				rm.hub.BroadcastRideEnded(rm.currentRideID, endTime) // Uncommented
				rm.resetRideState()
			}
		}
		rm.mu.Unlock()
	}
}

func (rm *RideManager) startNewRide(currentPosition models.Position) {
	// This function assumes rm.mu is already locked.
	rm.rideStartTime = currentPosition.Timestamp
	rideName := DetermineRideName(rm.rideStartTime, rm.cfg)

	id, err := database.CreateRide(rm.db, rideName, rm.rideStartTime)
	if err != nil {
		log.Printf("Error creating new ride in database: %v", err)
		rm.resetRideState() // Go back to idle if DB operation fails
		return
	}
	rm.currentRideID = id
	rm.currentState = StateTracking
	rm.pausedSince = time.Time{} // Clear any previous paused time

	log.Printf("Started new ride: ID %d, Name: %s, StartTime: %v", id, rideName, rm.rideStartTime)
	rm.hub.BroadcastRideStarted(rm.currentRideID, rideName, rm.rideStartTime, currentPosition) // Uncommented

	// Add the first point to this new ride
	err = database.AddPositionToRide(rm.db, rm.currentRideID, currentPosition)
	if err != nil {
		log.Printf("Error adding initial position to ride %d: %v", rm.currentRideID, err)
		// Potentially rollback ride creation or mark it as problematic
	} else {
		log.Printf("Added initial position (%f, %f) to ride %d", currentPosition.Latitude, currentPosition.Longitude, rm.currentRideID)
	}
}

func (rm *RideManager) endCurrentRide(endTime time.Time) {
	// This function assumes rm.mu is already locked.
	if rm.currentRideID == 0 {
		log.Println("endCurrentRide called but no current ride ID.")
		return
	}
	err := database.EndRide(rm.db, rm.currentRideID, endTime)
	if err != nil {
		log.Printf("Error ending ride %d in database: %v", rm.currentRideID, err)
	} else {
		log.Printf("Ended ride: ID %d, EndTime: %v", rm.currentRideID, endTime)
	}
}

func (rm *RideManager) resetRideState() {
	// This function assumes rm.mu is already locked.
	rm.currentState = StateIdle
	rm.currentRideID = 0
	rm.rideStartTime = time.Time{}
	rm.pausedSince = time.Time{}
	log.Println("RideManager state reset to Idle.")
}

// SetLockStatus updates the lock status of the bike
func (rm *RideManager) SetLockStatus(status string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.lockStatus = status
	log.Printf("Lock status updated to: %s", status)
}

// GetLockStatus returns the current lock status
func (rm *RideManager) GetLockStatus() string {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.lockStatus
}

// SetTheftAlertFunc sets the function to call when theft is detected
func (rm *RideManager) SetTheftAlertFunc(alertFunc func(lat, lon float64, timestamp time.Time)) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.theftAlertFunc = alertFunc
}

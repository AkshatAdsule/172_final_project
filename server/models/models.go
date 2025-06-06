package models

import "time"

// Position represents a single GPS data point.
type Position struct {
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	SpeedKnots float64   `json:"speed_knots,omitempty"` // omitempty as it's not always used in all contexts
	Timestamp  time.Time `json:"timestamp"`             // UTC
}

// RideSummary provides a brief overview of a ride.
type RideSummary struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`         // UTC
	EndTime   time.Time `json:"end_time,omitempty"` // UTC, omitempty if ride is ongoing
}

// RideDetail provides a comprehensive view of a ride, including all its positions.
type RideDetail struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	StartTime time.Time  `json:"start_time"`         // UTC
	EndTime   time.Time  `json:"end_time,omitempty"` // UTC, omitempty if ride is ongoing
	Positions []Position `json:"positions"`
}

// WebSocketMessage is a generic structure for messages sent over WebSocket.
// We can define more specific payload structures if needed.
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WSLocationPayload is for the 'current_location' WebSocket message.
type WSLocationPayload struct {
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Timestamp  time.Time `json:"timestamp"` // UTC
	SpeedKnots float64   `json:"speed_knots,omitempty"`
}

// WSRideStartedPayload is for the 'ride_started' WebSocket message.
type WSRideStartedPayload struct {
	RideID          int64     `json:"ride_id"`
	Name            string    `json:"name"`
	StartTime       time.Time `json:"start_time"` // UTC
	InitialPosition Position  `json:"initial_position"`
}

// WSRidePositionAddedPayload is for the 'ride_position_added' WebSocket message.
type WSRidePositionAddedPayload struct {
	RideID   int64    `json:"ride_id"`
	Position Position `json:"position"`
}

// WSRideEndedPayload is for the 'ride_ended' WebSocket message.
type WSRideEndedPayload struct {
	RideID  int64     `json:"ride_id"`
	EndTime time.Time `json:"end_time"` // UTC
}

package ws

import (
	"b3/server/models"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// upgrader is shared for all WebSocket connections handled by this hub.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Hub maintains the set of active clients and broadcasts messages to the
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub's event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Println("Client registered to hub")
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("Client unregistered from hub")
			}
		}
	}
}

// BroadcastMessage sends a message to all connected clients.
func (h *Hub) BroadcastMessage(messageType string, payload interface{}) {
	message := map[string]interface{}{
		"type":      messageType,
		"payload":   payload,
		"timestamp": time.Now().UTC(),
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling broadcast message: %v", err)
		return
	}

	log.Printf("Broadcasting message: %s", string(jsonMessage))
	for client := range h.clients {
		select {
		case client.send <- jsonMessage:
		default: // If client's send buffer is full, assume it's dead/stuck.
			log.Printf("Client send channel full or closed. Unregistering client.")
			close(client.send)
			delete(h.clients, client) // Important to prevent leaks and repeated attempts
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	hub.register <- client

	go client.writePump()
	go client.readPump()
	log.Println("ServeWs: Client created and pumps started, registration sent to hub.")
}

// RideEventPayload is a generic structure for ride event payloads
type RideEventPayload struct {
	RideID    int64            `json:"ride_id"`
	RideName  string           `json:"ride_name,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
	Position  *models.Position `json:"position,omitempty"` // For position updates
}

// BroadcastRideStarted sends a message when a new ride starts.
func (h *Hub) BroadcastRideStarted(rideID int64, rideName string, startTime time.Time, initialPosition models.Position) {
	payload := RideEventPayload{
		RideID:    rideID,
		RideName:  rideName,
		Timestamp: startTime,
		Position:  &initialPosition,
	}
	h.BroadcastMessage("RIDE_STARTED", payload)
}

// BroadcastRideEnded sends a message when a ride ends.
func (h *Hub) BroadcastRideEnded(rideID int64, endTime time.Time) {
	payload := RideEventPayload{
		RideID:    rideID,
		Timestamp: endTime,
	}
	h.BroadcastMessage("RIDE_ENDED", payload)
}

// BroadcastRidePositionAdded sends a message when a new position is added to an ongoing ride.
func (h *Hub) BroadcastRidePositionAdded(rideID int64, position models.Position) {
	payload := RideEventPayload{
		RideID:    rideID,
		Timestamp: position.Timestamp,
		Position:  &position,
	}
	h.BroadcastMessage("RIDE_POSITION_UPDATE", payload)
}

// BroadcastCurrentLocation sends a message for every valid GPS update received from MQTT.
func (h *Hub) BroadcastCurrentLocation(position models.Position) {
	payload := models.WSLocationPayload{
		Latitude:   position.Latitude,
		Longitude:  position.Longitude,
		Timestamp:  position.Timestamp,
		SpeedKnots: position.SpeedKnots,
	}
	h.BroadcastMessage("current_location", payload)
}

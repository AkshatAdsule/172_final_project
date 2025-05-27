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
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients. (Not used in this app, but good for structure)
	// broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		// broadcast:  make(chan []byte), // Not directly used for now
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
			// case message := <-h.broadcast: // Generic broadcast, not used for specific ride events
			// 	for client := range h.clients {
			// 		select {
			// 		case client.send <- message:
			// 		default:
			// 			close(client.send)
			// 			delete(h.clients, client)
			// 		}
			// 	}
		}
	}
}

// BroadcastMessage sends a message to all connected clients.
// This is a generic broadcaster. We'll add specific event broadcasters.
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
	// newClient is defined in client.go, it will register itself via c.hub.register
	// when its readPump starts if we design it that way, or we register here.
	// For now, newClient (as currently defined) will need the hub to then register itself.
	// The current newClient starts its own read/write pumps.

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	hub.register <- client

	// The newClient function in client.go already starts these goroutines.
	// If we call newClient directly, it handles this.
	// Let's stick to current newClient which is not exported and called internally by client.go logic that isn't directly used here.
	// Instead, we construct the client here and then start its pumps.
	go client.writePump()
	go client.readPump()
	log.Println("ServeWs: Client created and pumps started, registration sent to hub.")
}

// Specific event broadcasting methods for RideManager to call:

// RideEventPayload is a generic structure for ride event payloads
type RideEventPayload struct {
	RideID    int64            `json:"ride_id"`
	RideName  string           `json:"ride_name,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
	Position  *models.Position `json:"position,omitempty"` // For position updates
	// Add other relevant fields as needed
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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"b3/server/mqttsubscriber" // Import the new package

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ShadowStateDesired represents the "desired" part of the AWS IoT shadow state.
// We can add more specific fields if needed, for now, it's a generic map.
type ShadowStateDesired map[string]interface{}

// ShadowState holds the overall state from the shadow document.
type ShadowState struct {
	Desired  ShadowStateDesired     `json:"desired"`
	Reported map[string]interface{} `json:"reported"`
}

// ShadowDocument is the top-level structure of the AWS IoT device shadow.
type ShadowDocument struct {
	State     ShadowState            `json:"state"`
	Metadata  map[string]interface{} `json:"metadata"`
	Version   int                    `json:"version"`
	Timestamp int64                  `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

func handleWebSocket(ctx *gin.Context, msgChan <-chan []byte, errChan <-chan error) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket client connected")

	// Goroutine to read messages from MQTT and send to WebSocket client
	go func() {
		for {
			select {
			case message, ok := <-msgChan:
				if !ok {
					log.Println("MQTT message channel closed, closing WebSocket.")
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				log.Printf("Received raw MQTT message: %s", string(message))

				var shadowDoc ShadowDocument
				if err := json.Unmarshal(message, &shadowDoc); err != nil {
					log.Printf("Error unmarshalling shadow document: %v. Sending raw message.", err)
					// Send raw message if unmarshalling fails
					if writeErr := conn.WriteMessage(websocket.TextMessage, message); writeErr != nil {
						log.Printf("Error writing raw message to WebSocket: %v", writeErr)
						return
					}
					continue // Continue to next message
				}

				if shadowDoc.State.Desired == nil {
					log.Println("No 'desired' state in message, sending empty object.")
					desiredMsg, _ := json.Marshal(map[string]interface{}{})
					if writeErr := conn.WriteMessage(websocket.TextMessage, desiredMsg); writeErr != nil {
						log.Printf("Error writing empty desired state to WebSocket: %v", writeErr)
						return
					}
					continue
				}

				desiredStateBytes, err := json.Marshal(shadowDoc.State.Desired)
				if err != nil {
					log.Printf("Error marshalling desired state: %v. Sending raw message.", err)
					// Send raw message if marshalling desired state fails
					if writeErr := conn.WriteMessage(websocket.TextMessage, message); writeErr != nil {
						log.Printf("Error writing raw message to WebSocket: %v", writeErr)
						return
					}
					continue
				}

				log.Printf("Sending desired state to WebSocket: %s", string(desiredStateBytes))
				err = conn.WriteMessage(websocket.TextMessage, desiredStateBytes)
				if err != nil {
					log.Printf("Error writing desired state to WebSocket: %v", err)
					return // Stop sending if there's an error
				}
			case err, ok := <-errChan:
				if !ok {
					log.Println("MQTT error channel closed.")
					// Optionally send an error message to the client before closing
					return
				}
				log.Printf("Error from MQTT subscriber: %v. Closing WebSocket.", err)
				// Send an error message to client or just close
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, fmt.Sprintf("MQTT Error: %v", err)))
				return
			case <-ctx.Request.Context().Done():
				log.Println("Client disconnected (context done).")
				return
			}
		}
	}()

	// Keep the connection alive by reading messages from the client (optional)
	// This also helps detect if the client has disconnected.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf("WebSocket read error (client likely disconnected): %v", err)
			break // Exit when client disconnects or an error occurs
		}
	}
	log.Println("WebSocket client disconnected")
}

func main() {
	// MQTT Setup from Old_main
	brokerURL := "tls://a1edew9tp1yb1x-ats.iot.us-east-1.amazonaws.com:8883"
	clientID := "server-main" // Changed clientID to avoid conflict if Old_main is run
	topic := "$aws/things/akshat_cc3200board/shadow/update/accepted"
	certPath := "certs/certificate.pem.crt"
	keyPath := "certs/private.pem.key"
	caPath := "certs/AmazonRootCA1.pem"

	msgChan, errChan, closeFn, err := mqttsubscriber.SubscribeToShadowUpdates(brokerURL, clientID, topic, certPath, keyPath, caPath)
	if err != nil {
		log.Fatalf("Failed to subscribe to shadow updates: %v", err)
	}
	// defer closeFn() // We'll call this on sigterm

	fmt.Println("MQTT Listener started. Press Ctrl+C to stop server.")

	// Gin Router Setup
	router := gin.Default()

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// WebSocket endpoint
	router.GET("/ws", func(ctx *gin.Context) {
		handleWebSocket(ctx, msgChan, errChan)
	})

	// Start server
	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("Failed to run Gin server: %v", err)
		}
	}()
	log.Println("Gin server started on :8080")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down gracefully...")
	closeFn() // Close MQTT connection
	fmt.Println("Server shut down.")
}

// Old_main can be kept for testing or removed
func Old_main() {
	brokerURL := "tls://a1edew9tp1yb1x-ats.iot.us-east-1.amazonaws.com:8883"
	clientID := "server"
	topic := "$aws/things/akshat_cc3200board/shadow/update/accepted"
	certPath := "certs/certificate.pem.crt"
	keyPath := "certs/private.pem.key"
	caPath := "certs/AmazonRootCA1.pem"

	msgChan, errChan, closeFn, err := mqttsubscriber.SubscribeToShadowUpdates(brokerURL, clientID, topic, certPath, keyPath, caPath)
	if err != nil {
		log.Fatalf("Failed to subscribe to shadow updates: %v", err)
	}
	defer closeFn()

	fmt.Println("Listening for new events. Press Ctrl+C to stop.")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to listen for messages and errors
	go func() {
		for {
			select {
			case msgPayload, ok := <-msgChan:
				if !ok {
					log.Println("Message channel closed.")
					return // Exit goroutine if channel is closed
				}
				fmt.Printf("Received shadow update: %s\n", string(msgPayload))
				// Process the message payload here
			case err, ok := <-errChan:
				if !ok {
					log.Println("Error channel closed.")
					return // Exit goroutine if channel is closed
				}
				log.Printf("Error from MQTT subscriber: %v", err)
				// Potentially attempt to reconnect or handle the error
				return // Exit goroutine on error for now
			}
		}
	}()

	// Wait for signal
	<-sigChan

	fmt.Println("Shutting down gracefully...")
	// closeFn() is called by defer
}

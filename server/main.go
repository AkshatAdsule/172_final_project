package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"b3/server/api" // Added for API handlers
	"b3/server/config"
	"b3/server/database"
	"b3/server/models"
	"b3/server/mqttsubscriber"
	"b3/server/ride"
	"b3/server/util"
	"b3/server/ws"

	"github.com/gin-gonic/gin"
)

// Moved the correct ShadowStateDesired struct definition here
type ShadowStateDesired struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	SpeedKnots float64 `json:"speed_knots"`
	Timestamp  string  `json:"timestamp"` // "HHMMSS.SS"
	ValidFix   bool    `json:"valid_fix"`
}

// ShadowState holds the overall state from the shadow document.
type ShadowState struct {
	Desired  ShadowStateDesired     `json:"desired"` // Now correctly refers to the struct above
	Reported map[string]interface{} `json:"reported"`
}

// ShadowDocument is the top-level structure of the AWS IoT device shadow.
type ShadowDocument struct {
	State     ShadowState            `json:"state"`
	Metadata  map[string]interface{} `json:"metadata"`
	Version   int                    `json:"version"`
	Timestamp int64                  `json:"timestamp"` // Unix epoch for the document
}

func main() {
	// Load configuration from file if it exists, otherwise use defaults
	appConfig := config.Get()
	if _, err := os.Stat("config.json"); err == nil {
		loadedConfig, err := config.LoadConfigFromFile("config.json")
		if err != nil {
			log.Printf("Warning: Failed to load config.json, using defaults: %v", err)
		} else {
			appConfig = loadedConfig
			config.SetConfig(appConfig) // Update global config
			log.Println("Configuration loaded from config.json")
		}
	} else {
		log.Println("config.json not found, using default configuration")
	}

	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	wsHub := ws.NewHub()
	go wsHub.Run()
	log.Println("WebSocket Hub initialized and running.")

	rideManager := ride.NewRideManager(db, appConfig, wsHub)
	inactivityCheckInterval := time.Duration(appConfig.RideEndStaticSecs) * time.Second
	if inactivityCheckInterval <= 0 {
		inactivityCheckInterval = 30 * time.Second
	}
	go rideManager.CheckInactivityLoop(inactivityCheckInterval)
	log.Println("RideManager initialized and inactivity checker started.")

	msgChan, errChan, closeFn, err := mqttsubscriber.SubscribeToShadowUpdates(
		appConfig.MQTTBrokerURL,
		appConfig.MQTTClientID,
		appConfig.MQTTTopic,
		appConfig.MQTTCertPath,
		appConfig.MQTTKeyPath,
		appConfig.MQTTRootCAPath,
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to shadow updates: %v", err)
	}

	go handleMqttMessageProcessing(msgChan, errChan, rideManager, appConfig)
	fmt.Println("MQTT Listener and processor started. Press Ctrl+C to stop server.")

	router := gin.Default()

	// Register API Handlers under /api group
	apiGroup := router.Group("/api")
	api.RegisterRideHandlers(apiGroup, db)

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/ws", func(ctx *gin.Context) {
		ws.ServeWs(wsHub, ctx.Writer, ctx.Request)
	})

	go func() {
		log.Printf("Starting Gin server on %s", appConfig.ServerAddress)
		if err := router.Run(appConfig.ServerAddress); err != nil {
			log.Fatalf("Failed to run Gin server: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down gracefully...")
	closeFn()
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

func handleMqttMessageProcessing(msgChan <-chan []byte, errChan <-chan error, rideManager *ride.RideManager, appCfg config.Config) {
	// This function now only processes MQTT messages and passes them to RideManager
	// WebSocket connection handling is done by ws.ServeWs and the ws.Hub
	go func() {
		for {
			select {
			case message, ok := <-msgChan:
				if !ok {
					log.Println("MQTT message channel closed.")
					return
				}
				log.Printf("Received raw MQTT message for processing: %s", string(message))

				var shadowDoc ShadowDocument
				if err := json.Unmarshal(message, &shadowDoc); err != nil {
					log.Printf("Error unmarshalling shadow document: %v.", err)
					continue
				}

				if shadowDoc.State.Desired.Timestamp == "" || !shadowDoc.State.Desired.ValidFix {
					log.Printf("No 'desired' state, timestamp is empty, or fix is not valid. Desired: %+v", shadowDoc.State.Desired)
					continue
				}

				docTimestamp := time.Unix(shadowDoc.Timestamp, 0).UTC()
				eventTime, err := util.CombineDateTime(docTimestamp, shadowDoc.State.Desired.Timestamp)
				if err != nil {
					log.Printf("Error combining date and time: %v. Using document timestamp as fallback.", err)
					eventTime = docTimestamp
				}

				currentPosition := models.Position{
					Latitude:   shadowDoc.State.Desired.Latitude,
					Longitude:  shadowDoc.State.Desired.Longitude,
					SpeedKnots: shadowDoc.State.Desired.SpeedKnots,
					Timestamp:  eventTime,
				}
				rideManager.HandleGPSData(currentPosition)

			case err, ok := <-errChan:
				if !ok {
					log.Println("MQTT error channel closed.")
					return
				}
				log.Printf("Error from MQTT subscriber: %v.", err)
				// Consider how to propagate this error if needed, for now, just log
				return // Or attempt to re-establish subscription? For now, exits processor.
			}
		}
	}()
}

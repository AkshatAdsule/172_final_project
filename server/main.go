package main

import (
	"context"
	"encoding/json"
	"flag"
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
	"b3/server/snsnotifier"
	"b3/server/util"
	"b3/server/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var testMode = flag.Bool("testmode", false, "Enable test mode to mock MQTT and manually send updates")

// Moved the correct ShadowStateDesired struct definition here
type ShadowStateDesired struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	SpeedKnots float64 `json:"speed_knots"`
	Timestamp  string  `json:"timestamp"` // "HHMMSS.SS"
	ValidFix   bool    `json:"valid_fix"`
	Status     string  `json:"status,omitempty"`      // For crash detection
	LockStatus string  `json:"lock_status,omitempty"` // For lock mode: "LOCKED" or "UNLOCKED"
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
	flag.Parse()

	// Load configuration from file first
	appConfig, err := config.LoadConfigFromFile("config.json")
	if err != nil {
		log.Printf("Warning: Failed to load config.json, using default configuration: %v", err)
		appConfig = config.AppConfig // Fallback to defaults
	} else {
		config.AppConfig = appConfig // Set global AppConfig
	}

	// Override test mode from flag
	if *testMode {
		appConfig.TestMode = true
		log.Println("Test mode enabled via command-line flag.")
	}

	// Initialize database
	db, err := database.NewStore(appConfig.PostgresConnStr)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Setup WebSocket Hub
	wsHub := ws.NewHub()
	go wsHub.Run()

	// Initialize RideManager
	rideManager := ride.NewRideManager(db, appConfig, wsHub)
	inactivityCheckInterval := time.Duration(appConfig.RideEndStaticSecs) * time.Second
	if inactivityCheckInterval <= 0 {
		inactivityCheckInterval = 30 * time.Second
	}
	go rideManager.CheckInactivityLoop(inactivityCheckInterval)
	log.Println("RideManager initialized and inactivity checker started.")

	// Initialize Notifier only if SNS is enabled in config
	var crashNotifier *snsnotifier.Notifier
	if appConfig.SNSEnabled {
		var err error
		crashNotifier, err = snsnotifier.NewNotifier(context.Background())
		if err != nil {
			log.Printf("Failed to initialize SNS Notifier: %v. Continuing without crash notifications.", err)
			crashNotifier = nil // Ensure notifier is nil on failure
		}
	}

	// Set up theft alert function after SNS notifier is initialized
	theftAlertFunc := func(lat, lon float64, timestamp time.Time) {
		if crashNotifier != nil {
			theftMessage := fmt.Sprintf(
				"🚨 THEFT ALERT 🚨\n\nUnauthorized movement detected while bike is locked at %s.\nLocation: lat %f, lon %f.\n\nGoogle Maps: https://maps.google.com/?q=%f,%f",
				timestamp.Format(time.RFC1123),
				lat,
				lon,
				lat,
				lon,
			)

			err := crashNotifier.PublishSimple(appConfig.SNSTopicArn, theftMessage)
			if err != nil {
				log.Printf("Failed to publish theft alert to SNS: %v", err)
			} else {
				log.Println("Successfully published theft alert to SNS.")
			}
		} else {
			log.Println("Theft detected, but SNS notifier is not enabled.")
		}
	}
	rideManager.SetTheftAlertFunc(theftAlertFunc)

	var msgChan <-chan []byte
	var errChan <-chan error
	var closeFn func()

	if appConfig.TestMode {
		log.Println("Running in test mode. MQTT subscriber is mocked.")
		msgChan, errChan, closeFn = mqttsubscriber.SubscribeToShadowUpdatesMock()
	} else {
		var err error
		msgChan, errChan, closeFn, err = mqttsubscriber.SubscribeToShadowUpdates(
			appConfig.MQTTBrokerURL,
			appConfig.MQTTClientID,
			appConfig.MQTTTopic,
			appConfig.MQTTCertPEM,
			appConfig.MQTTKeyPEM,
			appConfig.MQTTRootCAPEM,
			appConfig.MQTTCertPath,
			appConfig.MQTTKeyPath,
			appConfig.MQTTRootCAPath,
		)
		if err != nil {
			log.Fatalf("Failed to subscribe to shadow updates: %v", err)
		}
	}

	// Initialize MQTT Publisher for shadow updates
	var mqttPublisher *mqttsubscriber.Publisher
	if !appConfig.TestMode {
		var err error
		mqttPublisher, err = mqttsubscriber.NewPublisher(
			appConfig.MQTTBrokerURL,
			appConfig.MQTTClientID,
			appConfig.MQTTUpdateTopic,
			appConfig.MQTTCertPEM,
			appConfig.MQTTKeyPEM,
			appConfig.MQTTRootCAPEM,
			appConfig.MQTTCertPath,
			appConfig.MQTTKeyPath,
			appConfig.MQTTRootCAPath,
		)
		if err != nil {
			log.Printf("Failed to initialize MQTT publisher: %v. Lock status updates will not be published.", err)
			mqttPublisher = nil
		} else {
			log.Println("MQTT Publisher initialized successfully.")
		}
	}

	go handleMqttMessageProcessing(msgChan, errChan, rideManager, appConfig, crashNotifier)
	fmt.Println("MQTT Listener and processor started. Press Ctrl+C to stop server.")

	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Configure CORS
	corsConfig := cors.Config{
		AllowOrigins: []string{
			"https://b3.aksads.tech",
			"http://b3.aksads.tech",
			"http://localhost:5173",
			"https://localhost:5173",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsConfig))

	// Register API Handlers
	apiGroup := router.Group("/api")
	api.RegisterRideHandlers(apiGroup, db)
	api.RegisterLockHandlers(apiGroup, rideManager, mqttPublisher)

	// Add test-only endpoints if in test mode
	if appConfig.TestMode {
		testGroup := router.Group("/api/test")
		{
			testGroup.POST("/location_update", func(ctx *gin.Context) {
				var shadowDoc ShadowDocument
				if err := ctx.ShouldBindJSON(&shadowDoc); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if shadowDoc.State.Desired.Timestamp == "" {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": "state.desired.timestamp is required"})
					return
				}

				payload, err := json.Marshal(shadowDoc)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal shadow document"})
					return
				}

				err = mqttsubscriber.PublishMockMessage(payload)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				ctx.JSON(http.StatusOK, gin.H{"message": "Location update sent to mock channel"})
			})
		}
	}

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
	if crashNotifier != nil {
		crashNotifier.Close()
	}
	if mqttPublisher != nil {
		mqttPublisher.Close()
	}
	fmt.Println("Server shut down.")
}

func handleMqttMessageProcessing(msgChan <-chan []byte, errChan <-chan error, rideManager *ride.RideManager, appCfg config.Config, crashNotifier *snsnotifier.Notifier) {
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

				// Check for lock status updates
				if shadowDoc.State.Desired.LockStatus != "" {
					log.Printf("Lock status update received: %s", shadowDoc.State.Desired.LockStatus)
					rideManager.SetLockStatus(shadowDoc.State.Desired.LockStatus)
				}

				// Check for crash detection
				if shadowDoc.State.Desired.Status == "CRASH_DETECTED" {
					if crashNotifier != nil {
						// Constructing the message for SNS
						crashMessage := fmt.Sprintf(
							"🚨 CRASH DETECTED 🚨\n\nCrash detected for device at %s.\nLast known location: lat %f, lon %f.\n\nGoogle Maps: https://maps.google.com/?q=%f,%f",
							time.Now().Format(time.RFC1123),
							shadowDoc.State.Desired.Latitude,
							shadowDoc.State.Desired.Longitude,
							shadowDoc.State.Desired.Latitude,
							shadowDoc.State.Desired.Longitude,
						)

						// Use PublishSimple for standard SNS topics (not FIFO)
						err := crashNotifier.PublishSimple(appCfg.SNSTopicArn, crashMessage)
						if err != nil {
							log.Printf("Failed to publish crash notification to SNS: %v", err)
						} else {
							log.Println("Successfully published crash notification to SNS.")
						}
					} else {
						log.Println("Crash detected, but SNS notifier is not enabled.")
					}
					continue // Don't process this as a regular GPS point for ride tracking
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
				return
			}
		}
	}()
}

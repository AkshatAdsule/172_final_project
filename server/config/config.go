package config

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	MQTTBrokerURL     string         `json:"mqtt_broker_url"`
	MQTTClientID      string         `json:"mqtt_client_id"`
	MQTTTopic         string         `json:"mqtt_topic"`
	MQTTCertPath      string         `json:"mqtt_cert_path"`
	MQTTKeyPath       string         `json:"mqtt_key_path"`
	MQTTRootCAPath    string         `json:"mqtt_root_ca_path"`
	MQTTCertPEM       string         `json:"-"`                           // Loaded from env, not json
	MQTTKeyPEM        string         `json:"-"`                           // Loaded from env, not json
	MQTTRootCAPEM     string         `json:"-"`                           // Loaded from env, not json
	PostgresConnStr   string         `json:"-"`                           // Loaded from env, not json
	ServerAddress     string         `json:"server_address"`              // e.g., ":8080"
	RideStartDistance float64        `json:"ride_start_distance_meters"`  // meters
	RideEndInactivity int            `json:"ride_end_inactivity_seconds"` // seconds
	RideEndStaticSecs int            `json:"ride_end_static_seconds"`     // seconds
	RideEndStaticDist float64        `json:"ride_end_static_dist_meters"` // meters
	Timezone          string         `json:"timezone"`                    // e.g., "America/Los_Angeles" for ride naming
	PSTLocation       *time.Location // Loaded based on Timezone or default to PST

	// SNS Configuration
	SNSTopicArn string `json:"sns_topic_arn,omitempty"` // Default SNS topic ARN for notifications
	SNSRegion   string `json:"sns_region,omitempty"`    // AWS region for SNS (optional, uses default AWS config if empty)
	SNSEnabled  bool   `json:"sns_enabled"`             // Whether SNS notifications are enabled
}

var defaultConfig = Config{
	MQTTBrokerURL:     "tls://a1edew9tp1yb1x-ats.iot.us-east-1.amazonaws.com:8883",
	MQTTClientID:      "server-ride-tracker",
	MQTTTopic:         "$aws/things/akshat_cc3200board/shadow/update/accepted",
	MQTTCertPath:      "certs/certificate.pem.crt", // Relative to executable or defined base path
	MQTTKeyPath:       "certs/private.pem.key",     // Relative
	MQTTRootCAPath:    "certs/AmazonRootCA1.pem",   // Relative
	ServerAddress:     ":8080",
	RideStartDistance: 8.0,                   // meters
	RideEndInactivity: 120,                   // seconds (2 minutes)
	RideEndStaticSecs: 120,                   // seconds (2 minutes)
	RideEndStaticDist: 8.0,                   // meters
	Timezone:          "America/Los_Angeles", // Default to PST as discussed

	// SNS defaults
	SNSTopicArn: "",    // To be set via config file or environment variable
	SNSRegion:   "",    // Uses default AWS config region if empty
	SNSEnabled:  false, // Disabled by default
}

// AppConfig is the global configuration instance.
var AppConfig Config

func init() {
	// Load .env file first to make variables available for the rest of init
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env file, relying on manually set environment variables or defaults")
	}

	// Load defaults first
	AppConfig = defaultConfig

	// Override with environment variables for PEM content if available
	if certPEM := os.Getenv("MQTT_CERT_PEM"); certPEM != "" {
		// decode PEM from base64
		decodedCertPEM, err := base64.StdEncoding.DecodeString(certPEM)
		if err != nil {
			log.Printf("Error decoding MQTT_CERT_PEM: %v", err)
			AppConfig.MQTTCertPEM = "" // Ensure it's empty on error
		} else {
			AppConfig.MQTTCertPEM = string(decodedCertPEM)
			log.Println("MQTT_CERT_PEM loaded from environment")
		}
	}
	if keyPEM := os.Getenv("MQTT_KEY_PEM"); keyPEM != "" {
		decodedKeyPEM, err := base64.StdEncoding.DecodeString(keyPEM)
		if err != nil {
			log.Printf("Error decoding MQTT_KEY_PEM: %v", err)
			AppConfig.MQTTKeyPEM = "" // Ensure it's empty on error
		} else {
			AppConfig.MQTTKeyPEM = string(decodedKeyPEM)
			log.Println("MQTT_KEY_PEM loaded from environment")
		}
	}
	if caPEM := os.Getenv("MQTT_ROOTCA_PEM"); caPEM != "" {
		decodedCaPEM, err := base64.StdEncoding.DecodeString(caPEM)
		if err != nil {
			log.Printf("Error decoding MQTT_ROOTCA_PEM: %v", err)
			AppConfig.MQTTRootCAPEM = "" // Ensure it's empty on error
		} else {
			AppConfig.MQTTRootCAPEM = string(decodedCaPEM)
			log.Println("MQTT_ROOTCA_PEM loaded from environment")
		}
	}

	// Override ServerAddress with PORT environment variable if available
	if port := os.Getenv("PORT"); port != "" {
		AppConfig.ServerAddress = ":" + port
		log.Printf("Server address set to %s from PORT environment variable", AppConfig.ServerAddress)
	}

	// Load PostgreSQL connection string from environment if available
	if postgresConn := os.Getenv("POSTGRES_CONNECTION_STRING"); postgresConn != "" {
		AppConfig.PostgresConnStr = postgresConn
		log.Println("PostgreSQL connection string loaded from environment")
	}

	// For now, load defaults. Later, we can load from a file or env vars.
	// AppConfig = defaultConfig // This line is now redundant due to above assignments

	// Ensure PSTLocation is loaded
	loc, err := time.LoadLocation(AppConfig.Timezone)
	if err != nil {
		// Fallback to fixed offset if location loading fails (e.g. in minimal containers)
		// This matches the PST offset used in util/timeutils.go.
		// log.Printf("Warning: Could not load timezone %s: %v. Falling back to fixed PST offset.", AppConfig.Timezone, err)
		AppConfig.PSTLocation = time.FixedZone("PST-Fallback", -7*60*60)
	} else {
		AppConfig.PSTLocation = loc
	}
}

// LoadConfigFromFile would load configuration from a JSON file.
// This is a placeholder for future implementation.
func LoadConfigFromFile(filePath string) (Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	// Preserve PEM values loaded from environment if they exist,
	// as they are not in config.json
	if AppConfig.MQTTCertPEM != "" {
		cfg.MQTTCertPEM = AppConfig.MQTTCertPEM
	}
	if AppConfig.MQTTKeyPEM != "" {
		cfg.MQTTKeyPEM = AppConfig.MQTTKeyPEM
	}
	if AppConfig.MQTTRootCAPEM != "" {
		cfg.MQTTRootCAPEM = AppConfig.MQTTRootCAPEM
	}
	if AppConfig.ServerAddress != "" {
		cfg.ServerAddress = AppConfig.ServerAddress
	}
	// Preserve PostgreSQL connection string from environment
	if AppConfig.PostgresConnStr != "" {
		cfg.PostgresConnStr = AppConfig.PostgresConnStr
	}

	// Ensure PSTLocation is loaded after reading from file
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		cfg.PSTLocation = time.FixedZone("PST-Fallback", -7*60*60)
	} else {
		cfg.PSTLocation = loc
	}
	return cfg, nil
}

// Get returns the global application configuration.
func Get() Config {
	return AppConfig
}

// SetConfig updates the global application configuration.
func SetConfig(cfg Config) {
	AppConfig = cfg
}

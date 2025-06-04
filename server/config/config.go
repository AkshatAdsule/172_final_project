package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	MQTTBrokerURL     string         `json:"mqtt_broker_url"`
	MQTTClientID      string         `json:"mqtt_client_id"`
	MQTTTopic         string         `json:"mqtt_topic"`
	MQTTCertPath      string         `json:"mqtt_cert_path"`
	MQTTKeyPath       string         `json:"mqtt_key_path"`
	MQTTRootCAPath    string         `json:"mqtt_root_ca_path"`
	DatabasePath      string         `json:"database_path"`               // e.g., "data/rides.db"
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
	MQTTCertPath:      "certs/certificate.pem.crt",       // Relative to executable or defined base path
	MQTTKeyPath:       "certs/private.pem.key",           // Relative
	MQTTRootCAPath:    "certs/AmazonRootCA1.pem",         // Relative
	DatabasePath:      filepath.Join("data", "rides.db"), // Store in a 'data' subdirectory
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
	// For now, load defaults. Later, we can load from a file or env vars.
	AppConfig = defaultConfig

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

	// Ensure the directory for the database exists
	dbDir := filepath.Dir(AppConfig.DatabasePath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if mkDirErr := os.MkdirAll(dbDir, 0755); mkDirErr != nil {
			// log.Fatalf("Failed to create data directory %s: %v", dbDir, mkDirErr)
			// For now, we'll let it potentially fail later if DB can't be created.
			// In a real app, you'd want to handle this more robustly.
		}
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

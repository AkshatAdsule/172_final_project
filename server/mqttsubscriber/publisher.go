package mqttsubscriber

import (
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Publisher handles MQTT publishing operations
type Publisher struct {
	client      mqtt.Client
	updateTopic string // e.g., "$aws/things/device_name/shadow/update"
}

// ShadowUpdatePayload represents the structure for shadow update messages
type ShadowUpdatePayload struct {
	State ShadowUpdateState `json:"state"`
}

type ShadowUpdateState struct {
	Desired map[string]interface{} `json:"desired"`
}

// NewPublisher creates a new MQTT publisher using existing connection parameters
func NewPublisher(brokerURL, clientID, updateTopic string, cfgMqttCertPEM, cfgMqttKeyPEM, cfgMqttRootCAPEM, cfgMqttCertPath, cfgMqttKeyPath, cfgMqttRootCAPath string) (*Publisher, error) {
	tlsConfig, err := NewTLSConfig(cfgMqttRootCAPEM, cfgMqttCertPEM, cfgMqttKeyPEM, cfgMqttRootCAPath, cfgMqttCertPath, cfgMqttKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID + "-publisher")
	opts.SetTLSConfig(tlsConfig)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect MQTT publisher: %w", token.Error())
	}

	return &Publisher{
		client:      client,
		updateTopic: updateTopic,
	}, nil
}

// UpdateLockStatus publishes a lock status update to the shadow
func (p *Publisher) UpdateLockStatus(lockStatus string) error {
	payload := ShadowUpdatePayload{
		State: ShadowUpdateState{
			Desired: map[string]interface{}{
				"lock_status": lockStatus,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal shadow update payload: %w", err)
	}

	token := p.client.Publish(p.updateTopic, 0, false, payloadBytes)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish lock status update: %w", token.Error())
	}

	log.Printf("Successfully published lock status update: %s", lockStatus)
	return nil
}

// Close disconnects the publisher
func (p *Publisher) Close() {
	if p.client != nil && p.client.IsConnected() {
		p.client.Disconnect(250)
		log.Println("MQTT Publisher disconnected.")
	}
}

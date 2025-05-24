package mqttsubscriber

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// NewTLSConfig sets up the TLS configuration for MQTT.
// It expects certificate files to be in a 'certs' directory relative to the execution path.
func NewTLSConfig(certPath, keyPath, caPath string) (*tls.Config, error) {
	certpool := x509.NewCertPool()
	pemCerts, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}
	certpool.AppendCertsFromPEM(pemCerts)

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client key pair: %w", err)
	}

	return &tls.Config{
		RootCAs:      certpool,
		ClientAuth:   tls.NoClientCert,
		ClientCAs:    nil,
		Certificates: []tls.Certificate{cert},
	}, nil
}

// SubscribeToShadowUpdates connects to the MQTT broker, subscribes to the AWS IoT shadow updates,
// and returns a channel that will receive message payloads.
// It also returns a channel for errors and a function to gracefully close the connection.
func SubscribeToShadowUpdates(brokerURL, clientID, topic, certPath, keyPath, caPath string) (<-chan []byte, <-chan error, func(), error) {
	tlsConfig, err := NewTLSConfig(certPath, keyPath, caPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create TLS config: %w", err)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)
	opts.SetTLSConfig(tlsConfig)

	messageChan := make(chan []byte)
	errorChan := make(chan error, 1) // Buffered error channel

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
		errorChan <- fmt.Errorf("connection lost: %w", err)
		close(messageChan) // Close message channel on connection loss
	})

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Println("Connected to MQTT broker.")
		if token := c.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
			log.Printf("Received message on topic %s", msg.Topic())
			// Send a copy of the payload to avoid issues if the underlying buffer is reused
			payloadCopy := make([]byte, len(msg.Payload()))
			copy(payloadCopy, msg.Payload())
			messageChan <- payloadCopy
		}); token.Wait() && token.Error() != nil {
			log.Printf("Failed to subscribe to topic %s: %v", topic, token.Error())
			errorChan <- fmt.Errorf("failed to subscribe: %w", token.Error())
			close(messageChan) // Close message channel on subscription failure
		} else {
			log.Printf("Successfully subscribed to topic: %s", topic)
		}
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	closeFn := func() {
		log.Println("Disconnecting MQTT client...")
		client.Unsubscribe(topic)
		client.Disconnect(250)
		close(messageChan)
		close(errorChan)
		log.Println("MQTT client disconnected.")
	}

	return messageChan, errorChan, closeFn, nil
}

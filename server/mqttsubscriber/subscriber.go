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
// It tries to load certs from PEM strings first, then falls back to file paths.
func NewTLSConfig(caPEM, certPEM, keyPEM, caPath, certPath, keyPath string) (*tls.Config, error) {
	certpool := x509.NewCertPool()

	if caPEM != "" {
		if !certpool.AppendCertsFromPEM([]byte(caPEM)) {
			return nil, fmt.Errorf("failed to append CA certificate from PEM string")
		}
	} else if caPath != "" {
		pemCerts, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate from path %s: %w", caPath, err)
		}
		if !certpool.AppendCertsFromPEM(pemCerts) {
			return nil, fmt.Errorf("failed to append CA certificate from file %s", caPath)
		}
	} else {
		log.Println("Warning: No CA certificate PEM string or file path provided. System CAs will be used if available, or connection may be insecure.")
		// If no CA is provided, it will use system CAs or might be insecure depending on server/client config.
		// For some brokers, not providing a RootCA means it won't be explicitly trusted, relying on public CAs.
	}

	var clientCert tls.Certificate
	var err error

	if certPEM != "" && keyPEM != "" {
		clientCert, err = tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
		if err != nil {
			return nil, fmt.Errorf("failed to load client key pair from PEM strings: %w", err)
		}
	} else if certPath != "" && keyPath != "" {
		clientCert, err = tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client key pair from paths %s, %s: %w", certPath, keyPath, err)
		}
	} else {
		log.Println("Warning: No client certificate PEMs or file paths provided. Proceeding without client certificate.")
		// No client certs to add if neither PEMs nor paths are provided
		return &tls.Config{
			RootCAs:    certpool,
			ClientAuth: tls.NoClientCert,
			ClientCAs:  nil,
		}, nil
	}

	return &tls.Config{
		RootCAs:      certpool,
		ClientAuth:   tls.NoClientCert, // Assuming server does not require client cert, or client cert is optional
		ClientCAs:    nil,
		Certificates: []tls.Certificate{clientCert},
	}, nil
}

// SubscribeToShadowUpdates connects to the MQTT broker, subscribes to the AWS IoT shadow updates,
// and returns a channel that will receive message payloads.
// It also returns a channel for errors and a function to gracefully close the connection.
func SubscribeToShadowUpdates(brokerURL, clientID, topic string, cfgMqttCertPEM, cfgMqttKeyPEM, cfgMqttRootCAPEM, cfgMqttCertPath, cfgMqttKeyPath, cfgMqttRootCAPath string) (<-chan []byte, <-chan error, func(), error) {
	tlsConfig, err := NewTLSConfig(cfgMqttRootCAPEM, cfgMqttCertPEM, cfgMqttKeyPEM, cfgMqttRootCAPath, cfgMqttCertPath, cfgMqttKeyPath)
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

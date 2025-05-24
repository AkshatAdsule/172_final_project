# AWS IoT Shadow MQTT to WebSocket Bridge

A Go-based server that bridges AWS IoT device shadow updates from MQTT to WebSocket connections, enabling real-time communication between IoT devices and web applications.

## Overview

This project provides a real-time bridge between AWS IoT Core device shadow updates and web clients. It subscribes to AWS IoT shadow update messages via MQTT and forwards the `desired` state to connected WebSocket clients. This is particularly useful for building web dashboards or control interfaces for IoT devices.

## Architecture

```
AWS IoT Device → AWS IoT Core → MQTT (TLS) → Go Server → WebSocket → Web Client
```

The server:
1. Connects to AWS IoT Core using MQTT over TLS with X.509 certificates
2. Subscribes to device shadow update topics
3. Parses shadow documents and extracts the `desired` state
4. Forwards the desired state to connected WebSocket clients
5. Provides a REST API endpoint for health checks

## Features

- **Secure MQTT Connection**: Uses TLS with X.509 certificates for AWS IoT Core
- **Real-time WebSocket**: Bi-directional WebSocket communication
- **Shadow State Parsing**: Automatically extracts and forwards the `desired` state from shadow documents
- **Error Handling**: Robust error handling with graceful connection management
- **Health Check**: REST endpoint for monitoring server status
- **Graceful Shutdown**: Proper cleanup on termination signals

## Prerequisites

- Go 1.24.3 or higher
- AWS IoT Core setup with:
  - A registered IoT device (e.g., `akshat_cc3200board`)
  - X.509 certificates for device authentication
  - Proper IAM policies for shadow access

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd server2
```

2. Install dependencies:
```bash
go mod download
```

3. Set up your AWS IoT certificates in the `certs/` directory:
   - `certificate.pem.crt` - Device certificate
   - `private.pem.key` - Private key
   - `public.pem.key` - Public key
   - `AmazonRootCA1.pem` - Amazon Root CA certificate

## Configuration

The server is configured for a specific AWS IoT setup. Update the following variables in `main.go` if needed:

```go
brokerURL := "tls://a1edew9tp1yb1x-ats.iot.us-east-1.amazonaws.com:8883"
clientID := "server-main"
topic := "$aws/things/akshat_cc3200board/shadow/update/accepted"
```

## Usage

### Running the Server

```bash
go run main.go
```

The server will:
- Start on port 8080
- Connect to AWS IoT Core via MQTT
- Accept WebSocket connections at `/ws`
- Provide health check at `/ping`

### API Endpoints

#### Health Check
```
GET /ping
```
Returns: `{"message": "pong"}`

#### WebSocket Connection
```
WebSocket: ws://localhost:8080/ws
```
- Connects to receive real-time shadow updates
- Receives JSON messages containing the `desired` state from device shadows

### Example WebSocket Client

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = function(event) {
    console.log('Connected to server');
};

ws.onmessage = function(event) {
    const desiredState = JSON.parse(event.data);
    console.log('Received desired state:', desiredState);
    // Handle the desired state update
};

ws.onclose = function(event) {
    console.log('Disconnected from server');
};
```

## Project Structure

```
server2/
├── main.go                 # Main server implementation
├── go.mod                  # Go module dependencies
├── go.sum                  # Go module checksums
├── README.md              # This file
├── certs/                 # AWS IoT certificates
│   ├── AmazonRootCA1.pem
│   ├── certificate.pem.crt
│   ├── private.pem.key
│   └── public.pem.key
└── mqttsubscriber/        # MQTT subscriber package
    └── subscriber.go      # MQTT connection and subscription logic
```

## Dependencies

- **Gin**: HTTP web framework for REST API and WebSocket upgrade
- **Gorilla WebSocket**: WebSocket implementation
- **Eclipse Paho MQTT**: MQTT client library
- Standard Go libraries for TLS, JSON, and signal handling

## Error Handling

The server includes comprehensive error handling:

- **MQTT Connection Errors**: Logged and propagated to WebSocket clients
- **Certificate Issues**: Server fails to start with descriptive error messages
- **WebSocket Disconnections**: Automatically detected and cleaned up
- **Malformed Messages**: Raw messages sent if JSON parsing fails
- **Graceful Shutdown**: SIGINT/SIGTERM signals trigger proper cleanup

## Development

### Testing MQTT Connection

The project includes an `Old_main()` function for testing MQTT connectivity without the WebSocket server:

```go
// Uncomment and call Old_main() instead of main() for testing
```

### Adding New Features

1. **Additional Shadow States**: Modify the `ShadowDocument` struct to handle `reported` states or metadata
2. **Multiple Topics**: Extend the subscriber to handle multiple shadow update topics
3. **Authentication**: Add WebSocket authentication if needed
4. **Logging**: Enhance logging with structured logging libraries

## Troubleshooting

### Common Issues

1. **Certificate Errors**: Ensure certificates are valid and paths are correct
2. **Connection Refused**: Verify AWS IoT endpoint URL and network connectivity
3. **Permission Denied**: Check IAM policies for the IoT device
4. **WebSocket Disconnections**: Monitor network stability and implement reconnection logic

### Logs

The server provides detailed logging for:
- MQTT connection status
- WebSocket client connections/disconnections
- Message processing and forwarding
- Error conditions

## License

This project is part of a school final project (Course 172).

## Contributing

This is an academic project. For improvements or bug fixes, please follow standard Go coding conventions and include appropriate tests.

# Real-Time GPS Ride Tracking Backend

> ℹ️ For overall project instructions, see the [root README](../README.md)

This Go-based server processes GPS updates from MQTT, tracks rides, stores them in an SQLite database, and exposes ride data via a RESTful API and live WebSocket events.

## 1. Overview

This backend system is designed to:
- Subscribe to GPS location updates from an MQTT topic (e.g., AWS IoT device shadow updates).
- Detect and define "rides" based on significant movement and periods of inactivity.
- Store detailed ride information, including all GPS positions, in an SQLite database.
- Provide a RESTful API to query for ride summaries and detailed ride data.
- Broadcast real-time ride events (start, end, new position) to connected WebSocket clients.

It replaces a simpler MQTT-to-WebSocket bridge functionality with a more comprehensive ride tracking and data management system.

## 2. Architecture

```
MQTT (e.g., AWS IoT GPS Data) → Go Server Backend:
                                 ├─ MQTT Subscriber
                                 ├─ Ride Manager (Stateful Ride Logic)
                                 ├─ SQLite Database (via database/store.go)
                                 ├─ WebSocket Hub (for live updates)
                                 └─ Gin HTTP Server:
                                     ├─ REST API (/api/rides, /api/rides/:id)
                                     └─ WebSocket Endpoint (/ws)
                                          ↓
                                    Web Clients (consuming API and WebSocket)
```

The server components:
1.  **MQTT Subscriber (`mqttsubscriber/`):** Connects to the MQTT broker (e.g., AWS IoT Core) using TLS with X.509 certificates and subscribes to GPS data topics.
2.  **Configuration (`config/`, `config.json`):** Manages application settings including MQTT credentials, database paths, and ride detection parameters.
3.  **Models (`models/`):** Defines data structures for `Position`, `RideSummary`, `RideDetail`, etc.
4.  **Utilities (`util/`):** Provides helper functions for tasks like Haversine distance calculation and time parsing.
5.  **Database Store (`database/store.go`):** Handles all interactions with the SQLite database, including table creation and CRUD operations for rides and positions.
6.  **Ride Service (`ride/service.go`):** Contains stateless logic for ride event determination (e.g., has a ride started/stopped based on new GPS point).
7.  **Ride Manager (`ride/manager.go`):** Stateful component that uses the Ride Service and Database Store to manage the lifecycle of a ride, process GPS points, and trigger events.
8.  **WebSocket Hub & Client (`ws/`):** Manages active WebSocket client connections and broadcasts structured ride event messages to all connected clients.
9.  **API Handlers (`api/handlers.go`):** Implements Gin handlers for the RESTful API endpoints.
10. **Main (`main.go`):** Initializes all components, sets up routing, and starts the server.

## 3. Features

- **Secure MQTT Connection**: Uses TLS with X.509 certificates (configurable for AWS IoT Core or other brokers).
- **Automatic Ride Detection**: Identifies rides based on configurable distance thresholds and inactivity periods.
- **Persistent Ride Storage**: Stores ride details and GPS tracks in an SQLite database.
- **RESTful API for Rides**: 
    - `GET /api/rides`: Retrieve a list of all ride summaries.
    - `GET /api/rides/:id`: Retrieve full details for a specific ride, including all GPS points.
- **Lock Mode & Theft Detection**: 
    - `POST /api/setLockStatus`: Set bike lock status (LOCKED/UNLOCKED).
    - `GET /api/getLockStatus`: Get current lock status.
    - When locked, movement detection triggers theft alerts via SNS instead of starting rides.
- **Real-time Ride Events via WebSocket**: Broadcasts structured JSON messages for:
    - `RIDE_STARTED`: When a new ride begins.
    - `RIDE_ENDED`: When a ride concludes.
    - `RIDE_POSITION_UPDATE`: When a new GPS point is added to an ongoing ride.
- **Alert Notifications**: SNS-based notifications for crash detection and theft alerts.
- **Configurable Parameters**: Ride detection logic (start distance, end inactivity, etc.) can be tuned via `config.json`.
- **Graceful Shutdown**: Proper cleanup of resources (DB connection, MQTT subscription) on termination signals.
- **Health Check**: `GET /ping` endpoint for basic server status.

## 4. Prerequisites

- Go 1.20 or higher.
- An MQTT broker publishing GPS data in the expected format (see `config.json` and `models.Position`).
- If using AWS IoT: an AWS IoT Core setup with a device, certificates, and IAM policies for shadow/topic access.

## 5. Installation

1.  Clone the repository:
    ```bash
    git clone <repository-url>
    cd <project-directory>/server
    ```

2.  Install dependencies:
    ```bash
    go mod tidy
    ```

3.  Build the server binary:
    ```bash
    go build -o server main.go
    ```

4.  **Crucial:** Set up your `certs/` directory if using TLS for MQTT (e.g., for AWS IoT):
    - `certificate.pem.crt` - Client/Device certificate
    - `private.pem.key` - Client/Device private key
    - `AmazonRootCA1.pem` (or your broker's CA) - Root CA certificate
    *Ensure the paths in `config.json` point to these files correctly.*

## 6. Configuration (`config.json`)

Create a `config.json` file in the root of the `server` directory. This file controls all operational parameters. **You must update this file with your specific settings.**

```json
{
  "mqtt_broker_url": "tls://your-iot-endpoint:8883",
  "mqtt_client_id": "your-server-client-id",
  "mqtt_topic": "$aws/things/your-thing-name/shadow/update/accepted",
  "mqtt_cert_path": "certs/certificate.pem.crt",
  "mqtt_key_path": "certs/private.pem.key",
  "mqtt_root_ca_path": "certs/AmazonRootCA1.pem",
  "database_path": "data/rides.db",
  "server_address": ":8080",
  "ride_start_distance_meters": 50.0,
  "ride_end_inactivity_seconds": 300,
  "ride_end_static_seconds": 120,
  "ride_end_static_dist_meters": 10.0,
  "timezone_offset_seconds": -25200, // Example: PST (UTC-7). Used for naming rides like "Morning Ride".
  "data_dir": "data" // Directory for database file, will be created if not exists.
}
```

**Key Configuration Fields:**
- `mqtt_broker_url`, `mqtt_client_id`, `mqtt_topic`: Your MQTT broker details.
- `mqtt_cert_path`, `mqtt_key_path`, `mqtt_root_ca_path`: Paths to your TLS certificates for MQTT.
- `database_path`: Path to the SQLite database file (e.g., `data/rides.db`). The `data_dir` will be created if it doesn't exist.
- `server_address`: Address and port for the HTTP server (e.g., `:8080`).
- `ride_start_distance_meters`: Minimum distance change to trigger a new ride.
- `ride_end_inactivity_seconds`: Time (seconds) of no GPS updates to automatically end a ride.
- `ride_end_static_seconds`: Time (seconds) a device can be static (not moving much) before ending a ride if paused.
- `ride_end_static_dist_meters`: Distance threshold (meters) below which a device is considered static/paused.
- `timezone_offset_seconds`: Used for determining ride names like "Morning Ride", "Evening Ride" based on local time.

## 7. Usage

### Running the Server

1.  Ensure `config.json` is correctly configured.
2.  Run the server:
    ```bash
    go run main.go
    # Or start using the built binary:
    ./server
    ```

The server will:
- Start the Gin HTTP server (default `:8080`).
- Connect to the MQTT broker.
- Initialize the SQLite database (creating the `data` directory and `rides.db` file if they don't exist).
- Accept WebSocket connections at `ws://<server_address>/ws`.
- Provide REST API endpoints under `/api`.

### API Endpoints

#### Health Check
- **`GET /ping`**
  - Returns: `{"message": "pong"}`

#### Rides API
- **`GET /api/rides`**
  - Description: Retrieves a list of all ride summaries.
  - Returns: `200 OK` with a JSON array of `RideSummary` objects.
    ```json
    [
      {
        "id": 1,
        "name": "Morning Ride",
        "start_time": "2023-10-27T10:00:00Z",
        "end_time": "2023-10-27T10:30:00Z"
      }
      // ... more rides
    ]
    ```
- **`GET /api/rides/:id`**
  - Description: Retrieves detailed information for a specific ride, including all its GPS positions.
  - URL Parameter: `id` (integer) - The ID of the ride.
  - Returns: `200 OK` with a JSON `RideDetail` object.
    ```json
    {
      "id": 1,
      "name": "Morning Ride",
      "start_time": "2023-10-27T10:00:00Z",
      "end_time": "2023-10-27T10:30:00Z",
      "positions": [
        {
          "latitude": 38.545288,
          "longitude": -121.739532,
          "speed_knots": 0.8,
          "timestamp": "2023-10-27T10:00:05Z"
        }
        // ... more positions
      ]
    }
    ```
  - Returns: `404 Not Found` if the ride ID does not exist.
  - Returns: `400 Bad Request` if the ID is not a valid integer.

#### Lock Mode API
- **`POST /api/setLockStatus`**
  - Description: Sets the bike's lock status and publishes the update to IoT Shadow.
  - Request Body: `{"status": "LOCKED"}` or `{"status": "UNLOCKED"}`
  - Returns: `200 OK` with the updated status.
  - Note: When locked, movement detection triggers theft alerts instead of starting rides.
- **`GET /api/getLockStatus`**
  - Description: Returns the current lock status.
  - Returns: `200 OK` with `{"status": "LOCKED"}` or `{"status": "UNLOCKED"}`

### WebSocket Events

- **Connection URL**: `ws://<server_address>/ws`
- **Messages**: JSON formatted messages indicating ride events.

**Common Message Structure:**
```json
{
  "type": "EVENT_TYPE_STRING",
  "payload": { ... event specific data ... },
  "timestamp": "YYYY-MM-DDTHH:MM:SSZ" // UTC timestamp of when the event was broadcast
}
```

**Event Types & Payloads:**

1.  **`RIDE_STARTED`**
    - Payload: `RideEventPayload` (see `ws/hub.go` for struct)
      ```json
      {
        "ride_id": 123,
        "ride_name": "Afternoon Ride",
        "timestamp": "2023-10-27T14:00:00Z", // Ride start time
        "position": { // Initial position that started the ride
          "latitude": 38.545,
          "longitude": -121.739,
          "speed_knots": 5.2,
          "timestamp": "2023-10-27T14:00:00Z"
        }
      }
      ```

2.  **`RIDE_ENDED`**
    - Payload: `RideEventPayload` (subset of fields)
      ```json
      {
        "ride_id": 123,
        "timestamp": "2023-10-27T14:45:10Z" // Ride end time
      }
      ```

3.  **`RIDE_POSITION_UPDATE`**
    - Payload: `RideEventPayload`
      ```json
      {
        "ride_id": 123,
        "timestamp": "2023-10-27T14:05:15Z", // Timestamp of this position update
        "position": {
          "latitude": 38.550,
          "longitude": -121.745,
          "speed_knots": 10.1,
          "timestamp": "2023-10-27T14:05:15Z"
        }
      }
      ```

**Example WebSocket Client (JavaScript):**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws'); // Adjust to your server address

ws.onopen = function(event) {
    console.log('Connected to WebSocket server');
};

ws.onmessage = function(event) {
    try {
        const message = JSON.parse(event.data);
        console.log('Received message:', message.type, message.payload);

        switch (message.type) {
            case 'RIDE_STARTED':
                // Handle ride started event
                // e.g., display new ride on a map
                break;
            case 'RIDE_ENDED':
                // Handle ride ended event
                // e.g., finalize ride display
                break;
            case 'RIDE_POSITION_UPDATE':
                // Handle position update for an ongoing ride
                // e.g., update marker on a map
                break;
            default:
                console.log('Received unknown message type:', message.type);
        }
    } catch (e) {
        console.error('Error parsing WebSocket message:', e, event.data);
    }
};

ws.onclose = function(event) {
    console.log('Disconnected from WebSocket server', event);
};

ws.onerror = function(error) {
    console.error('WebSocket Error:', error);
};
```

## 8. Project Structure

```
server/
├── main.go                 # Main server application
├── go.mod                  # Go module dependencies
├── go.sum                  # Go module checksums
├── README.md               # This file
├── config.json             # **User-created** configuration file
├── api/                    # API layer
│   └── handlers.go         # Gin handlers for REST API endpoints
├── certs/                  # (Example) Directory for MQTT TLS certificates
│   ├── certificate.pem.crt # (Example)
│   ├── private.pem.key     # (Example)
│   └── AmazonRootCA1.pem   # (Example)
├── config/                 # Configuration loading logic
│   └── config.go
├── database/               # Database interaction layer
│   └── store.go            # SQLite connection, table creation, CRUD operations
├── models/                 # Data structures (structs)
│   └── models.go
├── mqttsubscriber/         # MQTT subscriber package
│   └── subscriber.go
├── ride/                   # Ride detection and management logic
│   ├── manager.go          # Stateful ride management
│   └── service.go          # Stateless ride logic functions
├── util/                   # Utility functions
│   ├── geo.go              # Geolocation calculations (Haversine)
│   └── timeutils.go        # Time parsing and manipulation
└── ws/                     # WebSocket communication
    ├── client.go           # WebSocket client representation
    └── hub.go              # WebSocket hub for managing clients and broadcasting
```

## 9. Dependencies

- **Gin** (`github.com/gin-gonic/gin`): HTTP web framework.
- **Gorilla WebSocket** (`github.com/gorilla/websocket`): WebSocket implementation.
- **Eclipse Paho MQTT Go** (`github.com/eclipse/paho.mqtt.golang`): MQTT client library.
- **go-sqlite3** (`github.com/mattn/go-sqlite3`): SQLite driver.
- Standard Go libraries.

## 10. Development & Testing

- **MQTT Data:** Ensure your MQTT source is publishing GPS data. The `ShadowDocument` struct in `main.go` and `ShadowStateDesired` specifically expect `latitude`, `longitude`, `speed_knots`, `timestamp` (HHMMSS.SS string), and `valid_fix` within the `state.desired` part of the MQTT message. The top-level MQTT message should also have a `timestamp` field (Unix epoch seconds).
- **Database:** The database file specified in `config.json` (`database_path`) will be created automatically if it doesn't exist, along with the necessary tables.
- **Configuration:** Double-check all paths and parameters in `config.json`.
- **Logging:** The server provides logs for MQTT connections, ride processing, WebSocket events, and API requests.

## 11. Troubleshooting

- **Certificate Errors (MQTT):** Verify paths in `config.json` are correct and certificates are valid.
- **Database Issues:** Check permissions for the `data_dir` and `database_path` specified in `config.json`.
- **"No desired state" logs:** Your MQTT messages might not have the `state.desired.valid_fix: true` or `state.desired.timestamp` fields, or the `state.desired` object itself might be missing. Check the structure of your MQTT messages.

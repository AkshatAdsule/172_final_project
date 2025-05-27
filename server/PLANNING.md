# Backend Overhaul Plan: Ride Tracking and Storage

## 1. Overview

The goal is to refactor the existing backend to process MQTT GPS updates, identify and store "rides," and expose this ride data through a RESTful API. A ride starts when a significant GPS location change is detected and ends when GPS positions stabilize or stop updating. Rides will be stored in an SQLite database.

## 2. Core Components

### 2.1. MQTT Message Processing & Ride Detection
    - **Input Source:** AWS IoT shadow updates via MQTT.
    - **Relevant Data Structure (within `ShadowDocument.State.Desired`):**
      ```json
      {
        "latitude": 38.545288,
        "longitude": -121.739532,
        "speed_knots": 0.8,
        "timestamp": "191235.00", // HHMMSS.SS format (UTC)
        "valid_fix": true
      }
      ```
    - **Timestamp Handling:**
        - The primary timestamp for database records should be derived from the `ShadowDocument.Timestamp` field (Unix epoch `int64`), which represents the overall event time.
        - The `desired.timestamp` (HHMMSS.SS) can be used for fine-grained time if necessary, combined with the date from `ShadowDocument.Timestamp`.
    - **Pre-condition:** Only process updates where `valid_fix` is `true`.
    - **Ride Start Logic:**
        - Maintain the last known valid GPS position.
        - A new ride begins if the distance between the current and last valid position exceeds **8 meters** (approx. 25 feet).
        - Record the start time (using `ShadowDocument.Timestamp`).
    - **Ride End Logic:**
        - **Inactivity Timeout:** If no new valid GPS updates are received for **120 seconds** (2 minutes), the current ride ends.
        - **Static Position:** If GPS positions remain within an 8-meter radius (i.e., no "significant shift" detected) for **120 seconds**, the current ride ends.
        - Record the end time (using `ShadowDocument.Timestamp` of the last point or current time if timeout).
    - **Ride Naming (based on start time):**
        - Morning Ride: 06:00 - 11:59
        - Afternoon Ride: 12:00 - 17:59
        - Evening Ride: 18:00 - 21:59
        - Night Ride: 22:00 - 05:59 (local time of the server or a configured timezone)

### 2.2. Database Layer (SQLite)
    - **Database Schema:**
        - **`rides` Table:**
            - `id` (INTEGER, PRIMARY KEY, AUTOINCREMENT)
            - `name` (TEXT, NOT NULL) - e.g., "Morning Ride"
            - `start_time` (DATETIME, NOT NULL)  // ISO 8601 format, UTC
            - `end_time` (DATETIME)             // ISO 8601 format, UTC
        - **`ride_positions` Table:**
            - `id` (INTEGER, PRIMARY KEY, AUTOINCREMENT)
            - `ride_id` (INTEGER, NOT NULL, FOREIGN KEY references `rides(id)` ON DELETE CASCADE)
            - `latitude` (REAL, NOT NULL)
            - `longitude` (REAL, NOT NULL)
            - `speed_knots` (REAL)
            - `timestamp` (DATETIME, NOT NULL) // ISO 8601 format, UTC, from ShadowDocument.Timestamp
    - **Database Operations (Go functions/methods in `database` package):**
        - `InitializeDB(dataSourceName string) (*sql.DB, error)`
        - `CreateTables(db *sql.DB) error`
        - `CreateRide(db *sql.DB, name string, startTime time.Time) (rideID int64, err error)`
        - `AddPositionToRide(db *sql.DB, rideID int64, latitude float64, longitude float64, speedKnots float64, timestamp time.Time) error`
        - `EndRide(db *sql.DB, rideID int64, endTime time.Time) error`
        - `GetRideDetails(db *sql.DB, rideID int64) (*models.RideDetail, error)`
        - `GetAllRidesSummary(db *sql.DB) ([]models.RideSummary, error)`

### 2.3. RESTful API (using Gin)
    - **Base Path:** `/api`
    - **Endpoints:**
        - `GET /rides`:
            - Returns a list of all rides (summary view).
            - Query parameters: `?page=1&limit=10`, `?date=YYYY-MM-DD` (filters by start date).
            - Response: `[{ "id": 1, "name": "Morning Ride", "start_time": "...", "end_time": "..." }, ...]` (Uses `models.RideSummary`)
        - `GET /rides/{ride_id}`:
            - Returns details for a specific ride, including its GPS track.
            - Response: `{ "id": 1, "name": "Morning Ride", "start_time": "...", "end_time": "...", "positions": [{ "latitude": ..., "longitude": ..., "speed_knots": ..., "timestamp": "..." }, ...] }` (Uses `models.RideDetail`)
    - **Data Structures (Go structs in `models` package):**
        - `Position`
        - `RideSummary`
        - `RideDetail` (includes a slice of `Position`)

### 2.4. Configuration
    - **Source:** Environment variables or a `config.json` file.
    - **Parameters:**
        - MQTT Broker URL, Client ID, Topic, Cert Paths
        - SQLite database file path (e.g., `data/rides.db`)
        - GPS distance threshold for ride start (meters, e.g., `8.0`)
        - Inactivity timeout for ride end (seconds, e.g., `120`)
        - Static position radius for ride end (meters, e.g., `8.0`)
        - Timezone for ride naming (e.g., "America/Los_Angeles")

### 2.5. Real-time WebSocket Communication
    - **Endpoint:** `/ws` (existing, to be repurposed)
    - **Purpose:** Stream live GPS updates and ride status to connected clients.
    - **Message Types:**
        1.  **`current_location`**: Sent for every valid GPS update received from MQTT.
            - **Structure:** `{ "type": "current_location", "payload": { "latitude": float, "longitude": float, "timestamp": "ISO8601_string", "speed_knots": float } }`
        2.  **Ride Status Updates:**
            - **`ride_started`**: Sent when a new ride begins.
                - **Structure:** `{ "type": "ride_started", "payload": { "ride_id": int, "name": "string", "start_time": "ISO8601_string", "initial_position": { "latitude": float, "longitude": float, "timestamp": "ISO8601_string", "speed_knots": float } } }`
            - **`ride_position_added`**: Sent when a new GPS point is added to an ongoing ride.
                - **Structure:** `{ "type": "ride_position_added", "payload": { "ride_id": int, "position": { "latitude": float, "longitude": float, "timestamp": "ISO8601_string", "speed_knots": float } } }`
            - **`ride_ended`**: Sent when a ride concludes.
                - **Structure:** `{ "type": "ride_ended", "payload": { "ride_id": int, "end_time": "ISO8601_string" } }`
    - The WebSocket handler will need to manage client connections and broadcast these structured messages.

## 3. Implementation Steps & Structure

### 3.1. Project Structure (Potential additions/modifications)
    ```
    server/
    ├── api/                  // Gin handlers for REST API
    │   └── handlers.go
    ├── certs/
    ├── config/               // Configuration loading
    │   └── config.go
    ├── database/             // SQLite interaction logic
    │   └── store.go
    ├── models/               // Data structures (Ride, Position etc.)
    │   └── models.go
    ├── mqttsubscriber/       // Existing MQTT subscription logic
    ├── ride/                 // Ride detection, processing, state management
    │   └── service.go        // Core ride logic
    │   └── manager.go        // Manages active ride state, interacts with DB and websockets
    ├── util/                 // Utility functions (e.g. haversine distance, time parsing)
    │   └── geo.go
    │   └── timeutils.go
    ├── ws/                   // WebSocket handling
    │   └── hub.go            // Manages WebSocket clients and broadcasts
    │   └── handler.go        // Gin handler for /ws endpoint
    ├── .gitignore
    ├── go.mod
    ├── go.sum
    ├── main.go               // Ties everything together
    ├── PLANNING.md           // This document
    └── README.md
    ```

### 3.2. Development Workflow
    1. **Define Models (`models/models.go`):** Structs for `Position`, `RideSummary`, `RideDetail`.
    2. **Utilities (`util/`):** Implement Haversine formula for distance calculation (`geo.go`), GPS timestamp parsing (`timeutils.go`).
    3. **Configuration (`config/config.go`):** Load settings.
    4. **Database Layer (`database/store.go`):** SQLite setup, table creation, CRUD operations.
    5. **Ride Service (`ride/service.go`, `ride/manager.go`):**
        - `service.go`: Logic for determining ride start/end, naming.
        - `manager.go`: Manages the state of the current ride (if any), receives parsed GPS data, uses the service to make decisions, calls database methods, and triggers WebSocket broadcasts.
    6. **WebSocket Handling (`ws/`):**
        - `hub.go`: Client management, message broadcasting.
        - `handler.go`: Upgrades HTTP to WebSocket, registers clients with the hub.
    7. **API Layer (`api/handlers.go`):** Gin handlers for REST endpoints.
    8. **Update `main.go`:**
        - Initialize config, database, ride manager, WebSocket hub.
        - Modify MQTT message handling in `main.go` or move it:
            - Parse `ShadowDocument` to extract GPS data and main timestamp.
            - Pass `models.Position` data to the `ride.Manager` and `ws.Hub`.
        - Set up Gin router with REST API and WebSocket endpoints.
    9. **Testing:** Unit tests for utilities, ride logic, database functions. Integration tests for API endpoints.

## 4. Key Considerations & Remaining Questions

1.  **Timestamp Precision and Timezones:**
    - The `desired.timestamp` ("HHMMSS.SS") needs to be reliably combined with the date from `ShadowDocument.Timestamp`. Ensure UTC is used consistently for storage and API responses. The `ride naming` based on local time will require careful handling of timezones, possibly configured.
2.  **Error Handling and Resilience:**
    - How should database errors be handled by the `ride.Manager`? Log and continue, or attempt retries for critical operations?
    - MQTT connection drops: The current `mqttsubscriber` has some resilience. Ensure ride state is correctly handled on reconnect (e.g., does a drop automatically end a ride, or can it resume if points are close enough in time?). For now, a drop will likely mean a ride ends due to inactivity.
3.  **Data Volume and Database Performance:**
    - SQLite is fine for moderate data. If volume becomes very high, future optimizations (indexing, archiving) might be needed. Not an immediate concern.
4.  **Concurrency:**
    - The `ride.Manager` will be a critical shared resource accessed by the MQTT callback (writing data) and potentially by other services (reading state). It must be concurrency-safe (e.g., using mutexes).
    - The `ws.Hub` will also manage concurrent client connections.
5.  **Speed Data (`speed_knots`):**
    - The `speed_knots` field is available. Should this be used in ride detection logic? For example, a ride might only be considered "active" if speed is above a certain threshold, in addition to distance changes. (For now, sticking to distance and inactivity).
6.  **Initial State:** What happens if the server starts while a device is already in the middle of what could be a ride? The first few GPS points might be missed for that ride. This is generally acceptable.

## 5. Next Steps (Post-Planning)
    - Confirm any outstanding details from Section 4.
    - Begin implementation, following the workflow in Section 3.2.

# Design Considerations

**How to handle the ride naming if the server's local time is different from the user's timezone.**

The timestamps are in UTC. Since this is a simple prototype, assume the user lives in a PST time zone (UTC-7).

---
**Error Handling for Database Operation**

Yeah, logging errors and continueing seems good. This is a simple prototype, so I'm not too worried about edge cases.

---

**MQTT Connection Drops**

MQTT connection drops on the server side are a rare event. Assume they won't happen.

---

**Speed Data**

The speed data provided by the GPS isn't that great. I don't think its a great idea to use that for inactivity detection.
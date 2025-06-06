package database

import (
	"b3/server/models" // Adjust import path if your module path is different
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// NewStore initializes the PostgreSQL database connection and ensures tables are created.
func NewStore(postgresConnStr string) (*sql.DB, error) {
	if postgresConnStr == "" {
		return nil, fmt.Errorf("PostgreSQL connection string not provided. Please set POSTGRES_CONNECTION_STRING environment variable")
	}

	connStr := postgresConnStr
	// Add binary_parameters=yes to enable binary encoding of parameters [pg bouncer race condition fix]
	if !containsBinaryParams(connStr) {
		separator := "&"
		if !containsParams(connStr) {
			separator = "?"
		}
		connStr += separator + "binary_parameters=yes"
	}

	var err error
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Configure connection pool to prevent prepared statement conflicts
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	if err = createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("PostgreSQL database initialized and tables created successfully")
	return db, nil
}

// Helper function to check if connection string contains query parameters
func containsParams(connStr string) bool {
	return strings.Contains(connStr, "?")
}

// Helper function to check if the connection string contains binary_parameters=no
func containsBinaryParams(connStr string) bool {
	return strings.Contains(connStr, "binary_parameters=no")
}

func createTables(db *sql.DB) error {
	var ridesTableSQL string
	var positionsTableSQL string

	ridesTableSQL = `
	CREATE TABLE IF NOT EXISTS rides (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP
	);`

	positionsTableSQL = `
	CREATE TABLE IF NOT EXISTS ride_positions (
		id SERIAL PRIMARY KEY,
		ride_id INTEGER NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		speed_knots REAL,
		timestamp TIMESTAMP NOT NULL,
		FOREIGN KEY (ride_id) REFERENCES rides(id) ON DELETE CASCADE 
	);`

	if _, err := db.Exec(ridesTableSQL); err != nil {
		return fmt.Errorf("failed to create rides table: %w", err)
	}
	if _, err := db.Exec(positionsTableSQL); err != nil {
		return fmt.Errorf("failed to create ride_positions table: %w", err)
	}
	return nil
}

// CreateRide inserts a new ride into the database.
func CreateRide(db *sql.DB, name string, startTime time.Time) (int64, error) {
	// PostgreSQL doesn't support LastInsertId, use RETURNING instead
	var id int64
	query := "INSERT INTO rides(name, start_time) VALUES($1, $2) RETURNING id"
	err := db.QueryRow(query, name, startTime.UTC()).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to execute CreateRide statement: %w", err)
	}
	return id, nil
}

// AddPositionToRide adds a new GPS position to an existing ride.
func AddPositionToRide(db *sql.DB, rideID int64, position models.Position) error {
	query := "INSERT INTO ride_positions(ride_id, latitude, longitude, speed_knots, timestamp) VALUES($1, $2, $3, $4, $5)"
	_, err := db.Exec(query, rideID, position.Latitude, position.Longitude, position.SpeedKnots, position.Timestamp.UTC()) // Ensure storing in UTC
	if err != nil {
		return fmt.Errorf("failed to execute AddPositionToRide statement: %w", err)
	}
	return nil
}

// EndRide updates the end_time of a ride.
func EndRide(db *sql.DB, rideID int64, endTime time.Time) error {
	query := "UPDATE rides SET end_time = $1 WHERE id = $2"
	_, err := db.Exec(query, endTime.UTC(), rideID) // Ensure storing in UTC
	if err != nil {
		return fmt.Errorf("failed to execute EndRide statement: %w", err)
	}
	return nil
}

// GetRideDetails retrieves a specific ride and all its positions.
func GetRideDetails(db *sql.DB, rideID int64) (*models.RideDetail, error) {
	ride := &models.RideDetail{}
	var endTime sql.NullTime // Handle NULL end_time

	// First query: Get ride details
	rideQuery := "SELECT id, name, start_time, end_time FROM rides WHERE id = $1"
	row := db.QueryRow(rideQuery, rideID)
	if err := row.Scan(&ride.ID, &ride.Name, &ride.StartTime, &endTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ride with ID %d not found", rideID)
		}
		return nil, fmt.Errorf("failed to scan ride details: %w", err)
	}
	if endTime.Valid {
		ride.EndTime = endTime.Time
	}

	// Second query: Get ride positions with a different variable name
	positionsQuery := "SELECT latitude, longitude, speed_knots, timestamp FROM ride_positions WHERE ride_id = $1 ORDER BY timestamp ASC"
	log.Printf("Positions Query: %s", positionsQuery)
	log.Printf("RideID: %d", rideID)

	rows, err := db.Query(positionsQuery, rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ride positions for ride_id %d: %w", rideID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var pos models.Position
		var speedKnots sql.NullFloat64
		if err := rows.Scan(&pos.Latitude, &pos.Longitude, &speedKnots, &pos.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan ride position: %w", err)
		}
		if speedKnots.Valid {
			pos.SpeedKnots = speedKnots.Float64
		}
		ride.Positions = append(ride.Positions, pos)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for ride positions: %w", err)
	}

	// Ensure times are UTC
	ride.StartTime = ride.StartTime.UTC()
	if endTime.Valid {
		ride.EndTime = ride.EndTime.UTC()
	}
	for i := range ride.Positions {
		ride.Positions[i].Timestamp = ride.Positions[i].Timestamp.UTC()
	}

	log.Printf("Successfully retrieved ride %d with %d positions", rideID, len(ride.Positions))
	return ride, nil
}

// GetAllRidesSummary retrieves a summary of all rides.
func GetAllRidesSummary(db *sql.DB) ([]models.RideSummary, error) {
	rows, err := db.Query("SELECT id, name, start_time, end_time FROM rides ORDER BY start_time DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query all rides summary: %w", err)
	}
	defer rows.Close()

	var rides []models.RideSummary
	for rows.Next() {
		var ride models.RideSummary
		var endTime sql.NullTime // Handle NULL end_time
		if err := rows.Scan(&ride.ID, &ride.Name, &ride.StartTime, &endTime); err != nil {
			return nil, fmt.Errorf("failed to scan ride summary: %w", err)
		}
		if endTime.Valid {
			ride.EndTime = endTime.Time
		}
		// Ensure times are UTC
		ride.StartTime = ride.StartTime.UTC()
		if endTime.Valid {
			ride.EndTime = ride.EndTime.UTC()
		}
		rides = append(rides, ride)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for all rides summary: %w", err)
	}
	return rides, nil
}

// GetAllRidesSummaryWithPagination retrieves a summary of rides with pagination and optional date filtering.
func GetAllRidesSummaryWithPagination(db *sql.DB, page, limit int, dateFilter *time.Time) ([]models.RideSummary, error) {
	offset := (page - 1) * limit

	var query string
	var args []interface{}

	if dateFilter != nil {
		// Filter by start date (same day)
		startOfDay := time.Date(dateFilter.Year(), dateFilter.Month(), dateFilter.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := startOfDay.Add(24 * time.Hour)
		query = "SELECT id, name, start_time, end_time FROM rides WHERE start_time >= $1 AND start_time < $2 ORDER BY start_time DESC LIMIT $3 OFFSET $4"
		args = []interface{}{startOfDay, endOfDay, limit, offset}
	} else {
		query = "SELECT id, name, start_time, end_time FROM rides ORDER BY start_time DESC LIMIT $1 OFFSET $2"
		args = []interface{}{limit, offset}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query rides summary with pagination: %w", err)
	}
	defer rows.Close()

	var rides []models.RideSummary
	for rows.Next() {
		var ride models.RideSummary
		var endTime sql.NullTime // Handle NULL end_time
		if err := rows.Scan(&ride.ID, &ride.Name, &ride.StartTime, &endTime); err != nil {
			return nil, fmt.Errorf("failed to scan ride summary: %w", err)
		}
		if endTime.Valid {
			ride.EndTime = endTime.Time
		}
		// Ensure times are UTC
		ride.StartTime = ride.StartTime.UTC()
		if endTime.Valid {
			ride.EndTime = ride.EndTime.UTC()
		}
		rides = append(rides, ride)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for paginated rides summary: %w", err)
	}
	return rides, nil
}

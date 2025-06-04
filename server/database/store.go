package database

import (
	"b3/server/config" // To get DB path and other configs
	"b3/server/models" // Adjust import path if your module path is different
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var dbInstance *sql.DB

// InitDB initializes the database connection using the configuration
// and ensures tables are created.
func InitDB() (*sql.DB, error) {
	cfg := config.Get()

	var err error
	var driver string
	var dataSource string

	switch cfg.DatabaseType {
	case "postgres":
		if cfg.PostgresConnStr == "" {
			return nil, fmt.Errorf("PostgreSQL connection string not provided")
		}
		driver = "postgres"
		dataSource = cfg.PostgresConnStr
		log.Println("Connecting to PostgreSQL database...")
	case "sqlite":
		driver = "sqlite3"
		dataSource = cfg.DatabasePath + "?_foreign_keys=on" // Enable foreign key support
		log.Printf("Connecting to SQLite database at %s...", cfg.DatabasePath)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DatabaseType)
	}

	dbInstance, err = sql.Open(driver, dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = dbInstance.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err = createTables(dbInstance, cfg.DatabaseType); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Printf("Database initialized and tables created successfully using %s", cfg.DatabaseType)
	return dbInstance, nil
}

// GetDB returns the current database instance.
// Panics if InitDB has not been called successfully.
func GetDB() *sql.DB {
	if dbInstance == nil {
		log.Fatal("Database has not been initialized. Call InitDB first.")
	}
	return dbInstance
}

func createTables(db *sql.DB, dbType string) error {
	var ridesTableSQL string
	var positionsTableSQL string

	switch dbType {
	case "postgres":
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
	case "sqlite":
		ridesTableSQL = `
		CREATE TABLE IF NOT EXISTS rides (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME
		);`

		positionsTableSQL = `
		CREATE TABLE IF NOT EXISTS ride_positions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ride_id INTEGER NOT NULL,
			latitude REAL NOT NULL,
			longitude REAL NOT NULL,
			speed_knots REAL,
			timestamp DATETIME NOT NULL,
			FOREIGN KEY (ride_id) REFERENCES rides(id) ON DELETE CASCADE 
		);` // Added ON DELETE CASCADE
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

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
	stmt, err := db.Prepare("INSERT INTO rides(name, start_time) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare CreateRide statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(name, startTime.UTC()) // Ensure storing in UTC
	if err != nil {
		return 0, fmt.Errorf("failed to execute CreateRide statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID for CreateRide: %w", err)
	}
	return id, nil
}

// AddPositionToRide adds a new GPS position to an existing ride.
func AddPositionToRide(db *sql.DB, rideID int64, position models.Position) error {
	stmt, err := db.Prepare("INSERT INTO ride_positions(ride_id, latitude, longitude, speed_knots, timestamp) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare AddPositionToRide statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(rideID, position.Latitude, position.Longitude, position.SpeedKnots, position.Timestamp.UTC()) // Ensure storing in UTC
	if err != nil {
		return fmt.Errorf("failed to execute AddPositionToRide statement: %w", err)
	}
	return nil
}

// EndRide updates the end_time of a ride.
func EndRide(db *sql.DB, rideID int64, endTime time.Time) error {
	stmt, err := db.Prepare("UPDATE rides SET end_time = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare EndRide statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(endTime.UTC(), rideID) // Ensure storing in UTC
	if err != nil {
		return fmt.Errorf("failed to execute EndRide statement: %w", err)
	}
	return nil
}

// GetRideDetails retrieves a specific ride and all its positions.
func GetRideDetails(db *sql.DB, rideID int64) (*models.RideDetail, error) {
	ride := &models.RideDetail{}
	var endTime sql.NullTime // Handle NULL end_time

	row := db.QueryRow("SELECT id, name, start_time, end_time FROM rides WHERE id = ?", rideID)
	if err := row.Scan(&ride.ID, &ride.Name, &ride.StartTime, &endTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ride with ID %d not found", rideID)
		}
		return nil, fmt.Errorf("failed to scan ride details: %w", err)
	}
	if endTime.Valid {
		ride.EndTime = endTime.Time
	}

	rows, err := db.Query("SELECT latitude, longitude, speed_knots, timestamp FROM ride_positions WHERE ride_id = ? ORDER BY timestamp ASC", rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ride positions: %w", err)
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

	return ride, nil
}

// GetAllRidesSummary retrieves a summary of all rides.
// TODO: Add pagination and filtering as per PLANNING.md.
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
		query = "SELECT id, name, start_time, end_time FROM rides WHERE start_time >= ? AND start_time < ? ORDER BY start_time DESC LIMIT ? OFFSET ?"
		args = []interface{}{startOfDay, endOfDay, limit, offset}
	} else {
		query = "SELECT id, name, start_time, end_time FROM rides ORDER BY start_time DESC LIMIT ? OFFSET ?"
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

// CloseDB closes the database connection.
func CloseDB() {
	if dbInstance != nil {
		err := dbInstance.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed.")
		}
	}
}

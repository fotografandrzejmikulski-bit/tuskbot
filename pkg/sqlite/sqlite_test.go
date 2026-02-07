package sqlite

import (
	"database/sql"
	"testing"
)

func TestSQLiteVecExtension(t *testing.T) {
	// 1. Open database using the custom driver registered in init.go
	db, err := sql.Open("sqlite3_vec", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 2. Check connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// 3. Verify extension is loaded by calling a specific function from sqlite-vec
	var version string
	err = db.QueryRow("SELECT vec_version()").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to query vec_version(): %v. \nIt seems the extension is not linked or loaded correctly.", err)
	}

	t.Logf("Success! SQLite-Vec Version: %s", version)

	if version == "" {
		t.Error("Expected a version string, got empty")
	}
}

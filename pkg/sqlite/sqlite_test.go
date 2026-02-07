package sqlite

import (
	"bytes"
	"database/sql"
	"encoding/binary"
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

func TestMessageVectorRelation(t *testing.T) {
	db, err := sql.Open("sqlite3_vec", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 1. Setup schema (mimicking the app's schema)
	_, err = db.Exec(`CREATE TABLE messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}

	// Create virtual table for vectors.
	// Note: In sqlite-vec, rowid is the default primary key.
	_, err = db.Exec(`CREATE VIRTUAL TABLE messages_vec USING vec0(embedding float[3])`)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Insert a dummy message
	content := "test message content"
	res, err := db.Exec(`INSERT INTO messages (content) VALUES (?)`, content)
	if err != nil {
		t.Fatal(err)
	}
	msgID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	// 3. Insert vector tied to the message ID via rowid
	vec := []float32{0.1, 0.2, 0.3}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, vec); err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO messages_vec(rowid, embedding) VALUES (?, ?)`, msgID, buf.Bytes())
	if err != nil {
		t.Fatalf("Failed to insert vector with rowid: %v", err)
	}

	// 4. Verify the relation using a JOIN
	var retrievedContent string
	err = db.QueryRow(`
		SELECT m.content 
		FROM messages m 
		JOIN messages_vec v ON m.id = v.rowid 
		WHERE v.rowid = ?`, msgID).Scan(&retrievedContent)

	if err != nil {
		t.Fatalf("JOIN query failed: %v. This means the vector is not correctly linked to the message ID.", err)
	}

	if retrievedContent != content {
		t.Errorf("Expected content '%s', but got '%s'", content, retrievedContent)
	}

	t.Log("Success: Message and Vector are correctly linked via rowid")
}

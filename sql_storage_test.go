package filesystem

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func sqlStorageInitDB(filepath string) *sql.DB {
	os.Remove(filepath) // remove database
	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		panic(err)
	}

	return db
}

func TestSqlStoragePut(t *testing.T) {
	db := sqlStorageInitDB(":memory:")

	s, err := NewSqlStorage(SqlStorageOptions{
		DB:                 db,
		FilestoreTable:     "sqlstore",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if s == nil {
		t.Fatal("NewSqlStorage() returned nil")
	}

	err = s.Put("test.txt", []byte("test"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestSqlStorageReadFile(t *testing.T) {
	db := sqlStorageInitDB(":memory:")

	s, err := NewSqlStorage(SqlStorageOptions{
		DB:                 db,
		FilestoreTable:     "sqlstore",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if s == nil {
		t.Fatal("NewSqlStorage() returned nil")
	}

	err = s.Put("test.txt", []byte("test"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	data, err := s.ReadFile("test.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if string(data) != "test" {
		t.Fatal("unexpected data:", string(data))
	}

}

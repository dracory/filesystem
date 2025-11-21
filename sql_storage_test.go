package filesystem

import (
	"database/sql"
	"os"
	"strings"
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

func newTestSqlStorage(t *testing.T) *SQLStorage {
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

	return s
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

func TestSqlStorageExistsAndSize(t *testing.T) {
	s := newTestSqlStorage(t)

	err := s.Put("size.txt", []byte("abcdef"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	exists, err := s.Exists("size.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if !exists {
		t.Fatal("expected file to exist")
	}

	size, err := s.Size("size.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if size != 6 {
		t.Fatalf("unexpected size: %d", size)
	}
}

func TestSqlStorageDirectoriesAndFilesRoot(t *testing.T) {
	s := newTestSqlStorage(t)

	err := s.MakeDirectory("dir")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = s.Put("file1.txt", []byte("a"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = s.Put("file2.txt", []byte("b"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	dirs, err := s.Directories("/")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(dirs) == 0 {
		t.Fatal("expected at least one directory")
	}

	files, err := s.Files("/")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if len(files) != 2 {
		t.Fatalf("unexpected files count: %d", len(files))
	}
}

func TestSqlStorageLastModified(t *testing.T) {
	s := newTestSqlStorage(t)

	err := s.Put("last.txt", []byte("x"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	modified, err := s.LastModified("last.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if modified.IsZero() {
		t.Fatal("expected non-zero last modified time")
	}
}

func TestSqlStorageUrl(t *testing.T) {
	db := sqlStorageInitDB(":memory:")

	s, err := NewSqlStorage(SqlStorageOptions{
		DB:                 db,
		FilestoreTable:     "sqlstore",
		AutomigrateEnabled: true,
		URL:                "https://example.com",
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if s == nil {
		t.Fatal("NewSqlStorage() returned nil")
	}

	err = s.Put("url.txt", []byte("x"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	u, err := s.Url("url.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if u == "" {
		t.Fatal("expected non-empty url")
	}

	if !strings.HasPrefix(u, "https://example.com") {
		t.Fatalf("unexpected url: %s", u)
	}
}

func TestSqlStorageDeleteFile(t *testing.T) {
	s := newTestSqlStorage(t)

	err := s.Put("delete.txt", []byte("x"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = s.DeleteFile([]string{"delete.txt"})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	exists, err := s.Exists("delete.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if exists {
		t.Fatal("expected file to be deleted")
	}
}

func TestSqlStorageMoveAndCopy(t *testing.T) {
	s := newTestSqlStorage(t)

	err := s.Put("move.txt", []byte("x"))

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = s.Move("move.txt", "moved.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	existsOld, err := s.Exists("move.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if existsOld {
		t.Fatal("expected old path to not exist after move")
	}

	existsNew, err := s.Exists("moved.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if !existsNew {
		t.Fatal("expected new path to exist after move")
	}

	err = s.Copy("moved.txt", "copied.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	existsCopy, err := s.Exists("copied.txt")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if !existsCopy {
		t.Fatal("expected copied file to exist")
	}
}

func TestSqlStorageReadFileNotFound(t *testing.T) {
	s := newTestSqlStorage(t)

	_, err := s.ReadFile("does-not-exist.txt")

	if err == nil {
		t.Fatal("expected error when reading missing file")
	}
}

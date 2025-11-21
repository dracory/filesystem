package filesystem

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNewStorageDiskEmpty(t *testing.T) {
	_, err := NewStorage(Disk{})

	if err == nil {
		t.Fatal("expected error for empty disk")
	}
}

func TestNewStorageDriverRequired(t *testing.T) {
	_, err := NewStorage(Disk{Url: "https://example.com"})

	if err == nil {
		t.Fatal("expected error when driver is missing")
	}
}

func TestNewStorageUrlRequired(t *testing.T) {
	_, err := NewStorage(Disk{Driver: DRIVER_SQL})

	if err == nil {
		t.Fatal("expected error when url is missing")
	}
}

func TestNewStorageUnsupportedDriver(t *testing.T) {
	_, err := NewStorage(Disk{Driver: "unknown", Url: "https://example.com"})

	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}

func TestNewStorageStatic(t *testing.T) {
	storage, err := NewStorage(Disk{
		Driver: DRIVER_STATIC,
		Url:    "https://cdn.example.com",
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if storage == nil {
		t.Fatal("expected non-nil storage")
	}

	if _, ok := storage.(*StaticStorage); !ok {
		t.Fatalf("expected *StaticStorage, got %T", storage)
	}
}

func TestNewStorageSQL(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:?parseTime=true")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	storage, err := NewStorage(Disk{
		Driver:    DRIVER_SQL,
		Url:       "https://example.com",
		DB:        db,
		TableName: "sqlstore",
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if storage == nil {
		t.Fatal("expected non-nil storage")
	}

	if _, ok := storage.(*SQLStorage); !ok {
		t.Fatalf("expected *SQLStorage, got %T", storage)
	}
}

func TestNewStorageS3MissingFields(t *testing.T) {
	// Missing region
	_, err := NewStorage(Disk{
		Driver: DRIVER_S3,
		Url:    "https://s3.example.com",
	})

	if err == nil {
		t.Fatal("expected error when S3 region is missing")
	}

	// Missing key
	_, err = NewStorage(Disk{
		Driver: DRIVER_S3,
		Url:    "https://s3.example.com",
		Region: "us-east-1",
	})

	if err == nil {
		t.Fatal("expected error when S3 key is missing")
	}

	// Missing secret
	_, err = NewStorage(Disk{
		Driver: DRIVER_S3,
		Url:    "https://s3.example.com",
		Region: "us-east-1",
		Key:    "key",
	})

	if err == nil {
		t.Fatal("expected error when S3 secret is missing")
	}
}

func TestNewStorageS3Success(t *testing.T) {
	storage, err := NewStorage(Disk{
		Driver: DRIVER_S3,
		Url:    "https://bucket.s3.example.com",
		Region: "us-east-1",
		Key:    "key",
		Secret: "secret",
		Bucket: "bucket",
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if storage == nil {
		t.Fatal("expected non-nil storage")
	}

	if _, ok := storage.(*S3Storage); !ok {
		t.Fatalf("expected *S3Storage, got %T", storage)
	}
}

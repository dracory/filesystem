# Filesystem

[![Tests](https://github.com/dracory/filesystem/actions/workflows/tests.yml/badge.svg)](https://github.com/dracory/filesystem/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/filesystem)](https://goreportcard.com/report/github.com/dracory/filesystem)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/filesystem)](https://pkg.go.dev/github.com/dracory/filesystem)

`filesystem` provides a lightweight, driver-based abstraction over file storage backends. It exposes a single `StorageInterface` with helpers for common operations—copying, moving, deleting, listing, and reading files—while delegating the underlying work to the disk driver you configure.

## Features

- Unified interface for reading, writing, listing, and deleting files and directories @StorageInterface.go#5-18
- Disk configuration struct that captures S3, SQL, and static driver options @Disk.go#5-30
- S3 driver backed by the AWS SDK for Go v2, with automatic endpoint resolution and public-read uploads @S3Storage.go#21-315
- SQL driver that persists files inside a relational database using `github.com/dracory/sqlfilestore`, including optional automigrations @SqlStorage.go#17-406
- Static driver for CDN-style, read-only access that constructs public URLs without mutating state @StaticStorage.go#9-68

## Installation

```bash
go get github.com/dracory/filesystem
```

## Getting Started

All drivers begin with a `Disk` definition that captures the settings for the backend. `NewStorage` validates the disk and returns an implementation of `StorageInterface`. @Storage.go#11-87

```go
package main

import (
    "log"

    "github.com/dracory/filesystem"
)

func main() {
    storage, err := filesystem.NewStorage(filesystem.Disk{
        DiskName:             "media",
        Driver:               filesystem.DRIVER_S3,
        Url:                  "https://example.cdn.net",
        Region:               "us-east-1",
        Key:                  "ACCESS_KEY",
        Secret:               "SECRET_KEY",
        Bucket:               "my-bucket",
        UsePathStyleEndpoint: true,
    })
    if err != nil {
        log.Fatal(err)
    }

    if err := storage.Put("uploads/hello.txt", []byte("hello world")); err != nil {
        log.Fatal(err)
    }
}
```

### S3 driver

The S3 driver targets any S3-compatible provider. Required disk settings are `Url`, `Region`, `Key`, `Secret`, and `Bucket`. Optional `UsePathStyleEndpoint` toggles path-style addressing for services like MinIO. @Disk.go#18-28

The driver supports:

1. Object CRUD (`Copy`, `Move`, `DeleteFile`, `ReadFile`, `Put`) @S3Storage.go#52-315
2. Directory helpers (`Directories`, `Files`, `DeleteDirectory`, `MakeDirectory`) @S3Storage.go#94-266
3. Existence checks and signed metadata through `Exists` and `Missing` @S3Storage.go#207-257

### SQL driver

The SQL driver stores files and directories inside a database table managed by [`github.com/dracory/sqlfilestore`](https://github.com/dracory/sqlfilestore). Use `NewSqlStorage` when you need a structured datastore or offline storage.

```go
db, _ := sql.Open("sqlite", ":memory:")
storage, err := filesystem.NewSqlStorage(filesystem.SqlStorageOptions{
    DB:                 db,
    FilestoreTable:     "filestore",
    URL:                "https://media.example.com",
    AutomigrateEnabled: true,
})
```

- `Put`, `ReadFile`, `Move`, `DeleteFile`, and `DeleteDirectory` map to SQL operations and maintain directory hierarchies. @SqlStorage.go#99-368
- Helpers like `Exists`, `Files`, and `Directories` normalise paths and return database-backed results. @SqlStorage.go#235-369
- Optional URL prefixing allows generating absolute media URLs via `Url()`. @SqlStorage.go#503-516

### Static driver

The static driver exposes CDN-like, read-only access. It returns deterministic URLs and rejects mutating operations. Use it to map a logical disk to pre-existing assets.

```go
storage := &filesystem.StaticStorage{disk: filesystem.Disk{
    DiskName: "cdn",
    Driver:   filesystem.DRIVER_STATIC,
    Url:      "https://cdn.example.com",
}}

url, _ := storage.Url("images/logo.png")
// url == "https://cdn.example.com/images/logo.png"
```

## Testing

```bash
go test ./...
```

The SQL storage driver ships with basic tests covering file persistence. @SqlStorage_test.go#23-80

## License

This project is released under the MIT License. See [LICENSE](LICENSE) for details.

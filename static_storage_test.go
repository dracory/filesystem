package filesystem

import "testing"

func newTestStaticStorage() *StaticStorage {
	return &StaticStorage{
		disk: Disk{
			Url: "https://cdn.example.com/",
		},
	}
}

func TestStaticStorageUrl(t *testing.T) {
	storage := newTestStaticStorage()

	url, err := storage.Url("/path/to/file.jpg")

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if url != "https://cdn.example.com/path/to/file.jpg" {
		t.Fatalf("unexpected url: %s", url)
	}
}

func TestStaticStorageUnsupportedOperations(t *testing.T) {
	storage := newTestStaticStorage()

	if err := storage.Copy("a", "b"); err == nil {
		t.Fatal("expected error from Copy")
	}

	if err := storage.DeleteFile([]string{"a"}); err == nil {
		t.Fatal("expected error from DeleteFile")
	}

	if err := storage.DeleteDirectory("/dir"); err == nil {
		t.Fatal("expected error from DeleteDirectory")
	}

	if _, err := storage.Directories("/"); err == nil {
		t.Fatal("expected error from Directories")
	}

	if _, err := storage.Exists("/file"); err == nil {
		t.Fatal("expected error from Exists")
	}

	if _, err := storage.Files("/"); err == nil {
		t.Fatal("expected error from Files")
	}

	if err := storage.MakeDirectory("/dir"); err == nil {
		t.Fatal("expected error from MakeDirectory")
	}

	if _, err := storage.LastModified("/file"); err == nil {
		t.Fatal("expected error from LastModified")
	}

	if err := storage.Move("a", "b"); err == nil {
		t.Fatal("expected error from Move")
	}

	if _, err := storage.ReadFile("/file"); err == nil {
		t.Fatal("expected error from ReadFile")
	}

	if _, err := storage.Size("/file"); err == nil {
		t.Fatal("expected error from Size")
	}

	if err := storage.Put("/file", []byte("x")); err == nil {
		t.Fatal("expected error from Put")
	}
}

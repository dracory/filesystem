package filesystem

import (
	"errors"
	"strings"
	"time"
)

// StaticStorage implements StorageInterface
// it represents a static (read only) storage, i.e. CDN
// the only supported method is Url(filepath)
type StaticStorage struct {
	disk Disk
}

var _ StorageInterface = (*StaticStorage)(nil) // verify it extends the task interface

func (s *StaticStorage) Copy(originFile, targetFile string) error {
	return errors.New("not supported")
}

func (s *StaticStorage) DeleteFile(filePaths []string) error {
	return errors.New("not implemented")
}

func (s *StaticStorage) DeleteDirectory(dirPath string) error {
	return errors.New("not implemented")
}

func (s *StaticStorage) Directories(dirPath string) ([]string, error) {
	return []string{}, errors.New("not implemented")
}

func (s *StaticStorage) Exists(filePath string) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *StaticStorage) Files(dirPath string) ([]string, error) {
	return []string{}, errors.New("not implemented")
}

func (s *StaticStorage) MakeDirectory(dirPath string) error {
	return errors.New("not implemented")
}

func (s *StaticStorage) LastModified(filePath string) (time.Time, error) {
	return time.Time{}, errors.New("not implemented")
}

func (s *StaticStorage) Move(originFile, targetFile string) error {
	return errors.New("not implemented")
}

func (s *StaticStorage) ReadFile(filePath string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (s *StaticStorage) Size(filePath string) (int64, error) {
	return -1, errors.New("not implemented")
}

func (s *StaticStorage) Url(filePath string) (string, error) {
	return strings.TrimRight(s.disk.Url, "/") + "/" + strings.TrimLeft(filePath, "/"), errors.New("not implemented")
}

func (s *StaticStorage) Put(filePath string, content []byte) error {
	return errors.New("not implemented")
}

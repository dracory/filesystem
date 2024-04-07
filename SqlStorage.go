package filesystem

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/emirpasic/gods/utils"
	"github.com/golang-module/carbon/v2"
	"github.com/gouniverse/sb"
	"github.com/gouniverse/sqlfilestore"
)

var _ StorageInterface = (*SQLStorage)(nil) // verify it extends the storage interface

type SQLStorage struct {
	DB                 *sql.DB
	FilestoreTable     string
	URL                string
	AutomigrateEnabled bool
	DebugEnabled       bool
	store              *sqlfilestore.Store
}

type SqlStorageOptions struct {
	DB                 *sql.DB
	FilestoreTable     string
	URL                string
	AutomigrateEnabled bool
	DebugEnabled       bool
}

func NewSqlStorage(options SqlStorageOptions) (*SQLStorage, error) {
	if options.DB == nil {
		return nil, errors.New("DB is required")
	}

	if options.FilestoreTable == "" {
		return nil, errors.New("FilestoreTable is required")
	}

	storage := &SQLStorage{
		DB:                 options.DB,
		FilestoreTable:     options.FilestoreTable,
		URL:                options.URL,
		AutomigrateEnabled: options.AutomigrateEnabled,
		DebugEnabled:       options.DebugEnabled,
	}

	err := storage.init()

	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *SQLStorage) init() (err error) {
	s.store, err = sqlfilestore.NewStore(sqlfilestore.NewStoreOptions{
		DB:                 s.DB,
		TableName:          s.FilestoreTable,
		AutomigrateEnabled: s.AutomigrateEnabled,
		DebugEnabled:       s.DebugEnabled,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *SQLStorage) findParentDirectoryFromPath(path string) (*sqlfilestore.Record, error) {
	targetPathParts := strings.Split(path, PATH_SEPARATOR)

	if len(targetPathParts) < 1 {
		return nil, errors.New("invalid path")
	}

	targetDirPath := ROOT_PATH + strings.Join(targetPathParts[:len(targetPathParts)-1], PATH_SEPARATOR)

	targetDirectory, err := s.store.RecordFindByPath(targetDirPath, sqlfilestore.RecordQueryOptions{})

	if err != nil {
		return nil, err
	}

	if targetDirectory == nil {
		return nil, nil
	}

	return targetDirectory, nil
}

// DirectoriesList lists the sub-directories in the specified directory
func (s *SQLStorage) DirectoriesList(directoryPath string) ([]string, error) {
	directoryPath = s.fixPath(directoryPath)

	dir, err := s.store.RecordFindByPath(directoryPath, sqlfilestore.RecordQueryOptions{Columns: []string{"id"}})

	if err != nil {
		return nil, err
	}

	if dir == nil {
		return nil, errors.New("directory not found")
	}

	records, err := s.store.RecordList(sqlfilestore.RecordQueryOptions{
		ParentID:  dir.ID(),
		Type:      sqlfilestore.TYPE_DIRECTORY,
		OrderBy:   "path",
		SortOrder: sb.ASC,
	})

	if err != nil {
		return []string{}, err
	}

	if len(records) == 0 {
		return []string{}, nil
	}

	paths := make([]string, len(records))

	for i, record := range records {
		paths[i] = record.Path()
	}

	return paths, nil
}

func (s *SQLStorage) DirectoryCopy(originDirPath, targetDirPath string) error {
	record, err := s.store.RecordFindByPath(originDirPath, sqlfilestore.RecordQueryOptions{})

	if err != nil {
		return err
	}

	if record == nil {
		return errors.New("path not found")
	}

	if !record.IsDirectory() {
		return errors.New("not a directory")
	}

	return errors.New("not implemented")

	// targetDirectory, err := s.findParentDirectoryFromPath(targetDirPath)

	// if err != nil {
	// 	return err
	// }

	// file := sqlfilestore.NewFile().
	// 	SetParentID(targetDirectory.ID()).
	// 	SetName(record.Name()).
	// 	SetContents(record.Contents()).
	// 	SetSize(record.Size()).
	// 	SetExtension(record.Extension()).
	// 	SetPath(targetDirectory.Path() + PATH_SEPARATOR + record.Name())

	// err = s.store.RecordCreate(file)

	// if err != nil {
	// 	return err
	// }

	// return nil
}

func (s *SQLStorage) DirectoryCreate(directoryPath string) error {
	exists, err := s.Exists(directoryPath)

	if err != nil {
		return err
	}

	if exists {
		return errors.New("directory already exists")
	}

	parentDir, err := s.findParentDirectoryFromPath(directoryPath)

	if err != nil {
		return err
	}

	if parentDir == nil {
		return errors.New("parent directory not found")
	}

	directoryName := s.findFileName(directoryPath)

	directory := sqlfilestore.NewDirectory().
		SetParentID(parentDir.ID()).
		SetName(directoryName).
		SetPath(parentDir.Path() + PATH_SEPARATOR + directoryName)

	err = s.store.RecordCreate(directory)

	if err != nil {
		return err
	}

	return nil
}

// DirectoryDelete deletes a directory
func (s *SQLStorage) DirectoryDelete(directoryPath string) error {
	file, err := s.store.RecordFindByPath(directoryPath, sqlfilestore.RecordQueryOptions{
		Columns: []string{
			sqlfilestore.COLUMN_ID,
			sqlfilestore.COLUMN_TYPE,
		},
	})

	if err != nil {
		return err
	}

	if file == nil {
		return nil
	}

	if !file.IsDirectory() {
		return errors.New("not a directory")
	}

	children, err := s.store.RecordList(sqlfilestore.RecordQueryOptions{
		ParentID: file.ID(),
		Columns: []string{
			sqlfilestore.COLUMN_ID,
			sqlfilestore.COLUMN_TYPE,
			sqlfilestore.COLUMN_PATH,
		},
	})

	if err != nil {
		return err
	}

	for _, child := range children {
		if child.IsDirectory() {
			err = s.DirectoryDelete(child.Path())

			if err != nil {
				return err
			}

			continue
		}

		err = s.store.RecordSoftDelete(&child)

		if err != nil {
			return err
		}
	}

	err = s.store.RecordSoftDelete(file)

	return err
}

func (s *SQLStorage) DirectorySize(filePath string) (int64, error) {
	file, err := s.store.RecordFindByPath(filePath, sqlfilestore.RecordQueryOptions{Columns: []string{"size"}})

	if err != nil {
		return -1, err
	}

	sizeString := file.Size()

	if sizeString == "" {
		return 0, nil
	}

	size, err := strconv.ParseInt(sizeString, 10, 64)

	if err != nil {
		return -1, err
	}

	return size, nil
}

func (s *SQLStorage) DirectoryUrl(filePath string) (string, error) {
	file, err := s.store.RecordFindByPath(filePath, sqlfilestore.RecordQueryOptions{Columns: []string{"path"}})

	if err != nil {
		return "", err
	}

	path := file.Path()

	if s.URL != "" {
		path = s.URL + path
	}

	return path, nil
}

func (s *SQLStorage) FileCopy(originFilePath, targetFilePath string) error {
	record, err := s.store.RecordFindByPath(originFilePath, sqlfilestore.RecordQueryOptions{})

	if err != nil {
		return err
	}

	if record == nil {
		return errors.New("path not found")
	}

	if !record.IsFile() {
		return errors.New("not a file")
	}

	targetDirectory, err := s.findParentDirectoryFromPath(targetFilePath)

	if err != nil {
		return err
	}

	file := sqlfilestore.NewFile().
		SetParentID(targetDirectory.ID()).
		SetName(record.Name()).
		SetContents(record.Contents()).
		SetSize(record.Size()).
		SetExtension(record.Extension()).
		SetPath(targetDirectory.Path() + PATH_SEPARATOR + record.Name())

	err = s.store.RecordCreate(file)

	if err != nil {
		return err
	}

	return nil
}

func (s *SQLStorage) FileDelete(filePaths []string) error {
	for _, filePath := range filePaths {
		record, err := s.store.RecordFindByPath(filePath, sqlfilestore.RecordQueryOptions{
			Columns: []string{
				sqlfilestore.COLUMN_ID,
				sqlfilestore.COLUMN_TYPE,
				sqlfilestore.COLUMN_PATH,
			},
		})

		if err != nil {
			return err
		}

		if record == nil {
			continue
		}

		if record.IsDirectory() {
			err = s.DirectoryDelete(record.Path())

			if err != nil {
				return err
			}

			continue
		}

		if record.IsFile() {
			err = s.store.RecordSoftDelete(record)

			if err != nil {
				return err
			}
		}

		return errors.New("not a file or directory: " + record.Path())
	}
	return nil
}

// FilesList lists the files in the specified directory
func (s *SQLStorage) FilesList(directoryPath string) ([]string, error) {
	directoryPath = s.fixPath(directoryPath)

	dir, err := s.store.RecordFindByPath(directoryPath, sqlfilestore.RecordQueryOptions{Columns: []string{"id"}})

	if err != nil {
		return nil, err
	}

	if dir == nil {
		return nil, errors.New("directory not found")
	}

	records, err := s.store.RecordList(sqlfilestore.RecordQueryOptions{
		ParentID:  dir.ID(),
		Type:      sqlfilestore.TYPE_FILE,
		OrderBy:   "path",
		SortOrder: sb.ASC,
	})

	if err != nil {
		return []string{}, err
	}

	if len(records) == 0 {
		return []string{}, nil
	}

	paths := make([]string, len(records))

	for i, record := range records {
		paths[i] = record.Path()
	}

	return paths, nil
}

func (s *SQLStorage) Exists(path string) (bool, error) {
	fixedPath := s.fixPath(path)

	count, err := s.store.RecordCount(sqlfilestore.RecordQueryOptions{
		Path:    fixedPath,
		Columns: []string{"id"},
	})

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}



func (s *SQLStorage) FileLastModified(filePath string) (time.Time, error) {
	file, err := s.store.RecordFindByPath(filePath, sqlfilestore.RecordQueryOptions{Columns: []string{"updated_at"}})

	if err != nil {
		return carbon.Parse(sb.NULL_DATETIME).StdTime(), err
	}

	strUpdatedAt := file.UpdatedAt()

	return carbon.Parse(strUpdatedAt, carbon.UTC).StdTime(), nil
}

func (s *SQLStorage) FilePut(filePath string, content []byte) error {
	parentDir, err := s.findParentDirectoryFromPath(filePath)

	if err != nil {
		return err
	}

	if parentDir == nil {
		return errors.New("parent directory not found")
	}

	b64 := base64.StdEncoding.EncodeToString(content)
	fileName := s.findFileName(filePath)
	fileExtension := s.findExtension(filePath)
	filePath = parentDir.Path() + PATH_SEPARATOR + fileName

	file := sqlfilestore.NewFile().
		SetParentID(parentDir.ID()).
		SetName(fileName).
		SetContents(b64).
		SetExtension(fileExtension).
		SetPath(filePath).
		SetSize(utils.ToString(len(content)))

	err = s.store.RecordCreate(file)

	if err != nil {
		return err
	}

	return nil
}

func (s *SQLStorage) FileRead(filePath string) ([]byte, error) {
	file, err := s.store.RecordFindByPath(filePath, sqlfilestore.RecordQueryOptions{Columns: []string{"contents"}})

	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, errors.New("file not found")
	}

	b, err := base64.StdEncoding.DecodeString(file.Contents())

	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *SQLStorage) FileSize(filePath string) (int64, error) {
	file, err := s.store.RecordFindByPath(filePath, sqlfilestore.RecordQueryOptions{Columns: []string{"size"}})

	if err != nil {
		return -1, err
	}

	sizeString := file.Size()

	if sizeString == "" {
		return 0, nil
	}

	size, err := strconv.ParseInt(sizeString, 10, 64)

	if err != nil {
		return -1, err
	}

	return size, nil
}

func (s *SQLStorage) FileUrl(filePath string) (string, error) {
	file, err := s.store.RecordFindByPath(filePath, sqlfilestore.RecordQueryOptions{Columns: []string{"path"}})

	if err != nil {
		return "", err
	}

	path := file.Path()

	if s.URL != "" {
		path = s.URL + path
	}

	return path, nil
}

// List lists the files in the specified directory
func (s *SQLStorage) List(directoryPath string) ([]string, error) {
	directoryPath = s.fixPath(directoryPath)

	dir, err := s.store.RecordFindByPath(directoryPath, sqlfilestore.RecordQueryOptions{Columns: []string{"id"}})

	if err != nil {
		return nil, err
	}

	if dir == nil {
		return nil, errors.New("directory not found")
	}

	records, err := s.store.RecordList(sqlfilestore.RecordQueryOptions{
		ParentID:  dir.ID(),
		OrderBy:   "path",
		SortOrder: sb.ASC,
	})

	if err != nil {
		return []string{}, err
	}

	if len(records) == 0 {
		return []string{}, nil
	}

	paths := make([]string, len(records))

	for i, record := range records {
		paths[i] = record.Path()
	}

	return paths, nil
}

func (s *SQLStorage) Move(originFilePath, targetFilePath string) error {
	if originFilePath == targetFilePath {
		return errors.New("origin and target paths are the same")
	}

	record, err := s.store.RecordFindByPath(originFilePath, sqlfilestore.RecordQueryOptions{
		Columns: []string{
			sqlfilestore.COLUMN_ID,
			sqlfilestore.COLUMN_PARENT_ID,
		},
	})

	if err != nil {
		return err
	}

	if record == nil {
		return errors.New("origin file or folder path not found")
	}

	targetDirectory, err := s.findParentDirectoryFromPath(targetFilePath)

	if err != nil {
		return err
	}

	if targetDirectory == nil {
		return errors.New("target directory not found")
	}

	newName := s.findFileName(targetFilePath)

	record.SetParentID(targetDirectory.ID())
	record.SetName(newName)
	record.SetPath(targetDirectory.Path() + PATH_SEPARATOR + newName)

	err = s.store.RecordUpdate(record)

	if err != nil {
		return err
	}

	return s.store.RecordRecalculatePath(record, targetDirectory)
}

func (s *SQLStorage) fixPath(path string) string {
	if strings.HasPrefix(path, PATH_SEPARATOR) {
		return path
	}

	return PATH_SEPARATOR + path
}

// findExtension finds the file extension from a path.
//
// Parameter(s):
//   - path string - the path
//
// Return type(s):
//   - string - the file extension
func (s *SQLStorage) findExtension(path string) string {
	fileName := s.findFileName(path)

	if fileName == "" {
		return ""
	}

	nameParts := strings.Split(fileName, ".")

	if len(nameParts) < 2 {
		return ""
	}

	return nameParts[1]
}

// findFileName finds the file name from a path.
//
// Parameter(s):
//   - path string - the path
//
// Return type(s):
//   - string - the file name
func (s *SQLStorage) findFileName(path string) string {
	uriParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(uriParts) < 1 {
		return ""
	}

	return uriParts[len(uriParts)-1]
}

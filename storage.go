package filesystem

import (
	// "project/config"

	"errors"

	"github.com/samber/lo"
)

func NewStorage(disk Disk) (StorageInterface, error) {
	// var disks = map[string]Disk{
	// 	DEFAULT: {
	// 		DiskName: DEFAULT,
	// 		Driver:   DRIVER_S3,
	// 		Url:                  config.MediaUrl,
	// 		Region:               config.MediaRegion,
	// 		Key:                  config.MediaKey,
	// 		Secret:               config.MediaSecret,
	// 		UsePathStyleEndpoint: true,
	// 	},
	// 	CDN: {
	// 		DiskName: SQL,
	// 		Driver:   DRIVER_SQL,
	// 		DB:      db,
	//      TableName: "filestore",
	// 		Url:      config.STORAGE_URL,
	// 	},
	// 	CDN: {
	// 		DiskName: CDN,
	// 		Driver:   DRIVER_STATIC,
	// 		Url:      config.CDN_MEDIA_URL,
	// 	},
	// }

	// disk := lo.ValueOr(disks, diskName, Disk{})

	if lo.IsEmpty(disk) {
		return nil, errors.New("disk cannot be empty")
	}

	if disk.Driver == "" {
		return nil, errors.New("driver is required field")
	}

	if disk.Url == "" {
		return nil, errors.New("url is required field")
	}

	if disk.Driver == DRIVER_S3 && disk.Region == "" {
		return nil, errors.New("region is required field")
	}

	if disk.Driver == DRIVER_S3 && disk.Key == "" {
		return nil, errors.New("key is required field")
	}

	if disk.Driver == DRIVER_S3 && disk.Secret == "" {
		return nil, errors.New("secret is required field")
	}

	if disk.Driver == DRIVER_S3 {
		storage := &S3Storage{
			disk: disk,
		}
		return storage, nil
	}

	if disk.Driver == DRIVER_SQL {
		return NewSqlStorage(SqlStorageOptions{
			DB:                 disk.DB,
			FilestoreTable:     disk.TableName,
			AutomigrateEnabled: true,
			URL:                disk.Url,
		})
	}

	if disk.Driver == DRIVER_STATIC {
		storage := &StaticStorage{
			disk: disk,
		}
		return storage, nil
	}

	//storage := &OsStorage{disk: diskName}
	return nil, errors.New("driver not supported")
}

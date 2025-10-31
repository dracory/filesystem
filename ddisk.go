package filesystem

import "database/sql"

type Disk struct {
	DiskName   string
	Driver     string
	Url        string
	Visibility string // public, private (not implemented)

	// SQL options
	DB        *sql.DB // for sql
	TableName string  // for sql

	// Local options
	Root string // for local filesystem (not implemented)

	// S3 options
	Key      string // for s3
	Secret   string // for s3
	Region   string // for s3
	Bucket   string // for s3
	Endpoint string // for s3

	// Allows you to enable the client to use path-style addressing, i.e.,
	// https://s3.amazonaws.com/BUCKET/KEY . By default, the S3 client will use virtual
	// hosted bucket addressing when possible( https://BUCKET.s3.amazonaws.com/KEY ).
	UsePathStyleEndpoint bool // for s3

	Throw bool // for s3 (not implemented)
}

package s3minio

import "github.com/minio/minio-go/v7"

// BucketLookupType is type of url lookup supported by server.
type BucketLookupType int

// Different types of url lookup supported by the server.Initialized to BucketLookupAuto
const (
	BucketLookupAuto BucketLookupType = iota
	BucketLookupDNS
	BucketLookupPath
)

// Config defines the config for storage.
type Config struct {
	// Bucket
	// Default fiber-bucket
	Bucket string

	// Endpoint is a host name or an IP address
	Endpoint string

	// Region Set this value to override region cache
	// Optional
	Region string

	// BucketLookup Set this value to BucketLookupDNS or BucketLookupPath to override the default bucket lookup
	// Optional, Default is BucketLookupAuto
	BucketLookup BucketLookupType

	// Token Set this value to provide x-amz-security-token (AWS S3 specific)
	// Optional, Default is false
	Token string

	// Secure If set to true, https is used instead of http.
	// Default is false
	Secure bool

	// Reset clears any existing keys in existing Bucket
	// Optional. Default is false
	Reset bool

	// Credentials Minio access key and Minio secret key.
	// Need to be defined
	Credentials Credentials

	// GetObjectOptions Options for GET requests specifying additional options like encryption, If-Match
	GetObjectOptions minio.GetObjectOptions

	// PutObjectOptions
	// Allows user to set optional custom metadata, content headers, encryption keys and number of threads for multipart upload operation.
	PutObjectOptions minio.PutObjectOptions

	// ListObjectsOptions Options per to list objects
	ListObjectsOptions minio.ListObjectsOptions

	// RemoveObjectOptions Allows user to set options
	RemoveObjectOptions minio.RemoveObjectOptions
}

type Credentials struct {
	// AccessKeyID is like user-id that uniquely identifies your account.
	AccessKeyID string
	// SecretAccessKey is the password to your account.
	SecretAccessKey string
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Bucket:              "goe-bucket",
	Endpoint:            "",
	Region:              "",
	BucketLookup:        BucketLookupAuto,
	Token:               "",
	Secure:              false,
	Reset:               false,
	Credentials:         Credentials{},
	GetObjectOptions:    minio.GetObjectOptions{},
	PutObjectOptions:    minio.PutObjectOptions{},
	ListObjectsOptions:  minio.ListObjectsOptions{},
	RemoveObjectOptions: minio.RemoveObjectOptions{},
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.Bucket == "" {
		cfg.Bucket = ConfigDefault.Bucket
	}

	return cfg
}

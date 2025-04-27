# S3/MinIO Storage Module

The S3/MinIO Storage module provides integration with S3-compatible object storage services, including AWS S3 and MinIO. It offers a simple interface for file storage operations.

## Features

- File upload and download
- Bucket management
- Presigned URLs for secure access
- Support for AWS S3, MinIO, and other S3-compatible services
- Configurable storage options

## Usage

### Initialization

The S3/MinIO module can be initialized with configuration from environment variables:

```
# S3/MinIO Configuration
S3_ENDPOINT=play.min.io
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET_NAME=mybucket
S3_REGION=us-east-1
S3_BUCKET_LOOKUP=path  # Options: path, dns
S3_USE_SSL=true
S3_TOKEN=  # Optional session token
```

### Basic Operations

```go
import (
    "context"
    "go.oease.dev/goe/storages/s3minio"
)

func main() {
    // Initialize the S3/MinIO client
    s3Client, err := s3minio.NewMinioClient()
    if err != nil {
        panic(err)
    }
    
    // Upload a file
    ctx := context.Background()
    objectName := "example.txt"
    filePath := "/path/to/local/file.txt"
    contentType := "text/plain"
    
    err = s3Client.UploadFile(ctx, objectName, filePath, contentType)
    if err != nil {
        panic(err)
    }
    
    // Download a file
    err = s3Client.DownloadFile(ctx, objectName, "/path/to/save/downloaded-file.txt")
    if err != nil {
        panic(err)
    }
    
    // Get a presigned URL for temporary access
    url, err := s3Client.GetPresignedURL(ctx, objectName, 24*time.Hour) // 24 hour expiry
    if err != nil {
        panic(err)
    }
    fmt.Println("Presigned URL:", url)
    
    // Check if an object exists
    exists, err := s3Client.ObjectExists(ctx, objectName)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Object %s exists: %v\n", objectName, exists)
    
    // Delete an object
    err = s3Client.DeleteObject(ctx, objectName)
    if err != nil {
        panic(err)
    }
}
```

### Working with Buckets

```go
// Create a new bucket
err := s3Client.CreateBucket(ctx, "new-bucket", "us-east-1")
if err != nil {
    panic(err)
}

// List all buckets
buckets, err := s3Client.ListBuckets(ctx)
if err != nil {
    panic(err)
}
for _, bucket := range buckets {
    fmt.Println("Bucket:", bucket.Name)
}

// List objects in a bucket
objects, err := s3Client.ListObjects(ctx, "mybucket", "prefix/")
if err != nil {
    panic(err)
}
for _, object := range objects {
    fmt.Printf("Object: %s, Size: %d bytes\n", object.Key, object.Size)
}
```

### Advanced Usage

```go
// Upload a file with metadata
metadata := map[string]string{
    "Content-Disposition": "attachment; filename=\"custom-filename.txt\"",
    "Custom-Metadata": "some-value",
}
err = s3Client.UploadFileWithMetadata(ctx, objectName, filePath, contentType, metadata)
if err != nil {
    panic(err)
}

// Get object metadata
metadata, err := s3Client.GetObjectMetadata(ctx, objectName)
if err != nil {
    panic(err)
}
for key, value := range metadata {
    fmt.Printf("%s: %s\n", key, value)
}

// Copy an object
err = s3Client.CopyObject(ctx, "source-object.txt", "destination-object.txt")
if err != nil {
    panic(err)
}
```

## Implementation Details

The S3/MinIO module is built on top of the official MinIO Go client, which provides a comprehensive API for working with S3-compatible storage services.

The module includes:

- Simple methods for common operations (upload, download, delete)
- Bucket management
- Presigned URL generation for secure access
- Metadata handling
- Error handling and logging

The implementation is designed to be flexible and can work with various S3-compatible services by adjusting the configuration parameters.
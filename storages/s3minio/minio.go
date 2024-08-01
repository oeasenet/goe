package s3minio

import (
	"bytes"
	"context"
	"errors"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/valyala/bytebufferpool"
	"log"
	"net/http"
	"sync"
	"time"
)

// Storage interface that is implemented by storage providers
type Storage struct {
	minio *minio.Client
	cfg   Config
	ctx   context.Context
	mu    sync.Mutex
}

// New creates a new storage
func New(config ...Config) *Storage {

	// Set default config
	cfg := configDefault(config...)

	// Minio instance
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.Credentials.AccessKeyID, cfg.Credentials.SecretAccessKey, cfg.Token),
		Secure:       cfg.Secure,
		Region:       cfg.Region,
		BucketLookup: minio.BucketLookupType(cfg.BucketLookup),
	})
	if err != nil {
		panic(err)
	}

	storage := &Storage{minio: minioClient, cfg: cfg, ctx: context.Background()}

	// Reset all entries if set to true
	if cfg.Reset {
		if err = storage.Reset(); err != nil {
			panic(err)
		}
	}

	// check bucket
	err = storage.CheckBucket()
	if err != nil {
		// create bucket
		err = storage.CreateBucket()
		if err != nil {
			panic(err)
		}
	}

	return storage
}

// Get value by key
func (s *Storage) Get(key string) ([]byte, error) {

	if len(key) <= 0 {
		return nil, errors.New("the key value is required")
	}

	// get object
	object, err := s.minio.GetObject(s.ctx, s.cfg.Bucket, key, s.cfg.GetObjectOptions)
	if err != nil {
		return nil, err
	}

	// convert to byte
	bb := bytebufferpool.Get()
	defer bytebufferpool.Put(bb)
	_, err = bb.ReadFrom(object)
	if err != nil {
		return nil, err
	}
	return bb.Bytes(), nil
}

func (s *Storage) MustGet(key string) []byte {
	if len(key) <= 0 {
		return nil
	}
	// get object
	object, err := s.minio.GetObject(s.ctx, s.cfg.Bucket, key, s.cfg.GetObjectOptions)
	if err != nil {
		return nil
	}
	// convert to byte
	bb := bytebufferpool.Get()
	defer bytebufferpool.Put(bb)
	_, err = bb.ReadFrom(object)
	if err != nil {
		return nil
	}
	return bb.Bytes()
}

// Set key with value
// The method `Set` sets the value for a given key in the storage. It creates a Reader from the value byte slice,
// sets the content type in the storage configuration, and then puts the object in the bucket using the Minio client.
// It acquires a lock on the storage mutex to ensure thread safety when setting the configuration options.
// Finally, it returns any errors encountered while putting the object.
func (s *Storage) Set(key string, val []byte, exp time.Duration) error {

	if len(key) <= 0 {
		return errors.New("the key value is required")
	}

	// create Reader
	file := bytes.NewReader(val)

	// set content type
	s.mu.Lock()
	s.cfg.PutObjectOptions.ContentType = http.DetectContentType(val)

	// put object
	_, err := s.minio.PutObject(s.ctx, s.cfg.Bucket, key, file, file.Size(), s.cfg.PutObjectOptions)
	s.mu.Unlock()

	return err
}

// Delete entry by key
func (s *Storage) Delete(key string) error {

	if len(key) <= 0 {
		return errors.New("the key value is required")
	}

	// remove
	err := s.minio.RemoveObject(s.ctx, s.cfg.Bucket, key, s.cfg.RemoveObjectOptions)

	return err
}

// Reset all entries, including unexpired
// This method resets all entries in the storage, including unexpired entries. It deletes all objects in the storage bucket.
// The method achieves this by listing all the objects in the bucket and sending their names to a channel. A separate goroutine
// listens to the channel and removes the objects one by one using the minio client's RemoveObjects method. The method also logs
// any errors encountered during the deletion process.
// DANGER ZONE!!!!!!: This method is dangerous and should be used with caution. It deletes all objects in the storage bucket.
func (s *Storage) Reset() error {

	objectsCh := make(chan minio.ObjectInfo)

	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range s.minio.ListObjects(s.ctx, s.cfg.Bucket, s.cfg.ListObjectsOptions) {
			if object.Err != nil {
				log.Println(object.Err)
			}
			objectsCh <- object
		}
	}()

	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	for err := range s.minio.RemoveObjects(s.ctx, s.cfg.Bucket, objectsCh, opts) {
		log.Println("Error detected during deletion: ", err)
	}

	return nil
}

// Close the storage
func (s *Storage) Close() error {
	return nil
}

// CheckBucket Check to see if bucket already exists
func (s *Storage) CheckBucket() error {
	exists, err := s.minio.BucketExists(s.ctx, s.cfg.Bucket)
	if !exists || err != nil {
		return errors.New("the specified bucket does not exist")
	}
	return nil
}

// CreateBucket creates a new bucket if it does not exist
func (s *Storage) CreateBucket() error {
	return s.minio.MakeBucket(s.ctx, s.cfg.Bucket, minio.MakeBucketOptions{Region: s.cfg.Region})
}

// RemoveBucket removes the bucket if it is empty.
func (s *Storage) RemoveBucket() error {
	return s.minio.RemoveBucket(s.ctx, s.cfg.Bucket)
}

// Conn returns the minio client.
func (s *Storage) Conn() *minio.Client {
	return s.minio
}

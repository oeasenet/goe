package migrate

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	// lockID is the fixed document ID for the migration lock
	lockID = "_lock"
)

// MigrationLock represents the lock document in MongoDB
type MigrationLock struct {
	ID        string    `bson:"_id"`
	LockedBy  string    `bson:"locked_by"`
	LockedAt  time.Time `bson:"locked_at"`
	ExpiresAt time.Time `bson:"expires_at"`
	Heartbeat time.Time `bson:"heartbeat"`
}

// Lock defines the interface for distributed locking
type Lock interface {
	// Acquire attempts to acquire the migration lock
	Acquire(ctx context.Context) error

	// Release releases the migration lock
	Release(ctx context.Context) error

	// Refresh refreshes the lock's heartbeat
	Refresh(ctx context.Context) error

	// IsHeld returns true if the lock is currently held by this instance
	IsHeld() bool
}

// MongoLock implements distributed locking using MongoDB
type MongoLock struct {
	db         *mongo.Database
	collection string
	lockID     string
	hostname   string
	timeout    time.Duration
	heartbeat  time.Duration

	mu       sync.Mutex
	held     bool
	stopChan chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup // Track heartbeat goroutine
}

// NewMongoLock creates a new MongoDB-backed distributed lock
func NewMongoLock(db *mongo.Database, collection string, timeout, heartbeat time.Duration) *MongoLock {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = fmt.Sprintf("unknown-%d", os.Getpid())
	}

	return &MongoLock{
		db:         db,
		collection: collection,
		lockID:     lockID,
		hostname:   hostname,
		timeout:    timeout,
		heartbeat:  heartbeat,
	}
}

// coll returns the collection for lock documents
func (l *MongoLock) coll() *mongo.Collection {
	return l.db.Collection(l.collection)
}

// Acquire attempts to acquire the migration lock
func (l *MongoLock) Acquire(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.held {
		return nil // Already held
	}

	now := time.Now()
	expiresAt := now.Add(l.timeout)

	// First, try to clean up any expired locks
	_, _ = l.coll().DeleteOne(ctx, bson.M{
		"_id":        l.lockID,
		"expires_at": bson.M{"$lt": now},
	})

	// Now try to insert new lock
	lock := MigrationLock{
		ID:        l.lockID,
		LockedBy:  l.hostname,
		LockedAt:  now,
		ExpiresAt: expiresAt,
		Heartbeat: now,
	}

	_, err := l.coll().InsertOne(ctx, lock)
	if err != nil {
		// Check if lock is held by someone else
		if mongo.IsDuplicateKeyError(err) {
			// Lock exists, check who holds it
			var existingLock MigrationLock
			findErr := l.coll().FindOne(ctx, bson.M{"_id": l.lockID}).Decode(&existingLock)
			if findErr == nil {
				return NewLockError("acquire", existingLock.LockedBy, ErrLockAcquisitionFailed)
			}
			return NewLockError("acquire", "", ErrLockAcquisitionFailed)
		}
		return NewLockError("acquire", "", err)
	}

	l.held = true
	l.stopChan = make(chan struct{})
	l.stopOnce = sync.Once{} // Reset for new acquisition

	// Start heartbeat goroutine with WaitGroup tracking
	l.wg.Add(1)
	go l.runHeartbeat()

	return nil
}

// Release releases the migration lock
func (l *MongoLock) Release(ctx context.Context) error {
	l.mu.Lock()

	if !l.held {
		l.mu.Unlock()
		return nil
	}

	// Stop heartbeat - use sync.Once to ensure we only close once
	l.stopOnce.Do(func() {
		if l.stopChan != nil {
			close(l.stopChan)
		}
	})

	// Mark as not held before unlocking
	l.held = false
	l.mu.Unlock()

	// Wait for heartbeat goroutine to finish (with timeout)
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Goroutine finished
	case <-time.After(2 * time.Second):
		// Timeout waiting for goroutine
	}

	// Only delete if we own the lock
	result, err := l.coll().DeleteOne(ctx, bson.M{
		"_id":       l.lockID,
		"locked_by": l.hostname,
	})
	if err != nil {
		return NewLockError("release", "", err)
	}

	if result.DeletedCount == 0 {
		// Lock was stolen or already released
		return NewLockError("release", "", ErrLockLost)
	}

	return nil
}

// Refresh refreshes the lock's heartbeat and expiration
func (l *MongoLock) Refresh(ctx context.Context) error {
	l.mu.Lock()
	held := l.held
	l.mu.Unlock()

	if !held {
		return ErrLockLost
	}

	now := time.Now()
	expiresAt := now.Add(l.timeout)

	result, err := l.coll().UpdateOne(ctx,
		bson.M{
			"_id":       l.lockID,
			"locked_by": l.hostname,
		},
		bson.M{
			"$set": bson.M{
				"heartbeat":  now,
				"expires_at": expiresAt,
			},
		},
	)
	if err != nil {
		return NewLockError("refresh", "", err)
	}

	if result.MatchedCount == 0 {
		l.mu.Lock()
		l.held = false
		l.mu.Unlock()
		return NewLockError("refresh", "", ErrLockLost)
	}

	return nil
}

// IsHeld returns true if the lock is currently held by this instance
func (l *MongoLock) IsHeld() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.held
}

// runHeartbeat runs the heartbeat goroutine
// This does NOT take a context - it uses stopChan for cancellation
func (l *MongoLock) runHeartbeat() {
	defer l.wg.Done()

	ticker := time.NewTicker(l.heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopChan:
			return
		case <-ticker.C:
			// Use a background context with timeout for heartbeat
			// This ensures heartbeat doesn't depend on caller's context
			refreshCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := l.Refresh(refreshCtx)
			cancel()

			if err != nil {
				// Lock lost, mark as not held and exit
				l.mu.Lock()
				l.held = false
				l.mu.Unlock()
				return
			}
		}
	}
}

// GetLockInfo returns information about the current lock holder
func (l *MongoLock) GetLockInfo(ctx context.Context) (*MigrationLock, error) {
	var lock MigrationLock
	err := l.coll().FindOne(ctx, bson.M{"_id": l.lockID}).Decode(&lock)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &lock, nil
}

// ForceClear forcefully clears the lock (for emergency situations)
// This should only be used for manual intervention
func (l *MongoLock) ForceClear(ctx context.Context) error {
	l.mu.Lock()
	l.held = false
	l.mu.Unlock()

	_, err := l.coll().DeleteOne(ctx, bson.M{"_id": l.lockID})
	return err
}

// Init ensures the lock collection has the necessary indexes
func (l *MongoLock) Init(ctx context.Context) error {
	// Create TTL index on expires_at for automatic cleanup of very old locks
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}

	_, err := l.coll().Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		// Index might already exist, that's fine
		return nil
	}

	return nil
}

// String returns a string representation of the lock
func (l *MongoLock) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return fmt.Sprintf("MongoLock{collection=%s, hostname=%s, held=%v}", l.collection, l.hostname, l.held)
}

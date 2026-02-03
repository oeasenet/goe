package migrate

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MigrationStatus represents the status of a migration
type MigrationStatus string

const (
	// StatusApplied indicates the migration was successfully applied
	StatusApplied MigrationStatus = "applied"

	// StatusFailed indicates the migration failed
	StatusFailed MigrationStatus = "failed"

	// StatusRollingBack indicates the migration is being rolled back
	StatusRollingBack MigrationStatus = "rolling_back"
)

// MigrationRecord represents a migration record in the state store
type MigrationRecord struct {
	ID           any             `bson:"_id,omitempty"`
	Version      int64           `bson:"version"`
	Name         string          `bson:"name"`
	Status       MigrationStatus `bson:"status"`
	AppliedAt    time.Time       `bson:"applied_at"`
	RolledBackAt *time.Time      `bson:"rolled_back_at,omitempty"`
	DurationMs   int64           `bson:"duration_ms"`
	Checksum     string          `bson:"checksum,omitempty"`
	Batch        int             `bson:"batch"`
	Hostname     string          `bson:"hostname"`
	Error        string          `bson:"error,omitempty"`
}

// StateStore defines the interface for migration state persistence
type StateStore interface {
	// Init initializes the state store (creates collection, indexes, etc.)
	Init(ctx context.Context) error

	// GetApplied returns all applied migrations sorted by version
	GetApplied(ctx context.Context) ([]MigrationRecord, error)

	// GetVersion returns the current (highest applied) migration version
	GetVersion(ctx context.Context) (int64, error)

	// GetRecord returns the migration record for a specific version
	GetRecord(ctx context.Context, version int64) (*MigrationRecord, error)

	// MarkApplied records that a migration was successfully applied
	MarkApplied(ctx context.Context, record *MigrationRecord) error

	// MarkFailed records that a migration failed
	MarkFailed(ctx context.Context, version int64, name string, err error, durationMs int64) error

	// MarkRolledBack records that a migration was rolled back
	MarkRolledBack(ctx context.Context, version int64) error

	// GetNextBatch returns the next batch number
	GetNextBatch(ctx context.Context) (int, error)

	// GetDirty returns the dirty migration record if the database is in a dirty state
	GetDirty(ctx context.Context) (*MigrationRecord, error)

	// ClearDirty clears the dirty state (for manual intervention)
	ClearDirty(ctx context.Context, version int64) error
}

// MongoStateStore implements StateStore using MongoDB
type MongoStateStore struct {
	db         *mongo.Database
	collection string
	hostname   string
}

// NewMongoStateStore creates a new MongoDB-backed state store
func NewMongoStateStore(db *mongo.Database, collection, hostname string) *MongoStateStore {
	return &MongoStateStore{
		db:         db,
		collection: collection,
		hostname:   hostname,
	}
}

// coll returns the migration state collection
func (s *MongoStateStore) coll() *mongo.Collection {
	return s.db.Collection(s.collection)
}

// Init initializes the state store by creating necessary indexes
func (s *MongoStateStore) Init(ctx context.Context) error {
	// Create unique index on version
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := s.coll().Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return NewStateError("init", 0, err)
	}

	// Create index on status for quick dirty state checks
	statusIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	}

	_, err = s.coll().Indexes().CreateOne(ctx, statusIndex)
	if err != nil {
		return NewStateError("init", 0, err)
	}

	return nil
}

// GetApplied returns all applied migrations sorted by version
func (s *MongoStateStore) GetApplied(ctx context.Context) ([]MigrationRecord, error) {
	filter := bson.M{"status": StatusApplied}
	opts := options.Find().SetSort(bson.D{{Key: "version", Value: 1}})

	cursor, err := s.coll().Find(ctx, filter, opts)
	if err != nil {
		return nil, NewStateError("get_applied", 0, err)
	}
	defer cursor.Close(ctx)

	var records []MigrationRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, NewStateError("get_applied", 0, err)
	}

	return records, nil
}

// GetVersion returns the current (highest applied) migration version
func (s *MongoStateStore) GetVersion(ctx context.Context) (int64, error) {
	filter := bson.M{"status": StatusApplied}
	opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})

	var record MigrationRecord
	err := s.coll().FindOne(ctx, filter, opts).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, NewStateError("get_version", 0, err)
	}

	return record.Version, nil
}

// GetRecord returns the migration record for a specific version
func (s *MongoStateStore) GetRecord(ctx context.Context, version int64) (*MigrationRecord, error) {
	filter := bson.M{"version": version}

	var record MigrationRecord
	err := s.coll().FindOne(ctx, filter).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, NewStateError("get_record", version, err)
	}

	return &record, nil
}

// MarkApplied records that a migration was successfully applied
func (s *MongoStateStore) MarkApplied(ctx context.Context, record *MigrationRecord) error {
	record.Status = StatusApplied
	record.Hostname = s.hostname

	opts := options.UpdateOne().SetUpsert(true)
	filter := bson.M{"version": record.Version}
	update := bson.M{"$set": record}

	_, err := s.coll().UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return NewStateError("mark_applied", record.Version, err)
	}

	return nil
}

// MarkFailed records that a migration failed
func (s *MongoStateStore) MarkFailed(ctx context.Context, version int64, name string, migrationErr error, durationMs int64) error {
	var errorMsg string
	if migrationErr != nil {
		errorMsg = migrationErr.Error()
	}

	record := &MigrationRecord{
		Version:    version,
		Name:       name,
		Status:     StatusFailed,
		AppliedAt:  time.Now(),
		DurationMs: durationMs,
		Hostname:   s.hostname,
		Error:      errorMsg,
	}

	opts := options.UpdateOne().SetUpsert(true)
	filter := bson.M{"version": version}
	update := bson.M{"$set": record}

	_, err := s.coll().UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return NewStateError("mark_failed", version, err)
	}

	return nil
}

// MarkRolledBack records that a migration was rolled back by removing its record
func (s *MongoStateStore) MarkRolledBack(ctx context.Context, version int64) error {
	_, err := s.coll().DeleteOne(ctx, bson.M{"version": version})
	if err != nil {
		return NewStateError("mark_rolled_back", version, err)
	}
	return nil
}

// GetNextBatch returns the next batch number
func (s *MongoStateStore) GetNextBatch(ctx context.Context) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "batch", Value: -1}})

	var record MigrationRecord
	err := s.coll().FindOne(ctx, bson.M{}, opts).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 1, nil
		}
		return 0, NewStateError("get_next_batch", 0, err)
	}

	return record.Batch + 1, nil
}

// GetDirty returns the dirty migration record if the database is in a dirty state
func (s *MongoStateStore) GetDirty(ctx context.Context) (*MigrationRecord, error) {
	filter := bson.M{
		"status": bson.M{"$in": []MigrationStatus{StatusFailed, StatusRollingBack}},
	}

	var record MigrationRecord
	err := s.coll().FindOne(ctx, filter).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, NewStateError("get_dirty", 0, err)
	}

	return &record, nil
}

// ClearDirty clears the dirty state for a specific version
func (s *MongoStateStore) ClearDirty(ctx context.Context, version int64) error {
	_, err := s.coll().DeleteOne(ctx, bson.M{
		"version": version,
		"status":  bson.M{"$in": []MigrationStatus{StatusFailed, StatusRollingBack}},
	})
	if err != nil {
		return NewStateError("clear_dirty", version, err)
	}

	return nil
}

// GetAppliedVersions returns a set of applied migration versions for quick lookup
func (s *MongoStateStore) GetAppliedVersions(ctx context.Context) (map[int64]bool, error) {
	records, err := s.GetApplied(ctx)
	if err != nil {
		return nil, err
	}

	versions := make(map[int64]bool, len(records))
	for _, r := range records {
		versions[r.Version] = true
	}

	return versions, nil
}

// VerifyChecksum verifies a migration's checksum against the stored value
func (s *MongoStateStore) VerifyChecksum(ctx context.Context, version int64, checksum string) error {
	record, err := s.GetRecord(ctx, version)
	if err != nil {
		return err
	}

	if record == nil {
		// Migration not applied yet, no checksum to verify
		return nil
	}

	if record.Checksum != "" && record.Checksum != checksum {
		return NewChecksumError(version, record.Name, record.Checksum, checksum)
	}

	return nil
}

// GetLastBatchMigrations returns all migrations from the last batch
func (s *MongoStateStore) GetLastBatchMigrations(ctx context.Context) ([]MigrationRecord, error) {
	// Get the last batch number
	opts := options.FindOne().SetSort(bson.D{{Key: "batch", Value: -1}})
	var lastRecord MigrationRecord
	err := s.coll().FindOne(ctx, bson.M{"status": StatusApplied}, opts).Decode(&lastRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, NewStateError("get_last_batch", 0, err)
	}

	// Get all migrations from that batch
	filter := bson.M{
		"status": StatusApplied,
		"batch":  lastRecord.Batch,
	}
	findOpts := options.Find().SetSort(bson.D{{Key: "version", Value: -1}})

	cursor, err := s.coll().Find(ctx, filter, findOpts)
	if err != nil {
		return nil, NewStateError("get_last_batch", 0, err)
	}
	defer cursor.Close(ctx)

	var records []MigrationRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, NewStateError("get_last_batch", 0, err)
	}

	return records, nil
}

// String returns a string representation of the state store
func (s *MongoStateStore) String() string {
	return fmt.Sprintf("MongoStateStore{collection=%s, hostname=%s}", s.collection, s.hostname)
}

package migrate

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// IndexOption configures index creation
type IndexOption func(*indexConfig)

type indexConfig struct {
	unique        bool
	sparse        bool
	expireAfter   *time.Duration
	name          string
	partialFilter bson.D
	noVersionBump bool
	schemaVersion int64
	collation     *options.Collation
	background    bool
}

// Unique creates a unique index
func Unique() IndexOption {
	return func(c *indexConfig) {
		c.unique = true
	}
}

// Sparse creates a sparse index
func Sparse() IndexOption {
	return func(c *indexConfig) {
		c.sparse = true
	}
}

// ExpireAfter creates a TTL index
func ExpireAfter(d time.Duration) IndexOption {
	return func(c *indexConfig) {
		c.expireAfter = &d
	}
}

// Name sets a custom index name
func Name(n string) IndexOption {
	return func(c *indexConfig) {
		c.name = n
	}
}

// PartialFilter sets a partial filter expression for the index
func PartialFilter(filter bson.D) IndexOption {
	return func(c *indexConfig) {
		c.partialFilter = filter
	}
}

// NoVersionBump prevents automatic _goe_sv update for schema-changing operations
func NoVersionBump() IndexOption {
	return func(c *indexConfig) {
		c.noVersionBump = true
	}
}

// WithSchemaVersion explicitly sets the schema version for this migration
func WithSchemaVersion(v int64) IndexOption {
	return func(c *indexConfig) {
		c.schemaVersion = v
	}
}

// WithCollation sets the collation for the index
func WithCollation(collation *options.Collation) IndexOption {
	return func(c *indexConfig) {
		c.collation = collation
	}
}

// Background creates the index in the background (deprecated in MongoDB 4.2+, kept for compatibility)
func Background() IndexOption {
	return func(c *indexConfig) {
		c.background = true
	}
}

// buildIndexConfig builds an indexConfig from options
func buildIndexConfig(opts []IndexOption) *indexConfig {
	cfg := &indexConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// --- Index Operations ---

// CreateIndex returns a MigrationFunc that creates a single-field index
func CreateIndex(collection, field string, opts ...IndexOption) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		cfg := buildIndexConfig(opts)

		indexOpts := options.Index()
		if cfg.unique {
			indexOpts.SetUnique(true)
		}
		if cfg.sparse {
			indexOpts.SetSparse(true)
		}
		if cfg.expireAfter != nil {
			indexOpts.SetExpireAfterSeconds(int32(cfg.expireAfter.Seconds()))
		}
		if cfg.name != "" {
			indexOpts.SetName(cfg.name)
		}
		if cfg.partialFilter != nil {
			indexOpts.SetPartialFilterExpression(cfg.partialFilter)
		}

		indexModel := mongo.IndexModel{
			Keys:    bson.D{{Key: field, Value: 1}},
			Options: indexOpts,
		}

		_, err := db.Collection(collection).Indexes().CreateOne(ctx, indexModel)
		return err
	}
}

// CreateCompoundIndex returns a MigrationFunc that creates a compound index
func CreateCompoundIndex(collection string, fields []string, opts ...IndexOption) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		cfg := buildIndexConfig(opts)

		indexOpts := options.Index()
		if cfg.unique {
			indexOpts.SetUnique(true)
		}
		if cfg.sparse {
			indexOpts.SetSparse(true)
		}
		if cfg.name != "" {
			indexOpts.SetName(cfg.name)
		}
		if cfg.partialFilter != nil {
			indexOpts.SetPartialFilterExpression(cfg.partialFilter)
		}

		keys := make(bson.D, 0, len(fields))
		for _, field := range fields {
			keys = append(keys, bson.E{Key: field, Value: 1})
		}

		indexModel := mongo.IndexModel{
			Keys:    keys,
			Options: indexOpts,
		}

		_, err := db.Collection(collection).Indexes().CreateOne(ctx, indexModel)
		return err
	}
}

// CreateUniqueIndex is a shorthand for CreateIndex with Unique option
func CreateUniqueIndex(collection, field string, opts ...IndexOption) MigrationFunc {
	return CreateIndex(collection, field, append([]IndexOption{Unique()}, opts...)...)
}

// CreateTTLIndex returns a MigrationFunc that creates a TTL index
func CreateTTLIndex(collection, field string, expireAfter time.Duration, opts ...IndexOption) MigrationFunc {
	return CreateIndex(collection, field, append([]IndexOption{ExpireAfter(expireAfter)}, opts...)...)
}

// DropIndex returns a MigrationFunc that drops an index by name
func DropIndex(collection, indexName string) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		err := db.Collection(collection).Indexes().DropOne(ctx, indexName)
		if err != nil {
			// Ignore "index not found" errors for idempotency
			if mongo.IsNetworkError(err) {
				return err
			}
		}
		return nil
	}
}

// --- Collection Operations ---

// CreateCollection returns a MigrationFunc that creates a collection
func CreateCollection(name string, opts ...*options.CreateCollectionOptionsBuilder) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		var createOpts *options.CreateCollectionOptionsBuilder
		if len(opts) > 0 {
			createOpts = opts[0]
		}
		return db.CreateCollection(ctx, name, createOpts)
	}
}

// DropCollection returns a MigrationFunc that drops a collection
func DropCollection(name string) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		return db.Collection(name).Drop(ctx)
	}
}

// RenameCollection returns a MigrationFunc that renames a collection
func RenameCollection(oldName, newName string) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		// MongoDB rename is done via admin command
		result := db.RunCommand(ctx, bson.D{
			{Key: "renameCollection", Value: db.Name() + "." + oldName},
			{Key: "to", Value: db.Name() + "." + newName},
		})
		return result.Err()
	}
}

// SetValidation returns a MigrationFunc that sets JSON schema validation on a collection
func SetValidation(collection string, schema bson.M) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		result := db.RunCommand(ctx, bson.D{
			{Key: "collMod", Value: collection},
			{Key: "validator", Value: bson.M{"$jsonSchema": schema}},
		})
		return result.Err()
	}
}

// RemoveValidation returns a MigrationFunc that removes validation from a collection
func RemoveValidation(collection string) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		result := db.RunCommand(ctx, bson.D{
			{Key: "collMod", Value: collection},
			{Key: "validator", Value: bson.M{}},
		})
		return result.Err()
	}
}

// --- Field Operations (with automatic _goe_sv update) ---

// fieldOperationConfig holds configuration for field operations
type fieldOperationConfig struct {
	noVersionBump bool
	schemaVersion int64
	filter        bson.M
}

// FieldOption configures field operations
type FieldOption func(*fieldOperationConfig)

// WithFilter restricts the field operation to documents matching the filter
func WithFilter(filter bson.M) FieldOption {
	return func(c *fieldOperationConfig) {
		c.filter = filter
	}
}

// buildFieldConfig builds a fieldOperationConfig from IndexOptions and FieldOptions
func buildFieldConfig(indexOpts []IndexOption, fieldOpts []FieldOption) *fieldOperationConfig {
	cfg := &fieldOperationConfig{}

	// Process IndexOptions for NoVersionBump and SchemaVersion
	for _, opt := range indexOpts {
		iCfg := &indexConfig{}
		opt(iCfg)
		if iCfg.noVersionBump {
			cfg.noVersionBump = true
		}
		if iCfg.schemaVersion > 0 {
			cfg.schemaVersion = iCfg.schemaVersion
		}
	}

	// Process FieldOptions
	for _, opt := range fieldOpts {
		opt(cfg)
	}

	return cfg
}

// AddField returns a MigrationFunc that adds a field with a default value
// By default, this also sets _goe_sv to track the document's schema version
func AddField(collection, field string, defaultValue any, opts ...IndexOption) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		cfg := buildIndexConfig(opts)

		filter := bson.M{field: bson.M{"$exists": false}}
		update := bson.M{"$set": bson.M{field: defaultValue}}

		// Add schema version update unless explicitly disabled
		if !cfg.noVersionBump && cfg.schemaVersion > 0 {
			update["$set"].(bson.M)["_goe_sv"] = cfg.schemaVersion
		}

		_, err := db.Collection(collection).UpdateMany(ctx, filter, update)
		return err
	}
}

// AddFieldWithVersion is AddField that explicitly sets the schema version
func AddFieldWithVersion(collection, field string, defaultValue any, schemaVersion int64, opts ...IndexOption) MigrationFunc {
	return AddField(collection, field, defaultValue, append(opts, WithSchemaVersion(schemaVersion))...)
}

// RemoveField returns a MigrationFunc that removes a field from all documents
func RemoveField(collection, field string, opts ...IndexOption) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		cfg := buildIndexConfig(opts)

		filter := bson.M{field: bson.M{"$exists": true}}
		update := bson.M{"$unset": bson.M{field: ""}}

		// Add schema version update unless explicitly disabled
		if !cfg.noVersionBump && cfg.schemaVersion > 0 {
			update["$set"] = bson.M{"_goe_sv": cfg.schemaVersion}
		}

		_, err := db.Collection(collection).UpdateMany(ctx, filter, update)
		return err
	}
}

// RenameField returns a MigrationFunc that renames a field in all documents
func RenameField(collection, oldName, newName string, opts ...IndexOption) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		cfg := buildIndexConfig(opts)

		filter := bson.M{oldName: bson.M{"$exists": true}}
		update := bson.M{"$rename": bson.M{oldName: newName}}

		// Add schema version update unless explicitly disabled
		if !cfg.noVersionBump && cfg.schemaVersion > 0 {
			update["$set"] = bson.M{"_goe_sv": cfg.schemaVersion}
		}

		_, err := db.Collection(collection).UpdateMany(ctx, filter, update)
		return err
	}
}

// ConvertFieldType returns a MigrationFunc that converts a field's type
// The transformer function receives the old value and returns the new value
func ConvertFieldType(collection, field string, transformer func(any) any, opts ...IndexOption) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		cfg := buildIndexConfig(opts)
		coll := db.Collection(collection)

		// Find all documents with the field
		cursor, err := coll.Find(ctx, bson.M{field: bson.M{"$exists": true}})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		// Process each document
		for cursor.Next(ctx) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				return err
			}

			oldValue := doc[field]
			newValue := transformer(oldValue)

			update := bson.M{"$set": bson.M{field: newValue}}
			if !cfg.noVersionBump && cfg.schemaVersion > 0 {
				update["$set"].(bson.M)["_goe_sv"] = cfg.schemaVersion
			}

			_, err := coll.UpdateOne(ctx, bson.M{"_id": doc["_id"]}, update)
			if err != nil {
				return err
			}
		}

		return cursor.Err()
	}
}

// --- Schema Versioning Operations ---

// SetSchemaVersion returns a MigrationFunc that sets _goe_sv on matching documents
func SetSchemaVersion(collection string, version int64, filter bson.M) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		if filter == nil {
			filter = bson.M{}
		}
		update := bson.M{"$set": bson.M{"_goe_sv": version}}
		_, err := db.Collection(collection).UpdateMany(ctx, filter, update)
		return err
	}
}

// BumpSchemaVersion returns a MigrationFunc that transforms documents from one schema version to another
func BumpSchemaVersion(collection string, fromVersion, toVersion int64, transformer func(bson.M) bson.M) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collection)

		// Find documents at the old schema version
		filter := bson.M{"_goe_sv": fromVersion}
		cursor, err := coll.Find(ctx, filter)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		// Process each document
		for cursor.Next(ctx) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				return err
			}

			// Apply transformation
			transformed := transformer(doc)
			transformed["_goe_sv"] = toVersion

			// Replace the document
			_, err := coll.ReplaceOne(ctx, bson.M{"_id": doc["_id"]}, transformed)
			if err != nil {
				return err
			}
		}

		return cursor.Err()
	}
}

// --- Utility Operations ---

// Sequence returns a MigrationFunc that runs multiple operations in sequence
func Sequence(fns ...MigrationFunc) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		for _, fn := range fns {
			if err := fn(ctx, db); err != nil {
				return err
			}
		}
		return nil
	}
}

// NoOp returns a MigrationFunc that does nothing
// Useful for migrations that only need an up or down operation
func NoOp() MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		return nil
	}
}

// ProgressCallback is called during bulk update operations to report progress
type ProgressCallback func(processed, total int64)

// UpdateManyWithProgress returns a MigrationFunc that performs a bulk update with progress reporting
func UpdateManyWithProgress(collection string, filter, update bson.M, callback ProgressCallback) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		coll := db.Collection(collection)

		// Count total documents
		total, err := coll.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}

		if total == 0 {
			return nil
		}

		// Process in batches
		const batchSize = 1000
		var processed int64

		cursor, err := coll.Find(ctx, filter)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		var batch []any
		for cursor.Next(ctx) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				return err
			}

			batch = append(batch, doc["_id"])

			if len(batch) >= batchSize {
				_, err := coll.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": batch}}, update)
				if err != nil {
					return err
				}
				processed += int64(len(batch))
				if callback != nil {
					callback(processed, total)
				}
				batch = batch[:0]
			}
		}

		// Process remaining
		if len(batch) > 0 {
			_, err := coll.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": batch}}, update)
			if err != nil {
				return err
			}
			processed += int64(len(batch))
			if callback != nil {
				callback(processed, total)
			}
		}

		return cursor.Err()
	}
}

// RunCommand returns a MigrationFunc that runs a MongoDB command
func RunCommand(command bson.D) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		result := db.RunCommand(ctx, command)
		return result.Err()
	}
}

// InsertDocuments returns a MigrationFunc that inserts documents
func InsertDocuments(collection string, docs []any) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		if len(docs) == 0 {
			return nil
		}
		_, err := db.Collection(collection).InsertMany(ctx, docs)
		return err
	}
}

// DeleteDocuments returns a MigrationFunc that deletes documents matching a filter
func DeleteDocuments(collection string, filter bson.M) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(collection).DeleteMany(ctx, filter)
		return err
	}
}

// CreateTextIndex returns a MigrationFunc that creates a text index
func CreateTextIndex(collection string, fields []string, opts ...IndexOption) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		cfg := buildIndexConfig(opts)

		indexOpts := options.Index()
		if cfg.name != "" {
			indexOpts.SetName(cfg.name)
		}

		keys := make(bson.D, 0, len(fields))
		for _, field := range fields {
			keys = append(keys, bson.E{Key: field, Value: "text"})
		}

		indexModel := mongo.IndexModel{
			Keys:    keys,
			Options: indexOpts,
		}

		_, err := db.Collection(collection).Indexes().CreateOne(ctx, indexModel)
		return err
	}
}

// Create2DSphereIndex returns a MigrationFunc that creates a 2dsphere geospatial index
func Create2DSphereIndex(collection, field string, opts ...IndexOption) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		cfg := buildIndexConfig(opts)

		indexOpts := options.Index()
		if cfg.name != "" {
			indexOpts.SetName(cfg.name)
		}

		indexModel := mongo.IndexModel{
			Keys:    bson.D{{Key: field, Value: "2dsphere"}},
			Options: indexOpts,
		}

		_, err := db.Collection(collection).Indexes().CreateOne(ctx, indexModel)
		return err
	}
}

// EnsureIndex creates an index only if it doesn't already exist
func EnsureIndex(collection, field string, opts ...IndexOption) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		cfg := buildIndexConfig(opts)

		indexOpts := options.Index()
		if cfg.unique {
			indexOpts.SetUnique(true)
		}
		if cfg.sparse {
			indexOpts.SetSparse(true)
		}
		if cfg.expireAfter != nil {
			indexOpts.SetExpireAfterSeconds(int32(cfg.expireAfter.Seconds()))
		}
		if cfg.name != "" {
			indexOpts.SetName(cfg.name)
		}

		indexModel := mongo.IndexModel{
			Keys:    bson.D{{Key: field, Value: 1}},
			Options: indexOpts,
		}

		// CreateOne is idempotent - it will return the existing index name if it already exists
		_, err := db.Collection(collection).Indexes().CreateOne(ctx, indexModel)
		return err
	}
}

// ValidateCollection returns a MigrationFunc that validates a collection
func ValidateCollection(collection string) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		result := db.RunCommand(ctx, bson.D{
			{Key: "validate", Value: collection},
		})
		return result.Err()
	}
}

// CustomFunc allows wrapping any custom function as a MigrationFunc
// This is useful for complex migrations that don't fit the helper pattern
func CustomFunc(fn func(ctx context.Context, db *mongo.Database) error) MigrationFunc {
	return fn
}

// Log returns a MigrationFunc that logs a message (useful for debugging migrations)
func Log(message string) MigrationFunc {
	return func(ctx context.Context, db *mongo.Database) error {
		fmt.Println("[migrate]", message)
		return nil
	}
}

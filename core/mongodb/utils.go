package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// BuildIndexModel creates a simple index model
func BuildIndexModel(keys bson.D) mongo.IndexModel {
	return mongo.IndexModel{
		Keys: keys,
	}
}

// BuildUniqueIndex creates a unique index model with optional name
func BuildUniqueIndex(keys bson.D, name string) mongo.IndexModel {
	indexModel := mongo.IndexModel{
		Keys: keys,
	}

	if name != "" {
		indexModel.Options = options.Index().SetName(name).SetUnique(true)
	} else {
		indexModel.Options = options.Index().SetUnique(true)
	}

	return indexModel
}

// BuildTextIndex creates a text index model for full-text search
func BuildTextIndex(fields []string, name string) mongo.IndexModel {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: "text"})
	}

	indexModel := mongo.IndexModel{
		Keys: keys,
	}

	if name != "" {
		indexModel.Options = options.Index().SetName(name)
	}

	return indexModel
}

// BuildCompoundIndex creates a compound index model
func BuildCompoundIndex(fields map[string]int, name string) mongo.IndexModel {
	keys := bson.D{}
	for field, order := range fields {
		keys = append(keys, bson.E{Key: field, Value: order})
	}

	indexModel := mongo.IndexModel{
		Keys: keys,
	}

	if name != "" {
		indexModel.Options = options.Index().SetName(name)
	}

	return indexModel
}

// BuildTTLIndex creates a TTL (Time To Live) index model
func BuildTTLIndex(field string, expireAfter time.Duration, name string) mongo.IndexModel {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: field, Value: 1}},
	}

	if name != "" {
		indexModel.Options = options.Index().SetName(name).SetExpireAfterSeconds(int32(expireAfter.Seconds()))
	} else {
		indexModel.Options = options.Index().SetExpireAfterSeconds(int32(expireAfter.Seconds()))
	}

	return indexModel
}

// Context utilities

// NewContext creates a new context with the specified timeout
func NewContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// DefaultContext creates a new context with a default 10-second timeout
func DefaultContext() (context.Context, context.CancelFunc) {
	return NewContext(10 * time.Second)
}

// Error handling utilities

// IsNoDocumentsError checks if the error is a "no documents" error
func IsNoDocumentsError(err error) bool {
	return err != nil && err == mongo.ErrNoDocuments
}

// IsDuplicateKeyError checks if the error is a duplicate key error
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	if writeException, ok := err.(mongo.WriteException); ok {
		for _, writeError := range writeException.WriteErrors {
			if writeError.Code == 11000 {
				return true
			}
		}
	}

	return false
}

// IsNetworkError checks if the error is a network-related error
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	return mongo.IsNetworkError(err)
}

// IsTimeout checks if the error is a timeout error
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	// Check for context deadline exceeded or timeout
	return err == context.DeadlineExceeded
}

// Transaction utilities

// WithTransaction executes a function within a transaction
func WithTransaction(ctx context.Context, client *mongo.Client, fn func(sessCtx context.Context) error) error {
	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	if err := session.StartTransaction(); err != nil {
		return err
	}

	sessionCtx := mongo.NewSessionContext(ctx, session)
	if err := fn(sessionCtx); err != nil {
		if abortErr := session.AbortTransaction(sessionCtx); abortErr != nil {
			return fmt.Errorf("transaction failed: %w, abort error: %v", err, abortErr)
		}
		return err
	}

	return session.CommitTransaction(sessionCtx)
}

// WithTransactionOptions executes a function within a transaction with options.
// Options should be built using the v2 builder pattern, e.g.:
//
//	opts := options.Transaction().SetReadConcern(readconcern.Majority())
//	err := WithTransactionOptions(ctx, client, fn, opts)
func WithTransactionOptions(ctx context.Context, client *mongo.Client, fn func(sessCtx context.Context) error, opts ...options.Lister[options.TransactionOptions]) error {
	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	if err := session.StartTransaction(opts...); err != nil {
		return err
	}

	sessionCtx := mongo.NewSessionContext(ctx, session)
	if err := fn(sessionCtx); err != nil {
		if abortErr := session.AbortTransaction(sessionCtx); abortErr != nil {
			return fmt.Errorf("transaction failed: %w, abort error: %v", err, abortErr)
		}
		return err
	}

	return session.CommitTransaction(sessionCtx)
}

// Index management utilities

// EnsureIndexes creates indexes if they don't already exist
func EnsureIndexes(ctx context.Context, collection *mongo.Collection, indexes []mongo.IndexModel) error {
	if len(indexes) == 0 {
		return nil
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// DropIndex drops an index by name
func DropIndex(ctx context.Context, collection *mongo.Collection, indexName string) error {
	return collection.Indexes().DropOne(ctx, indexName)
}

// Bulk operation utilities

// CreateInsertOneModel creates an insert one model for bulk operations
func CreateInsertOneModel(document any) mongo.WriteModel {
	return mongo.NewInsertOneModel().SetDocument(document)
}

// CreateUpdateOneModel creates an update one model for bulk operations
func CreateUpdateOneModel(filter, update any) mongo.WriteModel {
	return mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
}

// CreateUpsertModel creates an upsert model (replace one with upsert) for bulk operations
func CreateUpsertModel(filter, replacement any) mongo.WriteModel {
	return mongo.NewReplaceOneModel().SetFilter(filter).SetReplacement(replacement).SetUpsert(true)
}

// CreateDeleteOneModel creates a delete one model for bulk operations
func CreateDeleteOneModel(filter any) mongo.WriteModel {
	return mongo.NewDeleteOneModel().SetFilter(filter)
}

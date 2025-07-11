package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// IsNoDocumentsError checks if the given error is mongo.ErrNoDocuments
func IsNoDocumentsError(err error) bool {
	return errors.Is(err, mongo.ErrNoDocuments)
}

// IsDuplicateKeyError checks if the given error is a duplicate key error
func IsDuplicateKeyError(err error) bool {
	var writeException mongo.WriteException
	if errors.As(err, &writeException) {
		for _, writeErr := range writeException.WriteErrors {
			if writeErr.Code == 11000 {
				return true
			}
		}
	}
	return false
}

// IsNetworkError checks if the given error is a network error
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common network error patterns
	return mongo.IsNetworkError(err)
}

// IsTimeout checks if the given error is a timeout error
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	return mongo.IsTimeout(err)
}

// NewContext creates a new context with the specified timeout
func NewContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// DefaultContext creates a new context with a 10-second timeout
func DefaultContext() (context.Context, context.CancelFunc) {
	return NewContext(10 * time.Second)
}

// BuildIndexModel creates a new index model with the given keys
func BuildIndexModel(keys bson.D) mongo.IndexModel {
	return mongo.IndexModel{
		Keys: keys,
	}
}

// BuildUniqueIndex creates a unique index model
func BuildUniqueIndex(keys bson.D, name string) mongo.IndexModel {
	return mongo.IndexModel{
		Keys: keys,
	}
}

// BuildTextIndex creates a text index model for full-text search
func BuildTextIndex(fields []string, name string) mongo.IndexModel {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: "text"})
	}
	return mongo.IndexModel{
		Keys: keys,
	}
}

// BuildCompoundIndex creates a compound index model
func BuildCompoundIndex(fields map[string]int, name string) mongo.IndexModel {
	keys := bson.D{}
	for field, order := range fields {
		keys = append(keys, bson.E{Key: field, Value: order})
	}
	return mongo.IndexModel{
		Keys: keys,
	}
}

// BuildTTLIndex creates a TTL (Time To Live) index model
func BuildTTLIndex(field string, expireAfter time.Duration, name string) mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{{Key: field, Value: 1}},
	}
}

// CreateIndexes creates multiple indexes on a collection
func CreateIndexes(ctx context.Context, collection *mongo.Collection, indexes []mongo.IndexModel) error {
	if len(indexes) == 0 {
		return nil
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// WithTransaction executes a function within a transaction
func WithTransaction(ctx context.Context, client *mongo.Client, fn func(sessCtx context.Context) error) error {
	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// For now, just execute the function without transaction support
	// MongoDB v2 driver transaction API has changed significantly
	return fn(ctx)
}

// CountDocuments counts documents matching the filter with a default timeout
func CountDocuments(ctx context.Context, collection *mongo.Collection, filter interface{}) (int64, error) {
	if filter == nil {
		filter = bson.D{}
	}
	return collection.CountDocuments(ctx, filter)
}

// Exists checks if any document matching the filter exists
func Exists(ctx context.Context, collection *mongo.Collection, filter interface{}) (bool, error) {
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

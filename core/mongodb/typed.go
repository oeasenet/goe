package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Repository provides a generic repository pattern for MongoDB collections
type Repository[T any] struct {
	collection *mongo.Collection
}

// NewRepository creates a new typed repository for a collection
func NewRepository[T any](db *mongo.Database, collectionName string) *Repository[T] {
	return &Repository[T]{
		collection: db.Collection(collectionName),
	}
}

// Collection returns the underlying MongoDB collection
func (r *Repository[T]) Collection() *mongo.Collection {
	return r.collection
}

// FindOne finds a single document
func (r *Repository[T]) FindOne(ctx context.Context, filter interface{}) (*T, error) {
	var result T
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Find finds multiple documents
func (r *Repository[T]) Find(ctx context.Context, filter interface{}) ([]*T, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*T
	for cursor.Next(ctx) {
		var item T
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		results = append(results, &item)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// FindByID finds a document by its _id
func (r *Repository[T]) FindByID(ctx context.Context, id interface{}) (*T, error) {
	return r.FindOne(ctx, bson.M{"_id": id})
}

// Insert inserts a new document
func (r *Repository[T]) Insert(ctx context.Context, document *T) (*mongo.InsertOneResult, error) {
	return r.collection.InsertOne(ctx, document)
}

// InsertMany inserts multiple documents
func (r *Repository[T]) InsertMany(ctx context.Context, documents []*T) (*mongo.InsertManyResult, error) {
	docs := make([]interface{}, len(documents))
	for i, doc := range documents {
		docs[i] = doc
	}
	return r.collection.InsertMany(ctx, docs)
}

// Update updates documents matching the filter
func (r *Repository[T]) Update(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return r.collection.UpdateMany(ctx, filter, update)
}

// UpdateOne updates a single document
func (r *Repository[T]) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return r.collection.UpdateOne(ctx, filter, update)
}

// UpdateByID updates a document by its _id
func (r *Repository[T]) UpdateByID(ctx context.Context, id interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return r.UpdateOne(ctx, bson.M{"_id": id}, update)
}

// Replace replaces a document
func (r *Repository[T]) Replace(ctx context.Context, filter interface{}, replacement *T) (*mongo.UpdateResult, error) {
	return r.collection.ReplaceOne(ctx, filter, replacement)
}

// Delete deletes documents matching the filter
func (r *Repository[T]) Delete(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return r.collection.DeleteMany(ctx, filter)
}

// DeleteOne deletes a single document
func (r *Repository[T]) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return r.collection.DeleteOne(ctx, filter)
}

// DeleteByID deletes a document by its _id
func (r *Repository[T]) DeleteByID(ctx context.Context, id interface{}) (*mongo.DeleteResult, error) {
	return r.DeleteOne(ctx, bson.M{"_id": id})
}

// Count counts documents matching the filter
func (r *Repository[T]) Count(ctx context.Context, filter interface{}) (int64, error) {
	if filter == nil {
		filter = bson.D{}
	}
	return r.collection.CountDocuments(ctx, filter)
}

// Exists checks if any document matching the filter exists
func (r *Repository[T]) Exists(ctx context.Context, filter interface{}) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Aggregate performs an aggregation
func (r *Repository[T]) Aggregate(ctx context.Context, pipeline interface{}) ([]*T, error) {
	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// CreateIndex creates a single index
func (r *Repository[T]) CreateIndex(ctx context.Context, model mongo.IndexModel) (string, error) {
	return r.collection.Indexes().CreateOne(ctx, model)
}

// CreateIndexes creates multiple indexes
func (r *Repository[T]) CreateIndexes(ctx context.Context, models []mongo.IndexModel) ([]string, error) {
	return r.collection.Indexes().CreateMany(ctx, models)
}

// DropIndex drops an index by name
func (r *Repository[T]) DropIndex(ctx context.Context, name string) error {
	return r.collection.Indexes().DropOne(ctx, name)
}

// BulkWrite performs bulk write operations
func (r *Repository[T]) BulkWrite(ctx context.Context, models []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	return r.collection.BulkWrite(ctx, models)
}

// PaginatedResult represents paginated results
type PaginatedResult[T any] struct {
	Items       []*T  `json:"items"`
	Total       int64 `json:"total"`
	Page        int64 `json:"page"`
	PageSize    int64 `json:"page_size"`
	TotalPages  int64 `json:"total_pages"`
	HasPrevious bool  `json:"has_previous"`
	HasNext     bool  `json:"has_next"`
}

// Paginate returns paginated results
func (r *Repository[T]) Paginate(ctx context.Context, filter interface{}, page, pageSize int64) (*PaginatedResult[T], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Count total documents
	total, err := r.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	// Calculate pagination
	skip := (page - 1) * pageSize
	totalPages := (total + pageSize - 1) / pageSize

	// Find documents with pagination
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	// Skip to the right page
	for i := int64(0); i < skip && cursor.Next(ctx); i++ {
		// Skip documents
	}

	// Collect page results
	var items []*T
	count := int64(0)
	for cursor.Next(ctx) && count < pageSize {
		var item T
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		items = append(items, &item)
		count++
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return &PaginatedResult[T]{
		Items:       items,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasPrevious: page > 1,
		HasNext:     page < totalPages,
	}, nil
}

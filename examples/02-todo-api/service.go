package main

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.oease.dev/goe/v2/contract"
	goecache "go.oease.dev/goe/v2/core/cache"
)

const (
	collectionName = "todos"
	cachePrefix    = "todo:"
	cacheTTL       = 5 * time.Minute
)

// TodoService handles business logic for todo operations.
type TodoService struct {
	db     contract.MongoDB
	cache  contract.Cache
	logger contract.Logger
}

// NewTodoService creates a new TodoService with injected dependencies.
func NewTodoService(db contract.MongoDB, cache contract.Cache, logger contract.Logger) *TodoService {
	return &TodoService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// collection returns the todos MongoDB collection.
func (s *TodoService) collection() *mongo.Collection {
	return s.db.Col(collectionName)
}

// cacheKey generates a cache key for a todo by ID.
func (s *TodoService) cacheKey(id string) string {
	return cachePrefix + id
}

// Create creates a new todo item.
func (s *TodoService) Create(ctx context.Context, req *CreateTodoRequest) (*Todo, error) {
	todo := &Todo{
		ID:          bson.NewObjectID(),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := s.collection().InsertOne(ctx, todo)
	if err != nil {
		s.logger.Errorw("Failed to create todo", "error", err)
		return nil, err
	}

	s.logger.Infow("Todo created", "id", todo.ID.Hex())

	// Invalidate stats cache since count changed
	_ = s.cache.Forget(cachePrefix + "stats")

	return todo, nil
}

// Get retrieves a todo by ID, using cache when available.
func (s *TodoService) Get(ctx context.Context, id string) (*Todo, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid todo ID")
	}

	// Try cache first using Remember pattern
	var todo Todo
	err = s.cache.Remember(s.cacheKey(id), &todo, cacheTTL, func() (any, error) {
		s.logger.Debugw("Cache miss, fetching from database", "id", id)

		var t Todo
		err := s.collection().FindOne(ctx, bson.M{"_id": objID}).Decode(&t)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errors.New("todo not found")
			}
			return nil, err
		}
		return t, nil
	})

	if err != nil {
		return nil, err
	}

	return &todo, nil
}

// List returns all todos (with optional pagination in production).
func (s *TodoService) List(ctx context.Context) ([]Todo, error) {
	cursor, err := s.collection().Find(ctx, bson.M{})
	if err != nil {
		s.logger.Errorw("Failed to list todos", "error", err)
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var todos []Todo
	if err := cursor.All(ctx, &todos); err != nil {
		return nil, err
	}

	// Return empty slice instead of nil for consistent JSON
	if todos == nil {
		todos = []Todo{}
	}

	return todos, nil
}

// Update updates a todo by ID.
func (s *TodoService) Update(ctx context.Context, id string, req *UpdateTodoRequest) (*Todo, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid todo ID")
	}

	// Build update document with only provided fields
	update := bson.M{"updated_at": time.Now()}
	if req.Title != nil {
		update["title"] = *req.Title
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.Completed != nil {
		update["completed"] = *req.Completed
	}

	result := s.collection().FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
	)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, errors.New("todo not found")
		}
		return nil, result.Err()
	}

	// Invalidate caches
	_ = s.cache.Forget(s.cacheKey(id))
	_ = s.cache.Forget(cachePrefix + "stats")

	s.logger.Infow("Todo updated", "id", id)

	// Fetch updated document
	return s.Get(ctx, id)
}

// Delete removes a todo by ID.
func (s *TodoService) Delete(ctx context.Context, id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid todo ID")
	}

	result, err := s.collection().DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		s.logger.Errorw("Failed to delete todo", "error", err, "id", id)
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("todo not found")
	}

	// Invalidate caches
	_ = s.cache.Forget(s.cacheKey(id))
	_ = s.cache.Forget(cachePrefix + "stats")

	s.logger.Infow("Todo deleted", "id", id)
	return nil
}

// Stats returns aggregate statistics, cached for performance. The typed
// generic helper replaces the pointer-binding Remember call; the untyped
// method remains available and both styles share the same cache.
func (s *TodoService) Stats(ctx context.Context) (*TodoStats, error) {
	stats, err := goecache.Remember(s.cache, cachePrefix+"stats", cacheTTL, func() (TodoStats, error) {
		s.logger.Debug("Cache miss for stats, calculating from database")

		total, err := s.collection().CountDocuments(ctx, bson.M{})
		if err != nil {
			return TodoStats{}, err
		}

		completed, err := s.collection().CountDocuments(ctx, bson.M{"completed": true})
		if err != nil {
			return TodoStats{}, err
		}

		return TodoStats{
			Total:     total,
			Completed: completed,
			Pending:   total - completed,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

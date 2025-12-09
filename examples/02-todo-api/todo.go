package main

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Todo represents a todo item in the database.
type Todo struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string        `bson:"title" json:"title"`
	Description string        `bson:"description,omitempty" json:"description,omitempty"`
	Completed   bool          `bson:"completed" json:"completed"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}

// CreateTodoRequest represents the request body for creating a todo.
type CreateTodoRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=200"`
	Description string `json:"description,omitempty" validate:"max=1000"`
}

// UpdateTodoRequest represents the request body for updating a todo.
type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Completed   *bool   `json:"completed,omitempty"`
}

// TodoStats represents aggregate statistics about todos.
type TodoStats struct {
	Total     int64 `json:"total"`
	Completed int64 `json:"completed"`
	Pending   int64 `json:"pending"`
}

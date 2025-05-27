package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/goe/modules/mongodb"
	"go.oease.dev/goe/modules/rbac" // Import the rbac package for Permission type
	"time"
)

// UserDirectPermission stores permissions directly assigned to a user,
// independent of their roles.
type UserDirectPermission struct {
	mongodb.DefaultModel `bson:",inline"`
	UserID               primitive.ObjectID   `json:"user_id" bson:"user_id"`
	Permissions          []rbac.Permission    `json:"permissions" bson:"permissions"`
}

// ColName returns the collection name for UserDirectPermission.
func (m *UserDirectPermission) ColName() string {
	return "user_direct_permissions"
}

// BeforeInsert is a hook called before inserting a new UserDirectPermission.
func (m *UserDirectPermission) BeforeInsert() error {
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = time.Now().UTC()
	if m.ID.IsZero() {
		m.ID = primitive.NewObjectID()
	}
	if m.Permissions == nil {
		m.Permissions = []rbac.Permission{} // Initialize to empty slice
	}
	return nil
}

// BeforeUpdate is a hook called before updating a UserDirectPermission.
func (m *UserDirectPermission) BeforeUpdate() error {
	m.UpdatedAt = time.Now().UTC()
	if m.Permissions == nil {
		m.Permissions = []rbac.Permission{} // Initialize to empty slice
	}
	return nil
}

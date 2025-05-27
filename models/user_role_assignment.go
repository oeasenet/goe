package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/goe/modules/mongodb"
	"time"
)

// UserRoleAssignment links a user to a role.
// A user can have multiple roles.
type UserRoleAssignment struct {
	mongodb.DefaultModel `bson:",inline"`
	UserID               primitive.ObjectID `json:"user_id" bson:"user_id"`
	RoleName             string             `json:"role_name" bson:"role_name"` // References Role.Name
}

// ColName returns the collection name for UserRoleAssignment.
func (m *UserRoleAssignment) ColName() string {
	return "user_role_assignments"
}

// BeforeInsert is a hook called before inserting a new UserRoleAssignment.
func (m *UserRoleAssignment) BeforeInsert() error {
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = time.Now().UTC()
	if m.ID.IsZero() {
		m.ID = primitive.NewObjectID()
	}
	return nil
}

// BeforeUpdate is a hook called before updating a UserRoleAssignment.
func (m *UserRoleAssignment) BeforeUpdate() error {
	m.UpdatedAt = time.Now().UTC()
	return nil
}

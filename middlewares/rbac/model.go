package rbac

import (
	"go.oease.dev/goe/modules/mongodb"
)

var UserIDKey = "user_id"

// Permission is a string representing an action on a feature.
// Format: "feature:action" (e.g., "article:read", "user:create")
type Permission string

// Role defines a named set of permissions.
type Role struct {
	mongodb.DefaultModel `bson:",inline"`
	Name                 string       `json:"name" bson:"name"`
	Permissions          []Permission `json:"permissions" bson:"permissions"`
}

// ColName returns the collection name for UserDirectPermission.
func (m *Role) ColName() string {
	return "roles"
}

// MatchMode defines how permissions should be checked.
type MatchMode int

const (
	// MatchAll requires all specified permissions to be present.
	MatchAll MatchMode = iota
	// MatchAtLeastOne requires at least one of the specified permissions to be present.
	MatchAtLeastOne
)

// UserDirectPermission stores permissions directly assigned to a user,
// independent of their roles.
type UserDirectPermission struct {
	mongodb.DefaultModel `bson:",inline"`
	UserID               string       `json:"user_id" bson:"user_id"`
	Permissions          []Permission `json:"permissions" bson:"permissions"`
}

// ColName returns the collection name for UserDirectPermission.
func (m *UserDirectPermission) ColName() string {
	return "user_direct_permissions"
}

// UserRoleAssignment links a user to a role.
// A user can have multiple roles.
type UserRoleAssignment struct {
	mongodb.DefaultModel `bson:",inline"`
	UserID               string `json:"user_id" bson:"user_id"`
	RoleName             string `json:"role_name" bson:"role_name"` // References Role.Name
}

// ColName returns the collection name for UserRoleAssignment.
func (m *UserRoleAssignment) ColName() string {
	return "user_role_assignments"
}

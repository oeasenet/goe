package contracts

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/goe/modules/rbac" // For rbac.Permission, rbac.Role
)

// RBACManager defines the interface for role-based access control operations.
type RBACManager interface {
	// Role Definition
	DefineRole(role rbac.Role)
	GetRole(name string) (rbac.Role, bool)

	// Role Assignment
	AssignRoleToUser(userID primitive.ObjectID, roleName string) error
	RevokeRoleFromUser(userID primitive.ObjectID, roleName string) error
	GetUserRoleNames(userID primitive.ObjectID) ([]string, error)

	// Direct Permissions
	GrantDirectPermission(userID primitive.ObjectID, permission rbac.Permission) error
	RevokeDirectPermission(userID primitive.ObjectID, permission rbac.Permission) error
	GetUserDirectPermissions(userID primitive.ObjectID) ([]rbac.Permission, error)

	// Combined Permissions
	GetAllUserPermissions(userID primitive.ObjectID) ([]rbac.Permission, error)
}

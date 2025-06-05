package contracts

//
//import (
//	"go.oease.dev/goe/middlewares/rbac"
//)
//
//// RBACManager defines the interface for role-based access control operations.
//type RBACManager interface {
//	// Role Definition
//	DefineRole(role rbac.Role)
//	GetRole(name string) (rbac.Role, bool)
//
//	// Role Assignment
//	AssignRoleToUser(userID string, roleName string) error
//	RevokeRoleFromUser(userID string, roleName string) error
//	GetUserRoleNames(userID string) ([]string, error)
//
//	// Direct Permissions
//	GrantDirectPermission(userID string, permission rbac.Permission) error
//	RevokeDirectPermission(userID string, permission rbac.Permission) error
//	GetUserDirectPermissions(userID string) ([]rbac.Permission, error)
//
//	// Combined Permissions
//	GetAllUserPermissions(userID string) ([]rbac.Permission, error)
//}

package rbac

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/goe/contracts"
)

// rbacManagerImpl implements the contracts.RBACManager interface.
type rbacManagerImpl struct {
	// We don't need to store state here as the functions from rbac.go
	// (which this manager will call) operate on global 'definedRoles'
	// or directly use the database via the GoeContainer.
}

// NewRBACManager creates a new instance of RBACManager.
// This function will be the primary way other parts of the application
// get an RBACManager instance.
func NewRBACManager() contracts.RBACManager {
	return &rbacManagerImpl{}
}

// --- Role Definition ---
func (m *rbacManagerImpl) DefineRole(role Role) {
	DefineRole(role) // Calls the existing package-level function
}

func (m *rbacManagerImpl) GetRole(name string) (Role, bool) {
	return GetRole(name) // Calls the existing package-level function
}

// --- Role Assignment ---
func (m *rbacManagerImpl) AssignRoleToUser(userID primitive.ObjectID, roleName string) error {
	return AssignRoleToUser(userID, roleName) // Calls the existing package-level function
}

func (m *rbacManagerImpl) RevokeRoleFromUser(userID primitive.ObjectID, roleName string) error {
	return RevokeRoleFromUser(userID, roleName) // Calls the existing package-level function
}

func (m *rbacManagerImpl) GetUserRoleNames(userID primitive.ObjectID) ([]string, error) {
	return GetUserRoleNames(userID) // Calls the existing package-level function
}

// --- Direct Permissions ---
func (m *rbacManagerImpl) GrantDirectPermission(userID primitive.ObjectID, permission Permission) error {
	return GrantDirectPermission(userID, permission) // Calls the existing package-level function
}

func (m *rbacManagerImpl) RevokeDirectPermission(userID primitive.ObjectID, permission Permission) error {
	return RevokeDirectPermission(userID, permission) // Calls the existing package-level function
}

func (m *rbacManagerImpl) GetUserDirectPermissions(userID primitive.ObjectID) ([]Permission, error) {
	return GetUserDirectPermissions(userID) // Calls the existing package-level function
}

// --- Combined Permissions ---
func (m *rbacManagerImpl) GetAllUserPermissions(userID primitive.ObjectID) ([]Permission, error) {
	return GetAllUserPermissions(userID) // Calls the existing package-level function
}

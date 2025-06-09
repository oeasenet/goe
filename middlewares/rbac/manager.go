package rbac

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
)

// RBACMiddleware provides middleware for checking user permissions.
type RBACMiddleware struct {
	userIdGetter func(ctx fiber.Ctx) string
}

// NewRBACMiddleware creates a new RBACMiddleware instance.
func NewRBACMiddleware() *RBACMiddleware {
	return &RBACMiddleware{}
}

func (m *RBACMiddleware) SetUserIdGetter(f func(ctx fiber.Ctx) string) {
	m.userIdGetter = f
}

// DefineRole --- Role Definition ---
func (m *RBACMiddleware) DefineRole(role *Role) error {
	return DefineRole(role)
}

func (m *RBACMiddleware) GetRole(name string) (*Role, bool) {
	return GetRole(name) // Calls the existing package-level function
}

func (m *RBACMiddleware) DeleteRole(name string) error {
	return DeleteRole(name)
}

func (m *RBACMiddleware) ListRoles(pageSize, currentPage int64) ([]*Role, error) {
	return ListRoles(pageSize, currentPage)
}

// AssignRoleToUser --- Role Assignment ---
func (m *RBACMiddleware) AssignRoleToUser(userID string, roleName string) error {
	return AssignRoleToUser(userID, roleName) // Calls the existing package-level function
}

func (m *RBACMiddleware) RevokeRoleFromUser(userID string, roleName string) error {
	return RevokeRoleFromUser(userID, roleName) // Calls the existing package-level function
}

func (m *RBACMiddleware) GetUserRoleNames(userID string) ([]string, error) {
	return GetUserRoleNames(userID) // Calls the existing package-level function
}

// --- Direct Permissions ---

func (m *RBACMiddleware) GrantDirectPermission(userID string, permission Permission) error {
	return GrantDirectPermission(userID, permission) // Calls the existing package-level function
}

func (m *RBACMiddleware) RevokeDirectPermission(userID string, permission Permission) error {
	return RevokeDirectPermission(userID, permission) // Calls the existing package-level function
}

func (m *RBACMiddleware) GetUserDirectPermissions(userID string) ([]Permission, error) {
	return GetUserDirectPermissions(userID) // Calls the existing package-level function
}

// --- Combined Permissions ---

func (m *RBACMiddleware) GetAllUserPermissions(userID string) ([]Permission, error) {
	return GetAllUserPermissions(userID) // Calls the existing package-level function
}

// GetDefaultUserIdGetter retrieves the user ID from the context locals using the UserIDKey.
func GetDefaultUserIdGetter(ctx fiber.Ctx) string {
	fmt.Print(ctx.Locals(UserIDKey))
	if ctx.Locals(UserIDKey) == nil {
		return ""
	}
	return ctx.Locals(UserIDKey).(string)
}

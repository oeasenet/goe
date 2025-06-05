package rbac

import (
	"context"
	"errors"
	"fmt"
	"github.com/gookit/goutil/strutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.oease.dev/goe/modules/mongodb"
	"strings"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/core"
	// "go.oease.dev/goe/models"; // No longer needed directly here if rbac module handles models
	"go.oease.dev/goe/webresult"
)

// CheckPermission returns a fiber.Handler that checks if the current user has
// the required permissions.
//
// userIDKey is the key used to retrieve the user's ID from ctx.Locals().
// It's assumed that a prior middleware (e.g., auth middleware) has already
// authenticated the user and stored their ID (as string or primitive.ObjectID)
// in fiber.Ctx.Locals().
//
// requiredPermissions are the permissions needed to access the route.
//
// mode specifies whether all or at least one of the requiredPermissions are needed.
func (m *RBACMiddleware) CheckPermission(requiredPermissions []Permission, mode MatchMode) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userID := m.userIdGetter(ctx)
		if strutil.IsBlank(userID) {
			core.UseGoeContainer().GetLogger().Warn("RBAC: User ID is BLANK")
			return webresult.Forbidden("Access denied. User identifier is zero.")
		}

		// Use actual functions from rbac module
		userRoleNames, err := GetUserRoleNames(userID) // UPDATED
		if err != nil {
			core.UseGoeContainer().GetLogger().Errorf("RBAC: Error fetching user roles for user '%s': %v", userID, err)
			return webresult.SystemBusy(errors.New("error checking permissions"))
		}

		directPermissions, err := GetUserDirectPermissions(userID) // UPDATED
		if err != nil {
			core.UseGoeContainer().GetLogger().Errorf("RBAC: Error fetching direct user permissions for user '%s': %v", userID, err)
			return webresult.SystemBusy(errors.New("error checking permissions"))
		}

		currentUserPermissions := make(map[Permission]bool)

		// Add direct permissions
		for _, p := range directPermissions {
			currentUserPermissions[p] = true
		}

		// Add permissions from roles
		for _, roleName := range userRoleNames {
			role, found := GetRole(roleName) // Using GetRole from modules/rbac/rbac.go
			if found {
				for _, p := range role.Permissions {
					currentUserPermissions[p] = true
				}
			} else {
				core.UseGoeContainer().GetLogger().Warnf("RBAC: Role '%s' assigned to user '%s' not defined.", roleName, userID)
			}
		}

		// Check permissions
		if len(requiredPermissions) == 0 {
			return ctx.Next() // No specific permissions required
		}

		hasPermission := false
		if mode == MatchAll {
			hasPermission = true // Assume true, then check if any are missing
			for _, reqP := range requiredPermissions {
				if !currentUserPermissions[reqP] {
					hasPermission = false
					break
				}
			}
		} else { // rbac.MatchAtLeastOne
			for _, reqP := range requiredPermissions {
				if currentUserPermissions[reqP] {
					hasPermission = true
					break
				}
			}
		}

		if hasPermission {
			return ctx.Next()
		}

		// Construct a more informative message for the log
		var reqPermsStr []string
		for _, p := range requiredPermissions {
			reqPermsStr = append(reqPermsStr, string(p))
		}
		var userPermsStr []string
		for p := range currentUserPermissions {
			userPermsStr = append(userPermsStr, string(p))
		}
		core.UseGoeContainer().GetLogger().Debugf(
			"RBAC: Access denied for user '%s'. Required: [%s] (Mode: %d). User has: [%s]",
			userID,
			strings.Join(reqPermsStr, ", "),
			mode,
			strings.Join(userPermsStr, ", "),
		)

		return webresult.Forbidden("Insufficient permissions.")
	}
}

// DefineRole adds or updates a role definition in the global store.
// This is intended to be called during application initialization.
func DefineRole(role *Role) error {
	if strutil.IsBlank(role.Name) {
		core.UseGoeContainer().GetLogger().Warn("RBAC: Attempted to define a role with an empty name.")
		return errors.New("role name is blank")
	}
	if len(role.Permissions) == 0 {
		core.UseGoeContainer().GetLogger().Warnf("RBAC: Role '%s' defined with no permissions.", role.Name)
		return errors.New("role has no permissions")
	}

	_, err := core.UseGoeContainer().GetMongo().Insert(role)
	if err != nil {
		return err
	}
	core.UseGoeContainer().GetLogger().Debugf("RBAC: Role '%s' defined with permissions: %v", role.Name, role.Permissions)
	return nil
}

// GetRole retrieves a defined role by its name.
// Returns the role and true if found, otherwise an empty Role and false.
func GetRole(name string) (*Role, bool) {
	role := &Role{}
	hasResult, err := core.UseGoeContainer().GetMongo().FindOne(role, bson.M{"name": name}, role)
	if err != nil {
		return nil, false
	}
	if !hasResult {
		core.UseGoeContainer().GetLogger().Debugf("RBAC: Attempted to get non-defined role '%s'", name)
		return nil, false
	}
	return role, hasResult
}

// === Role Management ===

// AssignRoleToUser assigns a role to a user.
// If the role is already assigned, it does nothing.
func AssignRoleToUser(userID string, roleName string) error {
	if strutil.IsBlank(userID) {
		return errors.New("RBAC: UserID cannot be blank")
	}
	if strutil.IsBlank(roleName) {
		return errors.New("RBAC: Role name cannot be blank")
	}

	// Check if the role is defined (optional, but good practice)
	if _, exists := GetRole(roleName); !exists {
		core.UseGoeContainer().GetLogger().Warnf("RBAC: Assigning non-globally-defined role '%s' to user '%s'. This role must exist elsewhere or its permissions won't be found by GetRole.", roleName, userID)
	}

	mongo := core.UseGoeContainer().GetMongo()
	assignmentModel := &UserRoleAssignment{}

	// Check if assignment already exists
	filter := bson.M{"user_id": userID, "role_name": roleName}
	exists, err := mongo.IsExist(assignmentModel, filter)
	if err != nil {
		return fmt.Errorf("RBAC: Error checking if role assignment exists for user %s, role %s: %w", userID, roleName, err)
	}
	if exists {
		core.UseGoeContainer().GetLogger().Debugf("RBAC: Role '%s' already assigned to user '%s'. No action taken.", roleName, userID)
		return nil // Already assigned
	}

	newAssignment := &UserRoleAssignment{
		UserID:   userID,
		RoleName: roleName,
	}
	// DefaultModel fields (ID, CreatedAt, UpdatedAt) will be set by BeforeInsert hook

	_, err = mongo.Insert(newAssignment)
	if err != nil {
		return fmt.Errorf("RBAC: Error assigning role '%s' to user '%s': %w", roleName, userID, err)
	}
	core.UseGoeContainer().GetLogger().Infof("RBAC: Assigned role '%s' to user '%s'.", roleName, userID)
	return nil
}

// RevokeRoleFromUser revokes a role from a user.
func RevokeRoleFromUser(userID string, roleName string) error {
	if strutil.IsBlank(userID) {
		return errors.New("RBAC: UserID cannot be zero")
	}
	if roleName == "" {
		return errors.New("RBAC: Role name cannot be empty")
	}

	mongo := core.UseGoeContainer().GetMongo()
	assignmentModel := &UserRoleAssignment{}
	filter := bson.M{"user_id": userID, "role_name": roleName}

	result, err := mongo.DeleteMany(assignmentModel, filter)
	if err != nil {
		return fmt.Errorf("RBAC: Error revoking role '%s' from user '%s': %w", roleName, userID, err)
	}
	if result.DeletedCount > 0 {
		core.UseGoeContainer().GetLogger().Infof("RBAC: Revoked role '%s' from user '%s'. Count: %d", roleName, userID, result.DeletedCount)
	} else {
		core.UseGoeContainer().GetLogger().Debugf("RBAC: No role assignment found for role '%s' and user '%s' to revoke.", roleName, userID)
	}
	return nil
}

// GetUserRoleNames retrieves all role names assigned to a user.
// This will be used by the RBAC middleware.
func GetUserRoleNames(userID string) ([]string, error) {
	if strutil.IsBlank(userID) {
		return nil, errors.New("RBAC: UserID cannot be zero")
	}

	mongo := core.UseGoeContainer().GetMongo()
	var assignments []string

	err := mongo.Find(&UserRoleAssignment{}, bson.M{"user_id": userID}).Distinct("role_name", &assignments)
	if err != nil {
		return nil, fmt.Errorf("RBAC: Error fetching role assignments for user '%s': %w", userID, err)
	}

	if len(assignments) == 0 {
		return []string{}, nil
	}

	return assignments, nil
}

// === Direct Permission Management ===

// GrantDirectPermission grants a direct permission to a user.
// It updates the user's UserDirectPermission document, adding the new permission
// if it doesn't already exist in their list.
func GrantDirectPermission(userID string, permission Permission) error {
	if strutil.IsBlank(userID) {
		return errors.New("RBAC: UserID cannot be zero")
	}
	if permission == "" {
		return errors.New("RBAC: Permission cannot be empty")
	}

	mongo := core.UseGoeContainer().GetMongo()
	directPermsModel := &UserDirectPermission{}

	filter := bson.M{"user_id": userID}
	update := bson.M{"$addToSet": bson.M{"permissions": permission}}
	// $addToSet ensures the permission is only added if it's not already present.

	// Upsert ensures that if the user document doesn't exist, it's created.
	// The BeforeInsert hook in UserDirectPermission model will initialize empty Permissions slice.
	err, _ := mongo.Collection(directPermsModel).Upsert(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("RBAC: Error granting direct permission '%s' to user '%s': %w", permission, userID, err)
	}

	// Ensure the UserDirectPermission document is properly initialized if it was just created
	// This is especially for the case where $addToSet on a non-existent array field might behave unexpectedly
	// or if we want to ensure CreatedAt/UpdatedAt are set on creation.
	// The BeforeInsert hook should handle this, but an explicit check or an UpdateOne
	// with "$setOnInsert" for CreatedAt might be more robust if upsert creates the doc.
	// The current DefaultModel and hooks should manage this.
	// Forcing an update to set UpdatedAt:
	// If the document was upserted and is new, BeforeInsert runs.
	// If it existed, we should ensure UpdatedAt is touched.
	// $currentDate might be better here if we don't rely on hooks for existing docs.
	// However, our current DefaultModel hooks only run for Insert/Update calls via the mongo wrapper.
	// A direct UpdateOne like this bypasses DefaultModel's BeforeUpdate.
	// Let's try to fetch and save to trigger hooks if it's cleaner.

	// Simpler approach: just log. The $addToSet and upsert are generally fine.
	// Hooks are for ORM-like methods (mongo.Insert, mongo.Update).
	// For now, we assume the update operation is sufficient and hooks are for the specific ORM methods.

	core.UseGoeContainer().GetLogger().Infof("RBAC: Granted direct permission '%s' to user '%s'.", permission, userID)
	return nil
}

// RevokeDirectPermission revokes a direct permission from a user.
func RevokeDirectPermission(userID string, permission Permission) error {
	if strutil.IsBlank(userID) {
		return errors.New("RBAC: UserID cannot be zero")
	}
	if permission == "" {
		return errors.New("RBAC: Permission cannot be empty")
	}

	mongo := core.UseGoeContainer().GetMongo()
	directPermsModel := &UserDirectPermission{}

	filter := bson.M{"user_id": userID}
	update := bson.M{"$pull": bson.M{"permissions": permission}}

	err := mongo.Collection(directPermsModel).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("RBAC: Error revoking direct permission '%s' from user '%s': %w", permission, userID, err)
	}

	if mongodb.IsNoResult(err) {
		core.UseGoeContainer().GetLogger().Debugf("RBAC: No direct permission '%s' found for user '%s' to revoke, or user document does not exist.", permission, userID)
	} else {
		core.UseGoeContainer().GetLogger().Infof("RBAC: Revoked direct permission '%s' from user '%s'.", permission, userID)
	}
	return nil
}

// GetUserDirectPermissions retrieves all direct permissions for a user.
// This will be used by the RBAC middleware.
func GetUserDirectPermissions(userID string) ([]Permission, error) {
	if strutil.IsBlank(userID) {
		return nil, errors.New("RBAC: UserID cannot be zero")
	}

	mongo := core.UseGoeContainer().GetMongo()
	var userPermsDoc UserDirectPermission

	found, err := mongo.FindOne(&userPermsDoc, bson.M{"user_id": userID}, &userPermsDoc)
	if err != nil {
		return nil, fmt.Errorf("RBAC: Error fetching direct permissions for user '%s': %w", userID, err)
	}
	if !found {
		return []Permission{}, nil // No document means no direct permissions
	}

	return userPermsDoc.Permissions, nil
}

// === Combined Permissions ===

// GetAllUserPermissions retrieves all unique permissions for a user,
// combining their direct permissions and permissions from all their assigned roles.
func GetAllUserPermissions(userID string) ([]Permission, error) {
	if strutil.IsBlank(userID) {
		return nil, errors.New("RBAC: UserID cannot be zero")
	}

	allPermissionsMap := make(map[Permission]bool)

	// Get direct permissions
	directPermissions, err := GetUserDirectPermissions(userID)
	if err != nil {
		return nil, fmt.Errorf("RBAC: Error getting direct permissions for user '%s': %w", userID, err)
	}
	for _, p := range directPermissions {
		allPermissionsMap[p] = true
	}

	// Get role-based permissions
	roleNames, err := GetUserRoleNames(userID)
	if err != nil {
		return nil, fmt.Errorf("RBAC: Error getting role names for user '%s': %w", userID, err)
	}

	for _, roleName := range roleNames {
		role, found := GetRole(roleName) // Get from code-defined roles
		if found {
			for _, p := range role.Permissions {
				allPermissionsMap[p] = true
			}
		} else {
			core.UseGoeContainer().GetLogger().Warnf("RBAC: Role '%s' (assigned to user '%s') not found in code-defined roles during GetAllUserPermissions.", roleName, userID)
		}
	}

	// Convert map to slice
	finalPermissionsList := make([]Permission, 0, len(allPermissionsMap))
	for p := range allPermissionsMap {
		finalPermissionsList = append(finalPermissionsList, p)
	}

	return finalPermissionsList, nil
}

// Placeholder functions fetchUserRoleNames and fetchUserDirectPermissions are now removed.

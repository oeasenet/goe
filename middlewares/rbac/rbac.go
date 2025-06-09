package rbac

import (
	"context"
	"errors"
	"fmt"
	"github.com/gookit/goutil/strutil"
	"go.mongodb.org/mongo-driver/bson"
	officialOpts "go.mongodb.org/mongo-driver/mongo/options"
	"go.oease.dev/goe/modules/mongodb"
	"go.oease.dev/omgo/options"
	"strings"

	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/core"
	"go.oease.dev/goe/webresult"
)

// CheckPermission returns a fiber.Handler that checks if the current user has
// the required permissions.
//
// requiredPermissions are the permissions needed to access the route.
//
// mode specifies whether all or at least one of the requiredPermissions are needed.
func (m *RBACMiddleware) CheckPermission(requiredPermissions []Permission, mode MatchMode) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userID := m.userIdGetter(ctx)
		if strutil.IsBlank(userID) {
			core.UseGoeContainer().GetLogger().Warn("RBAC: User ID is BLANK")
			return webresult.Forbidden("Access denied.")
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
			role, found := GetRole(roleName)
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

// DefineRole adds a role to mongoDB.
// if role exists, will update permission
func DefineRole(role *Role) error {
	if strutil.IsBlank(role.Name) {
		core.UseGoeContainer().GetLogger().Warn("RBAC: Attempted to define a role with an empty name.")
		return errors.New("role name is blank")
	}
	if len(role.Permissions) == 0 {
		core.UseGoeContainer().GetLogger().Warnf("RBAC: Role '%s' defined with no permissions.", role.Name)
		return errors.New("role has no permissions")
	}

	filter := bson.M{"name": role.Name}
	update := bson.M{"$set": bson.M{"permissions": role.Permissions}}

	opt := officialOpts.Update().SetUpsert(true)
	opts := options.UpdateOptions{UpdateOptions: opt}
	err := core.UseGoeContainer().GetMongo().Collection(role).UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return err
	}
	core.UseGoeContainer().GetLogger().Debugf("RBAC: Role '%s' defined with permissions: %v", role.Name, role.Permissions)
	return nil
}

// GetRole retrieves a defined role by its name.
// Returns the role and true if found, otherwise an empty Role and false.
func GetRole(name string) (*Role, bool) {
	if strutil.IsBlank(name) {
		core.UseGoeContainer().GetLogger().Warn("RBAC: Attempted to define a role with an empty name.")
		return nil, false
	}
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

// DeleteRole deletes a role by its name from the database.
func DeleteRole(name string) error {
	if strutil.IsBlank(name) {
		core.UseGoeContainer().GetLogger().Warn("RBAC: Attempted to define a role with an empty name.")
		return errors.New("role name is blank")
	}

	_, err := core.UseGoeContainer().GetMongo().DeleteMany(&Role{}, bson.M{"name": name})
	if err != nil {
		return err
	}

	_, err = core.UseGoeContainer().GetMongo().DeleteMany(&UserRoleAssignment{}, bson.M{"role_name": name})
	if err != nil {
		return err
	}
	return nil
}

func ListRoles(pageSize int64, currentPage int64) ([]*Role, error) {
	if pageSize <= 0 {
		return []*Role{}, errors.New("page size is illegal")
	}
	if currentPage <= 0 {
		return []*Role{}, errors.New("current page is illegal")
	}
	r := &Role{}
	roles := make([]*Role, pageSize)
	_, _ = core.UseGoeContainer().GetMongo().FindPage(r, nil, &roles, pageSize, currentPage)
	return roles, nil
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
	opt := officialOpts.Update().SetUpsert(true)
	opts := options.UpdateOptions{UpdateOptions: opt}
	err := mongo.Collection(directPermsModel).UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return fmt.Errorf("RBAC: Error granting direct permission '%s' to user '%s': %w", permission, userID, err)
	}

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

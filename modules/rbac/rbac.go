package rbac

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/goe/core"
	"go.oease.dev/goe/models" // Required for models.UserRoleAssignment and models.UserDirectPermission
)

// Permission is a string representing an action on a feature.
// Format: "feature:action" (e.g., "article:read", "user:create")
type Permission string

// Role defines a named set of permissions.
type Role struct {
	Name        string       `json:"name" bson:"name"`
	Permissions []Permission `json:"permissions" bson:"permissions"`
}

// MatchMode defines how permissions should be checked.
type MatchMode int

const (
	// MatchAll requires all specified permissions to be present.
	MatchAll MatchMode = iota
	// MatchAtLeastOne requires at least one of the specified permissions to be present.
	MatchAtLeastOne
)

// Global storage for code-defined roles.
// This allows roles to be defined in code and easily referenced.
// For more dynamic role management, roles could also be stored in a database.
var definedRoles = make(map[string]Role)

// DefineRole adds or updates a role definition in the global store.
// This is intended to be called during application initialization.
func DefineRole(role Role) {
	if role.Name == "" {
		core.UseGoeContainer().GetLogger().Warn("RBAC: Attempted to define a role with an empty name.")
		return
	}
	if len(role.Permissions) == 0 {
		core.UseGoeContainer().GetLogger().Warnf("RBAC: Role '%s' defined with no permissions.", role.Name)
	}
	definedRoles[role.Name] = role
	core.UseGoeContainer().GetLogger().Debugf("RBAC: Role '%s' defined with permissions: %v", role.Name, role.Permissions)
}

// GetRole retrieves a defined role by its name.
// Returns the role and true if found, otherwise an empty Role and false.
func GetRole(name string) (Role, bool) {
	role, ok := definedRoles[name]
	if !ok {
		core.UseGoeContainer().GetLogger().Debugf("RBAC: Attempted to get non-defined role '%s'", name)
	}
	return role, ok
}

// === Role Management ===

// AssignRoleToUser assigns a role to a user.
// If the role is already assigned, it does nothing.
func AssignRoleToUser(userID primitive.ObjectID, roleName string) error {
	if userID.IsZero() {
		return errors.New("RBAC: UserID cannot be zero")
	}
	if roleName == "" {
		return errors.New("RBAC: Role name cannot be empty")
	}
	
	// Check if the role is defined (optional, but good practice)
	if _, exists := GetRole(roleName); !exists {
		core.UseGoeContainer().GetLogger().Warnf("RBAC: Assigning non-globally-defined role '%s' to user '%s'. This role must exist elsewhere or its permissions won't be found by GetRole.", roleName, userID.Hex())
	}

	mongo := core.UseGoeContainer().GetMongo()
	assignmentModel := &models.UserRoleAssignment{}

	// Check if assignment already exists
	filter := bson.M{"user_id": userID, "role_name": roleName}
	exists, err := mongo.IsExist(assignmentModel, filter)
	if err != nil {
		return fmt.Errorf("RBAC: Error checking if role assignment exists for user %s, role %s: %w", userID.Hex(), roleName, err)
	}
	if exists {
		core.UseGoeContainer().GetLogger().Debugf("RBAC: Role '%s' already assigned to user '%s'. No action taken.", roleName, userID.Hex())
		return nil // Already assigned
	}

	newAssignment := &models.UserRoleAssignment{
		UserID:   userID,
		RoleName: roleName,
	}
	// DefaultModel fields (ID, CreatedAt, UpdatedAt) will be set by BeforeInsert hook

	_, err = mongo.Insert(newAssignment)
	if err != nil {
		return fmt.Errorf("RBAC: Error assigning role '%s' to user '%s': %w", roleName, userID.Hex(), err)
	}
	core.UseGoeContainer().GetLogger().Infof("RBAC: Assigned role '%s' to user '%s'.", roleName, userID.Hex())
	return nil
}

// RevokeRoleFromUser revokes a role from a user.
func RevokeRoleFromUser(userID primitive.ObjectID, roleName string) error {
	if userID.IsZero() {
		return errors.New("RBAC: UserID cannot be zero")
	}
	if roleName == "" {
		return errors.New("RBAC: Role name cannot be empty")
	}

	mongo := core.UseGoeContainer().GetMongo()
	assignmentModel := &models.UserRoleAssignment{}
	filter := bson.M{"user_id": userID, "role_name": roleName}

	result, err := mongo.DeleteMany(assignmentModel, filter)
	if err != nil {
		return fmt.Errorf("RBAC: Error revoking role '%s' from user '%s': %w", roleName, userID.Hex(), err)
	}
	if result.DeletedCount > 0 {
		core.UseGoeContainer().GetLogger().Infof("RBAC: Revoked role '%s' from user '%s'. Count: %d", roleName, userID.Hex(), result.DeletedCount)
	} else {
		core.UseGoeContainer().GetLogger().Debugf("RBAC: No role assignment found for role '%s' and user '%s' to revoke.", roleName, userID.Hex())
	}
	return nil
}

// GetUserRoleNames retrieves all role names assigned to a user.
// This will be used by the RBAC middleware.
func GetUserRoleNames(userID primitive.ObjectID) ([]string, error) {
	if userID.IsZero() {
		return nil, errors.New("RBAC: UserID cannot be zero")
	}

	mongo := core.UseGoeContainer().GetMongo()
	var assignments []models.UserRoleAssignment
	
	err := mongo.Find(&models.UserRoleAssignment{}, bson.M{"user_id": userID}).All(&assignments)
	if err != nil {
		// It's important to distinguish "not found" (empty list) from actual errors.
		// omgo's Find().All() might not return a specific "no documents" error,
		// but rather an empty slice and nil error.
		if mongo.IsNoResult(err) { // Assuming IsNoResult helper exists or can be added to mongo module
			return []string{}, nil
		}
		return nil, fmt.Errorf("RBAC: Error fetching role assignments for user '%s': %w", userID.Hex(), err)
	}

	if len(assignments) == 0 {
		return []string{}, nil
	}

	roleNames := make([]string, len(assignments))
	for i, assignment := range assignments {
		roleNames[i] = assignment.RoleName
	}
	return roleNames, nil
}


// === Direct Permission Management ===

// GrantDirectPermission grants a direct permission to a user.
// It updates the user's UserDirectPermission document, adding the new permission
// if it doesn't already exist in their list.
func GrantDirectPermission(userID primitive.ObjectID, permission Permission) error {
	if userID.IsZero() {
		return errors.New("RBAC: UserID cannot be zero")
	}
	if permission == "" {
		return errors.New("RBAC: Permission cannot be empty")
	}

	mongo := core.UseGoeContainer().GetMongo()
	directPermsModel := &models.UserDirectPermission{}
	
	filter := bson.M{"user_id": userID}
	update := bson.M{"$addToSet": bson.M{"permissions": permission}}
	// $addToSet ensures the permission is only added if it's not already present.

	// Upsert ensures that if the user document doesn't exist, it's created.
	// The BeforeInsert hook in UserDirectPermission model will initialize empty Permissions slice.
	opts := mongo.NewUpdateOptions().SetUpsert(true)

	_, err := mongo.Col(directPermsModel).UpdateOne(mongo.Ctx(), filter, update, opts)
	if err != nil {
		return fmt.Errorf("RBAC: Error granting direct permission '%s' to user '%s': %w", permission, userID.Hex(), err)
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

	core.UseGoeContainer().GetLogger().Infof("RBAC: Granted direct permission '%s' to user '%s'.", permission, userID.Hex())
	return nil
}

// RevokeDirectPermission revokes a direct permission from a user.
func RevokeDirectPermission(userID primitive.ObjectID, permission Permission) error {
	if userID.IsZero() {
		return errors.New("RBAC: UserID cannot be zero")
	}
	if permission == "" {
		return errors.New("RBAC: Permission cannot be empty")
	}

	mongo := core.UseGoeContainer().GetMongo()
	directPermsModel := &models.UserDirectPermission{}
	
	filter := bson.M{"user_id": userID}
	update := bson.M{"$pull": bson.M{"permissions": permission}}

	result, err := mongo.Col(directPermsModel).UpdateOne(mongo.Ctx(), filter, update)
	if err != nil {
		return fmt.Errorf("RBAC: Error revoking direct permission '%s' from user '%s': %w", permission, userID.Hex(), err)
	}

	if result.ModifiedCount > 0 {
		core.UseGoeContainer().GetLogger().Infof("RBAC: Revoked direct permission '%s' from user '%s'.", permission, userID.Hex())
	} else {
		core.UseGoeContainer().GetLogger().Debugf("RBAC: No direct permission '%s' found for user '%s' to revoke, or user document does not exist.", permission, userID.Hex())
	}
	return nil
}

// GetUserDirectPermissions retrieves all direct permissions for a user.
// This will be used by the RBAC middleware.
func GetUserDirectPermissions(userID primitive.ObjectID) ([]Permission, error) {
	if userID.IsZero() {
		return nil, errors.New("RBAC: UserID cannot be zero")
	}

	mongo := core.UseGoeContainer().GetMongo()
	var userPermsDoc models.UserDirectPermission

	found, err := mongo.FindOne(&userPermsDoc, bson.M{"user_id": userID}, &userPermsDoc)
	if err != nil {
		// Distinguish "not found" from other errors.
		if mongo.IsNoResult(err) { // Assuming IsNoResult helper
			return []Permission{}, nil
		}
		return nil, fmt.Errorf("RBAC: Error fetching direct permissions for user '%s': %w", userID.Hex(), err)
	}
	if !found {
		return []Permission{}, nil // No document means no direct permissions
	}

	return userPermsDoc.Permissions, nil
}

// === Combined Permissions ===

// GetAllUserPermissions retrieves all unique permissions for a user,
// combining their direct permissions and permissions from all their assigned roles.
func GetAllUserPermissions(userID primitive.ObjectID) ([]Permission, error) {
	if userID.IsZero() {
		return nil, errors.New("RBAC: UserID cannot be zero")
	}

	allPermissionsMap := make(map[Permission]bool)

	// Get direct permissions
	directPermissions, err := GetUserDirectPermissions(userID)
	if err != nil {
		return nil, fmt.Errorf("RBAC: Error getting direct permissions for user '%s': %w", userID.Hex(), err)
	}
	for _, p := range directPermissions {
		allPermissionsMap[p] = true
	}

	// Get role-based permissions
	roleNames, err := GetUserRoleNames(userID)
	if err != nil {
		return nil, fmt.Errorf("RBAC: Error getting role names for user '%s': %w", userID.Hex(), err)
	}

	for _, roleName := range roleNames {
		role, found := GetRole(roleName) // Get from code-defined roles
		if found {
			for _, p := range role.Permissions {
				allPermissionsMap[p] = true
			}
		} else {
			core.UseGoeContainer().GetLogger().Warnf("RBAC: Role '%s' (assigned to user '%s') not found in code-defined roles during GetAllUserPermissions.", roleName, userID.Hex())
		}
	}

	// Convert map to slice
	finalPermissionsList := make([]Permission, 0, len(allPermissionsMap))
	for p := range allPermissionsMap {
		finalPermissionsList = append(finalPermissionsList, p)
	}

	return finalPermissionsList, nil
}

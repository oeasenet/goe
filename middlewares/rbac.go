package middlewares

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/goe/core"
	// "go.oease.dev/goe/models"; // No longer needed directly here if rbac module handles models
	"go.oease.dev/goe/modules/rbac" // Now using this for permission logic
	"go.oease.dev/goe/webresult"
)

// RBACMiddleware provides middleware for checking user permissions.
type RBACMiddleware struct {
	// Potentially add configuration fields here in the future if needed
}

// NewRBACMiddleware creates a new RBACMiddleware instance.
func NewRBACMiddleware() *RBACMiddleware {
	return &RBACMiddleware{}
}

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
func (m *RBACMiddleware) CheckPermission(userIDKey string, requiredPermissions []rbac.Permission, mode rbac.MatchMode) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userVal := ctx.Locals(userIDKey)
		if userVal == nil {
			core.UseGoeContainer().GetLogger().Error("RBAC: User ID not found in context locals. Ensure auth middleware runs before RBAC.")
			return webresult.SendForbidden(ctx, "Access denied. User identifier missing.")
		}

		var userID primitive.ObjectID
		var err error

		switch v := userVal.(type) {
		case string:
			userID, err = primitive.ObjectIDFromHex(v)
			if err != nil {
				core.UseGoeContainer().GetLogger().Warnf("RBAC: Invalid User ID format in context locals: %v", err)
				return webresult.SendForbidden(ctx, "Access denied. Invalid user identifier.")
			}
		case primitive.ObjectID:
			userID = v
		default:
			core.UseGoeContainer().GetLogger().Warnf("RBAC: Unexpected User ID type in context locals: %T", userVal)
			return webresult.SendForbidden(ctx, "Access denied. Unexpected user identifier type.")
		}

		if userID.IsZero() {
			core.UseGoeContainer().GetLogger().Warn("RBAC: User ID is zero primitive.ObjectID.")
			return webresult.SendForbidden(ctx, "Access denied. User identifier is zero.")
		}
		
		// Use actual functions from rbac module
		userRoleNames, err := rbac.GetUserRoleNames(userID) // UPDATED
		if err != nil {
			core.UseGoeContainer().GetLogger().Errorf("RBAC: Error fetching user roles for user '%s': %v", userID.Hex(), err)
			return webresult.SystemBusy("Error checking permissions.")
		}

		directPermissions, err := rbac.GetUserDirectPermissions(userID) // UPDATED
		if err != nil {
			core.UseGoeContainer().GetLogger().Errorf("RBAC: Error fetching direct user permissions for user '%s': %v", userID.Hex(), err)
			return webresult.SystemBusy("Error checking permissions.")
		}
		
		currentUserPermissions := make(map[rbac.Permission]bool)

		// Add direct permissions
		for _, p := range directPermissions {
			currentUserPermissions[p] = true
		}

		// Add permissions from roles
		for _, roleName := range userRoleNames {
			role, found := rbac.GetRole(roleName) // Using GetRole from modules/rbac/rbac.go
			if found {
				for _, p := range role.Permissions {
					currentUserPermissions[p] = true
				}
			} else {
				core.UseGoeContainer().GetLogger().Warnf("RBAC: Role '%s' assigned to user '%s' not defined.", roleName, userID.Hex())
			}
		}
		
		// Check permissions
		if len(requiredPermissions) == 0 {
			return ctx.Next() // No specific permissions required
		}

		hasPermission := false
		if mode == rbac.MatchAll {
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
			userID.Hex(),
			strings.Join(reqPermsStr, ", "),
			mode,
			strings.Join(userPermsStr, ", "),
		)
		
		return webresult.SendForbidden(ctx, "Insufficient permissions.")
	}
}

// Placeholder functions fetchUserRoleNames and fetchUserDirectPermissions are now removed.

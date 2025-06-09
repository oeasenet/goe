package main

import (
	_ "embed"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe"
	"go.oease.dev/goe/middlewares/rbac"
	"go.oease.dev/goe/webresult"
)

func main() {
	err := goe.NewApp()
	if err != nil {
		panic(err)
	}

	// --- RBAC Setup Example ---
	// 1. Define Roles
	// Typically, you'd define roles once at application startup.
	_ = rbac.DefineRole(&rbac.Role{
		Name:        "admin",
		Permissions: []rbac.Permission{"user:create", "user:read", "user:update", "user:delete", "article:publish"},
	})
	_ = rbac.DefineRole(&rbac.Role{
		Name:        "editor",
		Permissions: []rbac.Permission{"article:create", "article:read", "article:update", "article:publish"},
	})
	_ = rbac.DefineRole(&rbac.Role{
		Name:        "viewer",
		Permissions: []rbac.Permission{"article:read"},
	})

	goe.UseLog().Info("RBAC roles defined.")

	// 2. Simulate User Permission/Role Assignment
	// In a real app, this would happen based on user registration, admin actions, etc.
	// Let's create a couple of simulated user IDs.
	// IMPORTANT: For this example to work consistently across runs without a real DB for users,
	// we would typically clear previous assignments for these test users if the DB persists.
	// However, the RBAC functions are designed to be idempotent where possible (e.g., AssignRoleToUser won't duplicate).

	// User 1: The Admin User (hex string for ObjectID)
	adminUserID := "adminUserID"

	// It's good practice to revoke any existing roles/permissions before assigning
	// to ensure a clean state for test users, especially if the underlying DB persists data.
	// For this example, we'll assume a clean slate or rely on idempotency.
	// Example: rbac.RevokeRoleFromUser(adminUserID, "admin") // if needed
	err = rbac.AssignRoleToUser(adminUserID, "admin")
	if err != nil {
		goe.UseLog().Errorf("Failed to assign admin role to user %s: %v", adminUserID, err)
	} else {
		goe.UseLog().Infof("Assigned 'admin' role to user %s", adminUserID)
	}

	// User 2: The Editor User
	editorUserID := "editorUserID"
	err = rbac.AssignRoleToUser(editorUserID, "editor")
	if err != nil {
		goe.UseLog().Errorf("Failed to assign editor role to user %s: %v", editorUserID, err)
	}
	// Let's also give this editor a special direct permission
	err = rbac.GrantDirectPermission(editorUserID, "article:feature")
	if err != nil {
		goe.UseLog().Errorf("Failed to grant direct 'article:feature' permission to user %s: %v", editorUserID, err)
	} else {
		goe.UseLog().Infof("Assigned 'editor' role and 'article:feature' direct permission to user %s", editorUserID)
	}

	// User 3: The Viewer User
	viewerUserID := "viewerUserID"
	err = rbac.AssignRoleToUser(viewerUserID, "viewer")
	if err != nil {
		goe.UseLog().Errorf("Failed to assign viewer role to user %s: %v", viewerUserID, err)
	} else {
		goe.UseLog().Infof("Assigned 'viewer' role to user %s", viewerUserID)
	}
	// --- End of RBAC Setup Example ---

	// Initialize RBAC Middleware
	rbacMw := rbac.NewRBACMiddleware()
	rbacMw.SetUserIdGetter(rbac.GetDefaultUserIdGetter)

	// Fiber App
	app := goe.UseFiber().App()

	app.Get("/hello", func(ctx fiber.Ctx) error {
		goe.UseLog().Error("test caller skip") // Original log line, seems for testing.
		// err := errors.New("errrrrrrr") // Original error simulation, commented out for RBAC example clarity
		// if err != nil {
		// 	return webresult.SystemBusy(err)
		// }
		// For this example, let's assume the /hello route is public
		return webresult.SendSucceed(ctx, "Hello, World! This is a public endpoint.")
	})

	// --- RBAC Protected Routes Examples ---

	// Public route (no auth, no RBAC)
	app.Get("/public/info", func(ctx fiber.Ctx) error {
		return webresult.SendSucceed(ctx, "This is public information.")
	})

	// optional, this is for test
	app.Use(SetUserIDMiddleware())

	// Viewer-accessible route (simulating viewer login)
	// Requires "article:read"
	app.Get("/article/view",
		func(ctx fiber.Ctx) error {
			return webresult.SendSucceed(ctx, fmt.Sprintf("user viewing article (article:read permission checked)."))
		},
		rbacMw.CheckPermission([]rbac.Permission{"article:read"}, rbac.MatchAtLeastOne),
	)

	// Editor-accessible route (simulating editor login)
	// Requires "article:update" AND "article:publish"
	app.Post("/article/publish",
		func(ctx fiber.Ctx) error {
			return webresult.SendSucceed(ctx, fmt.Sprintf("user publishing article (article:update AND article:publish checked)."))
		},
		rbacMw.CheckPermission([]rbac.Permission{"article:update", "article:publish"}, rbac.MatchAll),
	)

	// Editor-accessible route for special feature (direct permission)
	// Requires "article:feature"
	app.Get("/article/specialfeature",
		func(ctx fiber.Ctx) error {
			return webresult.SendSucceed(ctx, fmt.Sprintf("user accessing special article feature (article:feature direct permission checked)."))
		},
		rbacMw.CheckPermission([]rbac.Permission{"article:feature"}, rbac.MatchAtLeastOne),
	)

	// Admin-only route (simulating admin login)
	// Requires "user:create"
	app.Post("/admin/users/create",
		func(ctx fiber.Ctx) error {
			return webresult.SendSucceed(ctx, fmt.Sprintf("user creating a new user (user:create permission checked)."))
		},
		rbacMw.CheckPermission([]rbac.Permission{"user:create"}, rbac.MatchAtLeastOne),
	)

	// Test route for a user with NO relevant permissions for this route
	// Requires "secret:read"
	app.Get("/secret/data",
		func(ctx fiber.Ctx) error {
			// This should not be reached by the viewer if RBAC is working.
			// The RBAC middleware will return a 403 Forbidden.
			return webresult.SendSucceed(ctx, "User somehow accessed secret data. This indicates a problem if the user was the viewer.")
		},
		rbacMw.CheckPermission([]rbac.Permission{"secret:read"}, rbac.MatchAtLeastOne),
	)

	err = goe.Run()
	if err != nil {
		panic(err)
	}
}

func SetUserIDMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Query(rbac.UserIDKey)
		if userID == "" {
			return c.Status(fiber.StatusForbidden).SendString("Missing user ID")
		}
		c.Locals(rbac.UserIDKey, userID)
		fmt.Println(c.Locals(rbac.UserIDKey))
		return c.Next()
	}
}

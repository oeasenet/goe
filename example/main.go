package main

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson/primitive" // Needed for ObjectID
	"go.oease.dev/goe"
	"go.oease.dev/goe/middlewares"
	"go.oease.dev/goe/modules/rbac" // Needed for rbac types
	"go.oease.dev/goe/webresult"
)

//go:embed configs/msearch.json
var msearchConfig []byte

// MockAuthMiddleware simulates an authentication middleware that sets a userID.
// In a real application, this would verify a token (JWT, session, etc.)
// and set the authenticated user's ID.
func MockAuthMiddleware(simulatedUserID string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		objID, err := primitive.ObjectIDFromHex(simulatedUserID)
		if err != nil {
			// For this example, if the simulatedUserID is invalid, deny access.
			// In a real app, an invalid or missing token would typically result
			// in an unauthorized error *before* RBAC checks.
			return webresult.SendUnauthorized(ctx, "MockAuth: Invalid simulated UserID format.")
		}
		ctx.Locals("userID", objID) // RBAC middleware will look for "userID"
		// For demonstration, also log who is "logged in"
		goe.UseLog().Infof("MockAuth: User '%s' is now 'authenticated'.", simulatedUserID)
		return ctx.Next()
	}
}

func main() {
	err := goe.NewApp()
	if err != nil {
		panic(err)
	}

	// --- RBAC Setup Example ---
	// 1. Define Roles
	// Typically, you'd define roles once at application startup.
	goe.UseRBAC().DefineRole(rbac.Role{
		Name:        "admin",
		Permissions: []rbac.Permission{"user:create", "user:read", "user:update", "user:delete", "article:publish"},
	})
	goe.UseRBAC().DefineRole(rbac.Role{
		Name:        "editor",
		Permissions: []rbac.Permission{"article:create", "article:read", "article:update", "article:publish"},
	})
	goe.UseRBAC().DefineRole(rbac.Role{
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
	adminUserIDHex := "650000000000000000000001" 
	adminUserID, _ := primitive.ObjectIDFromHex(adminUserIDHex) 
	if !adminUserID.IsZero() {
		// It's good practice to revoke any existing roles/permissions before assigning
		// to ensure a clean state for test users, especially if the underlying DB persists data.
		// For this example, we'll assume a clean slate or rely on idempotency.
		// Example: goe.UseRBAC().RevokeRoleFromUser(adminUserID, "admin") // if needed
		err = goe.UseRBAC().AssignRoleToUser(adminUserID, "admin")
		if err != nil {
			goe.UseLog().Errorf("Failed to assign admin role to user %s: %v", adminUserIDHex, err)
		} else {
			goe.UseLog().Infof("Assigned 'admin' role to user %s", adminUserIDHex)
		}
	}

	// User 2: The Editor User
	editorUserIDHex := "650000000000000000000002"
	editorUserID, _ := primitive.ObjectIDFromHex(editorUserIDHex)
	if !editorUserID.IsZero() {
		err = goe.UseRBAC().AssignRoleToUser(editorUserID, "editor")
		if err != nil {
			goe.UseLog().Errorf("Failed to assign editor role to user %s: %v", editorUserIDHex, err)
		}
		// Let's also give this editor a special direct permission
		err = goe.UseRBAC().GrantDirectPermission(editorUserID, "article:feature")
		if err != nil {
			goe.UseLog().Errorf("Failed to grant direct 'article:feature' permission to user %s: %v", editorUserIDHex, err)
		} else {
			goe.UseLog().Infof("Assigned 'editor' role and 'article:feature' direct permission to user %s", editorUserIDHex)
		}
	}
	
	// User 3: The Viewer User
	viewerUserIDHex := "650000000000000000000003"
	viewerUserID, _ := primitive.ObjectIDFromHex(viewerUserIDHex)
	if !viewerUserID.IsZero() {
		err = goe.UseRBAC().AssignRoleToUser(viewerUserID, "viewer")
		if err != nil {
			goe.UseLog().Errorf("Failed to assign viewer role to user %s: %v", viewerUserIDHex, err)
		} else {
			goe.UseLog().Infof("Assigned 'viewer' role to user %s", viewerUserIDHex)
		}
	}
	// --- End of RBAC Setup Example ---


	// Initialize RBAC Middleware
	rbacMw := middlewares.NewRBACMiddleware()
	userIDKeyForRBAC := "userID" // The key where MockAuthMiddleware stores the user ID

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

	// Viewer-accessible route (simulating viewer login)
	// Requires "article:read"
	app.Get("/article/view",
		MockAuthMiddleware(viewerUserIDHex), // Simulate viewer login
		rbacMw.CheckPermission(userIDKeyForRBAC, []rbac.Permission{"article:read"}, rbac.MatchAtLeastOne),
		func(ctx fiber.Ctx) error {
			return webresult.SendSucceed(ctx, fmt.Sprintf("User %s viewing article (article:read permission checked).", viewerUserIDHex))
		},
	)

	// Editor-accessible route (simulating editor login)
	// Requires "article:update" AND "article:publish"
	app.Post("/article/publish",
		MockAuthMiddleware(editorUserIDHex), // Simulate editor login
		rbacMw.CheckPermission(userIDKeyForRBAC, []rbac.Permission{"article:update", "article:publish"}, rbac.MatchAll),
		func(ctx fiber.Ctx) error {
			return webresult.SendSucceed(ctx, fmt.Sprintf("User %s publishing article (article:update AND article:publish checked).", editorUserIDHex))
		},
	)
	
	// Editor-accessible route for special feature (direct permission)
	// Requires "article:feature"
	app.Get("/article/specialfeature",
		MockAuthMiddleware(editorUserIDHex), // Simulate editor login
		rbacMw.CheckPermission(userIDKeyForRBAC, []rbac.Permission{"article:feature"}, rbac.MatchAtLeastOne),
		func(ctx fiber.Ctx) error {
			return webresult.SendSucceed(ctx, fmt.Sprintf("User %s accessing special article feature (article:feature direct permission checked).", editorUserIDHex))
		},
	)

	// Admin-only route (simulating admin login)
	// Requires "user:create"
	app.Post("/admin/users/create",
		MockAuthMiddleware(adminUserIDHex), // Simulate admin login
		rbacMw.CheckPermission(userIDKeyForRBAC, []rbac.Permission{"user:create"}, rbac.MatchAtLeastOne),
		func(ctx fiber.Ctx) error {
			return webresult.SendSucceed(ctx, fmt.Sprintf("User %s creating a new user (user:create permission checked).", adminUserIDHex))
		},
	)
	
	// Test route for a user with NO relevant permissions for this route
	// Requires "secret:read"
	app.Get("/secret/data",
		MockAuthMiddleware(viewerUserIDHex), // Simulate viewer login (viewer does not have secret:read)
		rbacMw.CheckPermission(userIDKeyForRBAC, []rbac.Permission{"secret:read"}, rbac.MatchAtLeastOne),
		func(ctx fiber.Ctx) error {
			// This should not be reached by the viewer if RBAC is working.
			// The RBAC middleware will return a 403 Forbidden.
			return webresult.SendSucceed(ctx, "User somehow accessed secret data. This indicates a problem if the user was the viewer.")
		},
	)


	// File uploader example (remains from original example, can also be RBAC protected)
	fileUploader := middlewares.NewFileMiddlewares()
	// Example of protecting a file upload route with RBAC (e.g., only editors can upload)
	app.Post("/file/upload",
		MockAuthMiddleware(editorUserIDHex), // Simulate editor login
		rbacMw.CheckPermission(userIDKeyForRBAC, []rbac.Permission{"article:create"}, rbac.MatchAtLeastOne), // Assuming "article:create" implies upload rights
		fileUploader.HandleUpload(),
	)
	app.Get("/file/view/:id", fileUploader.HandleView()) // Public view for this example
	// Example of protecting file deletion (e.g., only admins can delete)
	app.Delete("/file/delete/:id",
		MockAuthMiddleware(adminUserIDHex), // Simulate admin login
		rbacMw.CheckPermission(userIDKeyForRBAC, []rbac.Permission{"user:delete"}, rbac.MatchAtLeastOne), // Example: using a general delete permission
		fileUploader.HandleDelete(),
	)
	app.Get("/file/match/:hash", fileUploader.HandleMatch())


	err = goe.Run()
	if err != nil {
		panic(err)
	}
}

package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson" // Required for cleanupMwTestData
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/goe"
	"go.oease.dev/goe/models" // Required for cleanupMwTestData
	"go.oease.dev/goe/modules/rbac"
	"go.oease.dev/goe/webresult"
)

var testApp *fiber.App
var rbacMw *RBACMiddleware
var testUserIDKey = "userID"

// Test User IDs for middleware tests
var mwTestUserAdmin primitive.ObjectID
var mwTestUserEditor primitive.ObjectID
var mwTestUserViewer primitive.ObjectID
var mwTestUserNoPerms primitive.ObjectID


func TestMain(m *testing.M) {
	// Setup: Initialize Goe framework.
	// Similar to rbac_test.go, ensure this points to a test environment.
	if err := os.Setenv("GOE_APP_NAME", "goe-rbac-mw-test"); err != nil {
		fmt.Println("Failed to set GOE_APP_NAME for middleware tests:", err)
		os.Exit(1)
	}
	// Potentially set MONGODB_DATABASE to a test-specific one if not already handled by global test config
	// os.Setenv("MONGODB_DATABASE", "goe_test_rbac_mw") // Example

	err := goe.NewApp()
	if err != nil {
		fmt.Println("Failed to initialize Goe app for middleware tests:", err)
		os.Exit(1)
	}

	// Initialize test user ObjectIDs
	mwTestUserAdmin = primitive.NewObjectID()
	mwTestUserEditor = primitive.NewObjectID()
	mwTestUserViewer = primitive.NewObjectID()
	mwTestUserNoPerms = primitive.NewObjectID()

	// Define roles and assign permissions for test users
	// These interact with the actual RBAC service, potentially hitting the DB.
	// Ensure DB is clean or use unique user IDs for each test run if state is an issue.
	// The rbac_test.go TestMain should handle general cleanup if using the same DB.
	setupTestRBACData()

	// Initialize Fiber app and RBAC middleware
	testApp = fiber.New(fiber.Config{
		ErrorHandler: func(ctx fiber.Ctx, err error) error { // Custom error handler for tests
			// Check if the error is a fiber.Error
			var fiberError *fiber.Error
			if errors.As(err, &fiberError) {
				// Use fiberError.Code and fiberError.Message
				return ctx.Status(fiberError.Code).JSON(fiber.Map{"message": fiberError.Message})
			}
			// Handle other errors (like those from webresult or direct errors)
			// webresult.SendForbidden etc often write the response themselves.
			// If an error reaches here that *hasn't* written a response, this is a fallback.
			// Many of the webresult functions might not return an error that makes it here if they write response.
			if !ctx.Response().Written() {
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
			}
			return nil // Response already sent
		},
	})
	rbacMw = NewRBACMiddleware()

	// Define a common success handler for protected routes
	successHandler := func(ctx fiber.Ctx) error {
		return webresult.SendSucceed(ctx, "Access Granted")
	}

	// Setup test routes
	// Route requiring "article:read" (Viewer)
	testApp.Get("/view-article",
		func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserViewer); return ctx.Next() },
		rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:read"}, rbac.MatchAtLeastOne),
		successHandler,
	)

	// Route requiring "article:edit" (Editor)
	testApp.Get("/edit-article",
		func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserEditor); return ctx.Next() },
		rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:edit"}, rbac.MatchAtLeastOne),
		successHandler,
	)
	
	// Route requiring "article:publish" AND "article:review" (Admin via role, Editor needs direct for review)
	testApp.Get("/publish-reviewed-article",
		func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserAdmin); return ctx.Next() }, // Test with Admin
		rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:publish", "article:review"}, rbac.MatchAll),
		successHandler,
	)
	testApp.Get("/publish-reviewed-article-editor-fail", // Editor lacks article:review
		func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserEditor); return ctx.Next() },
		rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:publish", "article:review"}, rbac.MatchAll),
		successHandler,
	)
	
	// Route for user with no permissions
	testApp.Get("/no-perms-test",
		func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserNoPerms); return ctx.Next() },
		rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:read"}, rbac.MatchAtLeastOne),
		successHandler,
	)

	// Route for missing userID in locals
	testApp.Get("/missing-userid",
		rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:read"}, rbac.MatchAtLeastOne),
		successHandler,
	)
	
	// Route for invalid userID format in locals
	testApp.Get("/invalid-userid-format",
		func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, "not-an-object-id"); return ctx.Next() },
		rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:read"}, rbac.MatchAtLeastOne),
		successHandler,
	)
	
	// Route for zero primitive.ObjectID in locals
	testApp.Get("/zero-objectid",
		func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, primitive.NilObjectID); return ctx.Next() },
		rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:read"}, rbac.MatchAtLeastOne),
		successHandler,
	)


	exitVal := m.Run()
	
	// Teardown: Clean up RBAC data for these specific test users
	cleanupMwTestData()

	os.Exit(exitVal)
}

func setupTestRBACData() {
	// Define Roles
	goe.UseRBAC().DefineRole(rbac.Role{Name: "mw_admin", Permissions: []rbac.Permission{"article:read", "article:edit", "article:publish", "article:review"}})
	goe.UseRBAC().DefineRole(rbac.Role{Name: "mw_editor", Permissions: []rbac.Permission{"article:read", "article:edit", "article:publish"}}) // Does not have article:review
	goe.UseRBAC().DefineRole(rbac.Role{Name: "mw_viewer", Permissions: []rbac.Permission{"article:read"}})

	// Assign roles to test users
	err := goe.UseRBAC().AssignRoleToUser(mwTestUserAdmin, "mw_admin")
	if err != nil { fmt.Printf("Failed to assign mw_admin role: %v\n", err) } // Added \n
	
	err = goe.UseRBAC().AssignRoleToUser(mwTestUserEditor, "mw_editor")
	if err != nil { fmt.Printf("Failed to assign mw_editor role: %v\n", err) } // Added \n
	// Grant editor a direct permission for a specific test
	err = goe.UseRBAC().GrantDirectPermission(mwTestUserEditor, "feature:special")
	if err != nil { fmt.Printf("Failed to grant direct permission to mwTestUserEditor: %v\n", err) } // Added \n


	err = goe.UseRBAC().AssignRoleToUser(mwTestUserViewer, "mw_viewer")
	if err != nil { fmt.Printf("Failed to assign mw_viewer role: %v\n", err) } // Added \n
	
	// mwTestUserNoPerms has no roles or direct permissions assigned by default
}

func cleanupMwTestData() {
	fmt.Println("Cleaning up middleware test RBAC data...")
	// Check if GoeContainer and MongoDB are initialized
	if goe.UseLog() == nil { // Use a simple check like UseLog() to see if container is up
		fmt.Println("Middleware cleanup: Goe container not fully initialized, skipping DB cleanup.")
		return
	}
	mongoInstance := goe.UseMongo()
	if mongoInstance == nil {
		fmt.Println("Middleware cleanup: MongoDB not available.")
		return
	}

	usersToClean := []primitive.ObjectID{mwTestUserAdmin, mwTestUserEditor, mwTestUserViewer, mwTestUserNoPerms}

	for _, userID := range usersToClean {
		// Could call rbac.RevokeRoleFromUser for all known test roles,
		// but direct DB deletion is more thorough for cleanup.
		// Need to use the actual model types here for DeleteMany
		_, err := mongoInstance.DeleteMany(&models.UserRoleAssignment{}, bson.M{"user_id": userID})
		if err != nil {fmt.Printf("MW Cleanup error (UserRoleAssignment) for %s: %v\n", userID.Hex(), err)} // Added \n
		
		_, err = mongoInstance.DeleteMany(&models.UserDirectPermission{}, bson.M{"user_id": userID})
		if err != nil {fmt.Printf("MW Cleanup error (UserDirectPermission) for %s: %v\n", userID.Hex(), err)} // Added \n
	}
	fmt.Println("Middleware test RBAC data cleaned.")
}


func TestRBACMiddleware_AccessGranted_Viewer(t *testing.T) {
	req := httptest.NewRequest("GET", "/view-article", nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Viewer should access /view-article")
}

func TestRBACMiddleware_AccessGranted_Editor(t *testing.T) {
	req := httptest.NewRequest("GET", "/edit-article", nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Editor should access /edit-article")
}

func TestRBACMiddleware_AccessGranted_Admin_MatchAll(t *testing.T) {
	req := httptest.NewRequest("GET", "/publish-reviewed-article", nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin should access /publish-reviewed-article with MatchAll")
}

func TestRBACMiddleware_AccessDenied_Editor_MatchAll_MissingOne(t *testing.T) {
	req := httptest.NewRequest("GET", "/publish-reviewed-article-editor-fail", nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Editor should be denied /publish-reviewed-article-editor-fail due to missing 'article:review'")
}


func TestRBACMiddleware_AccessDenied_NoPermissions(t *testing.T) {
	req := httptest.NewRequest("GET", "/no-perms-test", nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "User with no permissions should be denied")
}

func TestRBACMiddleware_AccessDenied_MissingUserID(t *testing.T) {
	req := httptest.NewRequest("GET", "/missing-userid", nil)
	resp, _ := testApp.Test(req)
	// The default fiber error handler might return 500 if Locals returns nil and it's not caught gracefully
	// Our middleware explicitly returns a Forbidden if userIDVal is nil.
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Request with missing userID should be denied")
}

func TestRBACMiddleware_AccessDenied_InvalidUserIDFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/invalid-userid-format", nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Request with invalid userID format should be denied")
}

func TestRBACMiddleware_AccessDenied_ZeroObjectID(t *testing.T) {
	req := httptest.NewRequest("GET", "/zero-objectid", nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Request with Zero ObjectID should be denied")
}


// Test direct permission for editor
func TestRBACMiddleware_DirectPermission_Editor(t *testing.T) {
	// Define a route specifically for the direct permission test
	testApp.Get("/feature-special-direct",
		func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserEditor); return ctx.Next() },
		rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"feature:special"}, rbac.MatchAtLeastOne),
		func(ctx fiber.Ctx) error { return webresult.SendSucceed(ctx, "Direct Access Granted") },
	)
	req := httptest.NewRequest("GET", "/feature-special-direct", nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Editor should access /feature-special-direct via direct permission")
}


// Test MatchAtLeastOne: User has one of the required permissions
func TestRBACMiddleware_MatchAtLeastOne_HasOne(t *testing.T) {
    // mwTestUserViewer has "article:read"
    testApp.Get("/atleastone-hasone",
        func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserViewer); return ctx.Next() },
        rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:read", "article:write"}, rbac.MatchAtLeastOne),
        func(ctx fiber.Ctx) error { return webresult.SendSucceed(ctx, "Access Granted - AtLeastOne") },
    )
    req := httptest.NewRequest("GET", "/atleastone-hasone", nil)
    resp, _ := testApp.Test(req)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test MatchAtLeastOne: User has none of the required permissions
func TestRBACMiddleware_MatchAtLeastOne_HasNone(t *testing.T) {
    // mwTestUserNoPerms has no permissions
    testApp.Get("/atleastone-hasnone",
        func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserNoPerms); return ctx.Next() },
        rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:delete", "article:write"}, rbac.MatchAtLeastOne),
        func(ctx fiber.Ctx) error { return webresult.SendSucceed(ctx, "Access Granted - AtLeastOne") }, // Should not reach here
    )
    req := httptest.NewRequest("GET", "/atleastone-hasnone", nil)
    resp, _ := testApp.Test(req)
    assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// Test MatchAll: User has all required (single permission)
func TestRBACMiddleware_MatchAll_HasAll_Single(t *testing.T) {
    // mwTestUserViewer has "article:read"
    testApp.Get("/matchall-hassingle",
        func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserViewer); return ctx.Next() },
        rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:read"}, rbac.MatchAll),
        func(ctx fiber.Ctx) error { return webresult.SendSucceed(ctx, "Access Granted - MatchAll Single") },
    )
    req := httptest.NewRequest("GET", "/matchall-hassingle", nil)
    resp, _ := testApp.Test(req)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test MatchAll: User is missing one of multiple required
func TestRBACMiddleware_MatchAll_MissingOne_Multiple(t *testing.T) {
    // mwTestUserViewer has "article:read" but not "article:write"
    testApp.Get("/matchall-missingmultiple",
        func(ctx fiber.Ctx) error { ctx.Locals(testUserIDKey, mwTestUserViewer); return ctx.Next() },
        rbacMw.CheckPermission(testUserIDKey, []rbac.Permission{"article:read", "article:write"}, rbac.MatchAll),
        func(ctx fiber.Ctx) error { return webresult.SendSucceed(ctx, "Access Granted - MatchAll Multiple") }, // Should not reach
    )
    req := httptest.NewRequest("GET", "/matchall-missingmultiple", nil)
    resp, _ := testApp.Test(req)
    assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// Helper to deal with potential fiber.Error in custom error handler
func errorsAs(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	// This is a simplified version. In go 1.13+ use errors.As
	// For older versions, you might need more complex type assertion.
	// For this context, assuming fiber.Error is the main concern.
	_, ok := err.(*fiber.Error)
	return ok
}
```

package rbac

import (
	"fmt"
	"go.oease.dev/goe/core"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson" // Required for cleanupMwTestData
	"go.oease.dev/goe"
	"go.oease.dev/goe/webresult"
)

var testApp *fiber.App
var rbacMw *RBACMiddleware

// Test User IDs for middleware tests
var mwTestUserAdmin = "TestUserAdmin"
var mwTestUserEditor = "TestUserEditor"
var mwTestUserViewer = "TestUserViewer"
var mwTestUserNoPerms = "TestUserNoPerms"

var testUser1ID = "testUser1ID"
var testUser2ID = "testUser2ID"

func TestMain(m *testing.M) {
	setWorkingDir()

	err := goe.NewApp()
	if err != nil {
		panic(err)
	}

	// Define roles and assign permissions for test users
	// These interact with the actual RBAC service, potentially hitting the DB.
	// Ensure DB is clean or use unique user IDs for each test run if state is an issue.
	// The rbac_test.go TestMain should handle general cleanup if using the same DB.
	err = setupTestRBACData()
	if err != nil {
		fmt.Println("Failed to setup test RBAC data for middleware tests:", err)
	}
	goe.UseLog().Info("Test RBAC Data defined.")

	// Fiber App
	testApp = goe.UseFiber().App()
	rbacMw = NewRBACMiddleware()
	rbacMw.SetUserIdGetter(GetDefaultUserIdGetter)

	// Define a common success handler for protected routes
	successHandler := func(ctx fiber.Ctx) error {
		return webresult.SendSucceed(ctx, "Access Granted")
	}

	// optional, this is for test
	testApp.Use(setUserIDMiddleware())

	// Setup test routes
	// Route requiring "article:read" (Viewer)
	testApp.Get("/view-article",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:read"}, MatchAtLeastOne),
	)

	// Route requiring "article:edit" (Editor)
	testApp.Get("/edit-article",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:edit"}, MatchAtLeastOne),
	)

	// Route requiring "article:publish" AND "article:review" (Admin via role, Editor needs direct for review)
	testApp.Get("/publish-reviewed-article",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:publish", "article:review"}, MatchAll),
	)
	testApp.Get("/publish-reviewed-article-editor-fail", // Editor lacks article:review
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:publish", "article:review"}, MatchAll),
	)

	// Route for user with no permissions
	testApp.Get("/no-perms-test",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:read"}, MatchAtLeastOne),
	)

	// Route for missing userID in locals
	testApp.Get("/missing-userid",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:read"}, MatchAtLeastOne),
	)

	// Route for invalid userID format in locals
	testApp.Get("/invalid-userid-format",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:read"}, MatchAtLeastOne),
	)

	// Define a route specifically for the direct permission test
	testApp.Get("/feature-special-direct",
		successHandler,
		rbacMw.CheckPermission([]Permission{"feature:special"}, MatchAtLeastOne),
	)

	// mwTestUserViewer has "article:read"
	testApp.Get("/atleastone-hasone",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:read", "article:write"}, MatchAtLeastOne),
	)

	// mwTestUserNoPerms has no permissions
	testApp.Get("/atleastone-hasnone",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:delete", "article:write"}, MatchAtLeastOne),
	)

	// mwTestUserViewer has "article:read"
	testApp.Get("/matchall-hassingle",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:read"}, MatchAll),
	)

	// mwTestUserViewer has "article:read" but not "article:write"
	testApp.Get("/matchall-missingmultiple",
		successHandler,
		rbacMw.CheckPermission([]Permission{"article:read", "article:write"}, MatchAll),
	)

	exitVal := m.Run()

	// Teardown: Clean up RBAC data for these specific test users
	cleanupMwTestData()
	os.Exit(exitVal)
}

func setWorkingDir() {
	// 获取当前文件的路径
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get current file path")
	}

	// 从 /goe/middleware/rbac 跳转到项目根目录 /goe
	rootPath := filepath.Join(filepath.Dir(currentFilePath), "../../")

	// 切换工作目录到根目录
	err := os.Chdir(rootPath)
	if err != nil {
		panic("Failed to change working directory: %v")
	}
}

func setUserIDMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID := c.Query(UserIDKey)
		if userID == "" {
			return c.Status(fiber.StatusForbidden).SendString("Missing user ID")
		}
		c.Locals(UserIDKey, userID)
		fmt.Println(c.Locals(UserIDKey))
		return c.Next()
	}
}

func setupTestRBACData() error {
	// Define Roles
	err := DefineRole(&Role{Name: "mw_admin", Permissions: []Permission{"article:read", "article:edit", "article:publish", "article:review"}})
	if err != nil {
		return err
	}
	err = DefineRole(&Role{Name: "mw_editor", Permissions: []Permission{"article:read", "article:edit", "article:publish"}})
	if err != nil {
		return err
	} // Does not have article:review
	err = DefineRole(&Role{Name: "mw_viewer", Permissions: []Permission{"article:read"}})
	if err != nil {
		return err
	}

	// Define some roles for testing
	err = DefineRole(&Role{Name: "test_admin", Permissions: []Permission{"user:create", "user:read"}})
	if err != nil {
		return err
	}
	err = DefineRole(&Role{Name: "test_editor", Permissions: []Permission{"article:create", "article:read"}})
	if err != nil {
		return err
	}
	err = DefineRole(&Role{Name: "test_viewer", Permissions: []Permission{"article:read"}})
	if err != nil {
		return err
	}

	// Assign roles to test users
	err = AssignRoleToUser(mwTestUserAdmin, "mw_admin")
	if err != nil {
		fmt.Printf("Failed to assign mw_admin role: %v\n", err)
	}

	err = AssignRoleToUser(mwTestUserEditor, "mw_editor")
	if err != nil {
		fmt.Printf("Failed to assign mw_editor role: %v\n", err)
	}
	// Grant editor a direct permission for a specific test
	err = GrantDirectPermission(mwTestUserEditor, "feature:special")
	if err != nil {
		fmt.Printf("Failed to grant direct permission to mwTestUserEditor: %v\n", err)
	}

	err = AssignRoleToUser(mwTestUserViewer, "mw_viewer")
	if err != nil {
		fmt.Printf("Failed to assign mw_viewer role: %v\n", err)
	}

	return nil
}

func cleanupMwTestData() {
	fmt.Println("Cleaning up middleware test RBAC data...")
	// Check if GoeContainer and MongoDB are initialized
	if goe.UseLog() == nil { // Use a simple check like UseLog() to see if container is up
		fmt.Println("Middleware cleanup: Goe container not fully initialized, skipping DB cleanup.")
		return
	}
	mongoInstance := goe.UseDB()
	if mongoInstance == nil {
		fmt.Println("Middleware cleanup: MongoDB not available.")
		return
	}

	usersToClean := []string{mwTestUserAdmin, mwTestUserEditor, mwTestUserViewer, mwTestUserNoPerms, testUser1ID, testUser2ID}

	// Delete from user_role_assignments
	roleAssignmentModel := &UserRoleAssignment{}
	_, err := mongoInstance.DeleteMany(roleAssignmentModel, bson.M{"user_id": bson.M{"$in": usersToClean}})
	if err != nil {
		fmt.Printf("Error cleaning up UserRoleAssignments: %v\n", err) // Added newline
	} else {
		fmt.Println("UserRoleAssignments cleaned.")
	}

	// Delete from user_direct_permissions
	directPermissionModel := &UserDirectPermission{}
	_, err = mongoInstance.DeleteMany(directPermissionModel, bson.M{"user_id": bson.M{"$in": usersToClean}})
	if err != nil {
		fmt.Printf("Error cleaning up UserDirectPermissions: %v\n", err) // Added newline
	} else {
		fmt.Println("UserDirectPermissions cleaned.")
	}

	// delete test roles
	role := &Role{}
	rolesToClean := []string{"mw_admin", "mw_editor", "mw_viewer", "test_admin", "test_editor", "test_viewer"}
	_, err = mongoInstance.DeleteMany(role, bson.M{"name": bson.M{"$in": rolesToClean}})
	if err != nil {
		fmt.Printf("Error cleaning up Roles: %v\n", err) // Added newline
	} else {
		fmt.Println("Roles cleaned.")
	}

	fmt.Println("Middleware test RBAC data cleaned.")
}

func TestRBACMiddleware_AccessGranted_Viewer(t *testing.T) {
	req := httptest.NewRequest("GET", "/view-article"+"?user_id="+mwTestUserViewer, nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Viewer should access /view-article")
}

func TestRBACMiddleware_AccessGranted_Editor(t *testing.T) {
	req := httptest.NewRequest("GET", "/edit-article"+"?user_id="+mwTestUserEditor, nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Editor should access /edit-article")
}

func TestRBACMiddleware_AccessGranted_Admin_MatchAll(t *testing.T) {
	req := httptest.NewRequest("GET", "/publish-reviewed-article"+"?user_id="+mwTestUserAdmin, nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin should access /publish-reviewed-article with MatchAll")
}

func TestRBACMiddleware_AccessDenied_Editor_MatchAll_MissingOne(t *testing.T) {
	req := httptest.NewRequest("GET", "/publish-reviewed-article-editor-fail"+"?user_id="+mwTestUserEditor, nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Editor should be denied /publish-reviewed-article-editor-fail due to missing 'article:review'")
}

func TestRBACMiddleware_AccessDenied_NoPermissions(t *testing.T) {
	req := httptest.NewRequest("GET", "/no-perms-test"+"?user_id="+mwTestUserNoPerms, nil)
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
	req := httptest.NewRequest("GET", "/invalid-userid-format"+"?user_id="+"not-an-user-id", nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Request with invalid userID format should be denied")
}

// Test direct permission for editor
func TestRBACMiddleware_DirectPermission_Editor(t *testing.T) {
	req := httptest.NewRequest("GET", "/feature-special-direct"+"?user_id="+mwTestUserEditor, nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Editor should access /feature-special-direct via direct permission")
}

// Test MatchAtLeastOne: User has one of the required permissions
func TestRBACMiddleware_MatchAtLeastOne_HasOne(t *testing.T) {
	req := httptest.NewRequest("GET", "/atleastone-hasone"+"?user_id="+mwTestUserViewer, nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test MatchAtLeastOne: User has none of the required permissions
func TestRBACMiddleware_MatchAtLeastOne_HasNone(t *testing.T) {
	req := httptest.NewRequest("GET", "/atleastone-hasnone"+"?user_id="+mwTestUserNoPerms, nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// Test MatchAll: User has all required (single permission)
func TestRBACMiddleware_MatchAll_HasAll_Single(t *testing.T) {
	req := httptest.NewRequest("GET", "/matchall-hassingle"+"?user_id="+mwTestUserViewer, nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// Test MatchAll: User is missing one of multiple required
func TestRBACMiddleware_MatchAll_MissingOne_Multiple(t *testing.T) {
	req := httptest.NewRequest("GET", "/matchall-missingmultiple"+"?user_id="+mwTestUserViewer, nil)
	resp, _ := testApp.Test(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRoleDefinition(t *testing.T) {
	// Roles are defined in TestMain.
	role, found := GetRole("test_admin")
	assert.True(t, found, "Expected 'test_admin' role to be found")
	assert.Equal(t, "test_admin", role.Name)
	assert.Contains(t, role.Permissions, Permission("user:read"))

	_, found = GetRole("non_existent_role")
	assert.False(t, found, "Expected 'non_existent_role' not to be found")

	// Test defining a role with empty name (should be error DefineRole)
	err := DefineRole(&Role{Name: "", Permissions: []Permission{"p:a"}})
	assert.Error(t, err, "Error defining role")

	// test get roles by empty role name
	role, found = GetRole("")
	assert.False(t, found, "Role with empty name was defined, should be found")
	assert.Nil(t, role)

	// Test defining a role with nil/empty permissions (should be Error)
	err = DefineRole(&Role{Name: "empty_perm_role", Permissions: nil})
	assert.Error(t, err)
}

func TestRoleAssignment(t *testing.T) {
	// Ensure user is clean for this test
	cleanupUser(testUser1ID)

	// Assign role
	err := AssignRoleToUser(testUser1ID, "test_editor")
	assert.NoError(t, err)

	// Assign same role again (should be idempotent)
	err = AssignRoleToUser(testUser1ID, "test_editor")
	assert.NoError(t, err)

	// Get assigned roles
	roles, err := GetUserRoleNames(testUser1ID)
	assert.NoError(t, err)
	assert.Contains(t, roles, "test_editor")
	assert.Len(t, roles, 1, "Expected only one role assignment for test_editor after idempotent calls")

	// Assign another role
	err = AssignRoleToUser(testUser1ID, "test_viewer")
	assert.NoError(t, err)
	roles, err = GetUserRoleNames(testUser1ID)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"test_editor", "test_viewer"}, roles)

	// Revoke role
	err = RevokeRoleFromUser(testUser1ID, "test_editor")
	assert.NoError(t, err)
	roles, err = GetUserRoleNames(testUser1ID)
	assert.NoError(t, err)
	assert.NotContains(t, roles, "test_editor")
	assert.Contains(t, roles, "test_viewer")

	// Revoke non-existent role for user
	err = RevokeRoleFromUser(testUser1ID, "test_admin") // was never assigned
	assert.NoError(t, err)                              // should not error

	// Revoke last role
	err = RevokeRoleFromUser(testUser1ID, "test_viewer")
	assert.NoError(t, err)
	roles, err = GetUserRoleNames(testUser1ID)
	assert.NoError(t, err)
	assert.Empty(t, roles)

	// Test with blank string
	err = AssignRoleToUser("", "test_admin")
	assert.Error(t, err, "Assigning role to blank string should error")
}

func TestDirectPermissions(t *testing.T) {
	// Ensure user is clean for this test
	cleanupUser(testUser2ID)

	// Grant permission
	perm1 := Permission("article:publish")
	err := GrantDirectPermission(testUser2ID, perm1)
	assert.NoError(t, err)

	// Grant same permission again (should be idempotent due to $addToSet)
	err = GrantDirectPermission(testUser2ID, perm1)
	assert.NoError(t, err)

	// Get direct permissions
	perms, err := GetUserDirectPermissions(testUser2ID)
	assert.NoError(t, err)
	assert.Contains(t, perms, perm1)
	assert.Len(t, perms, 1, "Expected only one direct permission after idempotent calls")

	// Grant another permission
	perm2 := Permission("user:delete")
	err = GrantDirectPermission(testUser2ID, perm2)
	assert.NoError(t, err)
	perms, err = GetUserDirectPermissions(testUser2ID)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []Permission{perm1, perm2}, perms)

	// Revoke permission
	err = RevokeDirectPermission(testUser2ID, perm1)
	assert.NoError(t, err)
	perms, err = GetUserDirectPermissions(testUser2ID)
	assert.NoError(t, err)
	assert.NotContains(t, perms, perm1)
	assert.Contains(t, perms, perm2)

	// Revoke non-existent permission for user
	err = RevokeDirectPermission(testUser2ID, Permission("non:existent"))
	assert.NoError(t, err) // should not error

	// Revoke last permission
	err = RevokeDirectPermission(testUser2ID, perm2)
	assert.NoError(t, err)
	perms, err = GetUserDirectPermissions(testUser2ID)
	assert.NoError(t, err)
	assert.Empty(t, perms)

	// Test with blank string
	err = GrantDirectPermission("", "test:perm")
	assert.Error(t, err, "Granting permission to blank string should error")
}

func TestGetAllUserPermissions(t *testing.T) {
	// Setup: User1 gets "test_admin" role. User2 gets "test_editor" role and direct "article:feature" perm.
	// Cleanup previous assignments for these users for this specific test
	cleanupUser(testUser1ID)
	cleanupUser(testUser2ID)

	// Assign roles and permissions
	err := AssignRoleToUser(testUser1ID, "test_admin") // user:create, user:read
	assert.NoError(t, err)

	err = AssignRoleToUser(testUser2ID, "test_editor") // article:create, article:read
	assert.NoError(t, err)
	directPermUser2 := Permission("article:feature")
	err = GrantDirectPermission(testUser2ID, directPermUser2)
	assert.NoError(t, err)
	err = GrantDirectPermission(testUser2ID, Permission("article:read")) // Grant a direct perm that also exists in role
	assert.NoError(t, err)

	// Test User1 (only role-based)
	user1Perms, err := GetAllUserPermissions(testUser1ID)
	assert.NoError(t, err)
	expectedUser1Perms := []Permission{"user:create", "user:read"}
	assert.ElementsMatch(t, expectedUser1Perms, user1Perms)

	// Test User2 (role-based + direct, with overlap)
	user2Perms, err := GetAllUserPermissions(testUser2ID)
	assert.NoError(t, err)
	expectedUser2Perms := []Permission{"article:create", "article:read", "article:feature"}
	assert.ElementsMatch(t, expectedUser2Perms, user2Perms)

	// Test user with no roles or direct permissions
	freshUserID := "freshUserID"
	// Ensure this user is clean in case it was used in a previous failed run or other test
	cleanupUser(freshUserID)
	noPerms, err := GetAllUserPermissions(freshUserID)
	assert.NoError(t, err)
	assert.Empty(t, noPerms)
}

// cleanupUser is a helper to clear a specific user's RBAC data
func cleanupUser(userID string) {
	// Check if container and mongo are available, especially if tests run individually or TestMain had issues.
	if core.UseGoeContainer() == nil || core.UseGoeContainer().GetMongo() == nil {
		fmt.Printf("Skipping cleanupUser for %s: MongoDB connection not available.\n", userID)
		return
	}
	mongoClient := core.UseGoeContainer().GetMongo()
	roleAssignmentModel := &UserRoleAssignment{}
	_, _ = mongoClient.DeleteMany(roleAssignmentModel, bson.M{"user_id": userID})
	directPermissionModel := &UserDirectPermission{}
	_, _ = mongoClient.DeleteMany(directPermissionModel, bson.M{"user_id": userID})
}

// Small test for GetUserRoleNames when no roles are assigned (should be empty, not error)
func TestGetUserRoleNames_NoRoles(t *testing.T) {
	freshUserID := "freshUserID"
	cleanupUser(freshUserID) // Ensure clean state
	roles, err := GetUserRoleNames(freshUserID)
	assert.NoError(t, err)
	assert.Empty(t, roles)
}

// Small test for GetUserDirectPermissions when no direct perms (should be empty, not error)
func TestGetUserDirectPermissions_NoDirectPerms(t *testing.T) {
	freshUserID := "freshUserID"
	cleanupUser(freshUserID) // Ensure clean state
	perms, err := GetUserDirectPermissions(freshUserID)
	assert.NoError(t, err)
	assert.Empty(t, perms)
}

func TestDeleteRole(t *testing.T) {
	freshUserID := "freshUserID"
	cleanupUser(freshUserID)
	//should be empty, not error
	perms, err := GetUserDirectPermissions(freshUserID)
	assert.Empty(t, perms)

	role := &Role{
		Name:        "Expired",
		Permissions: []Permission{"expired:view", "expired:update"},
	}
	err = DefineRole(role)
	assert.NoError(t, err)

	err = AssignRoleToUser(freshUserID, role.Name)
	assert.NoError(t, err)

	permissions, err := GetAllUserPermissions(freshUserID)
	assert.NoError(t, err)
	assert.ElementsMatch(t, permissions, role.Permissions)

	err = DeleteRole(role.Name)
	assert.NoError(t, err)

	permissions, err = GetAllUserPermissions(freshUserID)
	assert.NoError(t, err)
	assert.Empty(t, permissions)

	r, found := GetRole(role.Name)
	assert.Nil(t, r)
	assert.False(t, found)
}

func TestListRoles(t *testing.T) {
	cleanupMwTestData()
	roles := []*Role{
		{
			Name:        "ExpiredOne",
			Permissions: []Permission{"expired:view", "expired:update"},
		},
		{
			Name:        "ExpiredTwo",
			Permissions: []Permission{"expired:create", "expired:delete"},
		},
		{
			Name:        "ExpiredThree",
			Permissions: []Permission{"expired:update", "expired:create"},
		},
	}
	for _, role := range roles {
		err := DefineRole(role)
		assert.NoError(t, err)
	}

	result, err := ListRoles(3, 1)
	var actual []*Role
	for _, r := range result {
		actual = append(actual, &Role{
			Name:        r.Name,
			Permissions: r.Permissions,
		})
	}
	assert.NoError(t, err)
	assert.ElementsMatch(t, actual, roles)

	err = DeleteRole(roles[0].Name)
	assert.NoError(t, err)
	err = DeleteRole(roles[1].Name)
	assert.NoError(t, err)
	err = DeleteRole(roles[2].Name)
	assert.NoError(t, err)
}

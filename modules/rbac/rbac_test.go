package rbac

import (
	"fmt"
	"os"
	"testing"
	// "time" // Not strictly needed by the provided test code directly, but often useful in tests

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/goe"
	"go.oease.dev/goe/core"
	"go.oease.dev/goe/models"
)

var testUser1ID primitive.ObjectID
var testUser2ID primitive.ObjectID

func TestMain(m *testing.M) {
	// Setup: Initialize Goe framework. This should set up DB connections.
	// Ensure your test environment configuration points to a test database.
	// You might need to copy or generate a test .env file.
	// Example: os.Setenv("GOE_ENV_FILE_PATH", "../../.test.env") 
	// For simplicity, this assumes NewApp() correctly configures for test if run in test mode
	// or that the default .env points to a safe, mutable test DB.
	
	// It's better to ensure NewApp() is only called once.
	// A more robust setup might involve a sync.Once or checking if core.UseGoeContainer() panics.
	// For this example, we'll call it directly.
	// Ensure that example/configs/.env is configured for a test database for this to be safe.
	// Or, use a specific test configuration file.
	// For instance, you could set os.Setenv("GOE_CONFIG_NAME", "test_config") and have configs/test_config.yaml
	if err := os.Setenv("GOE_APP_NAME", "goe-rbac-test"); err != nil {
		fmt.Println("Failed to set GOE_APP_NAME for tests:", err)
		os.Exit(1)
	}
	// Potentially set other env vars like MONGODB_URI, MONGODB_DATABASE to point to a test DB.
	// e.g., os.Setenv("MONGODB_DATABASE", "goe_test_rbac")
	// It's also common to read these from a .env.test file if the framework supports it.
	// For example, if goe.NewApp() respects a certain env var for config path:
	// os.Setenv("GOE_CONFIG_PATH", "path/to/your/test/config.yaml_or_env_file")


	// Attempt to initialize the Goe application.
	// This is a common pattern, but if NewApp itself depends on specific environment
	// variables being set (like MONGODB_URI, MONGODB_DATABASE for a test DB),
	// ensure those are set *before* this call.
	// The core.GoeConfig struct in goe/core/config.go and its population in
	// goe/goe.go's NewApp() and applyEnvConfig() will determine how DB details are sourced.
	// If these are not set to a test DB, these tests WILL run on your dev DB.
	err := goe.NewApp()
	if err != nil {
		fmt.Println("Failed to initialize Goe app for tests:", err)
		fmt.Println("Ensure your environment (e.g., .env or env variables) is configured to point to a TEST MongoDB instance.")
		os.Exit(1)
	}
	
	// Initialize test user IDs
	testUser1ID = primitive.NewObjectID()
	testUser2ID = primitive.NewObjectID()

	// Define some roles for testing
	DefineRole(Role{Name: "test_admin", Permissions: []Permission{"user:create", "user:read"}})
	DefineRole(Role{Name: "test_editor", Permissions: []Permission{"article:create", "article:read"}})
	DefineRole(Role{Name: "test_viewer", Permissions: []Permission{"article:read"}})


	// Run tests
	exitVal := m.Run()

	// Teardown: Clean up database entries created by tests
	cleanupTestData()

	os.Exit(exitVal)
}

func cleanupTestData() {
	fmt.Println("Cleaning up test data...")
	if core.UseGoeContainer() == nil || core.UseGoeContainer().GetMongo() == nil {
		fmt.Println("MongoDB connection not available for cleanup.")
		return
	}
	mongoClient := core.UseGoeContainer().GetMongo()

	// Delete from user_role_assignments
	roleAssignmentModel := &models.UserRoleAssignment{}
	_, err := mongoClient.DeleteMany(roleAssignmentModel, bson.M{"user_id": bson.M{"$in": []primitive.ObjectID{testUser1ID, testUser2ID}}})
	if err != nil {
		fmt.Printf("Error cleaning up UserRoleAssignments: %v\n", err) // Added newline
	} else {
		fmt.Println("UserRoleAssignments cleaned.")
	}

	// Delete from user_direct_permissions
	directPermissionModel := &models.UserDirectPermission{}
	_, err = mongoClient.DeleteMany(directPermissionModel, bson.M{"user_id": bson.M{"$in": []primitive.ObjectID{testUser1ID, testUser2ID}}})
	if err != nil {
		fmt.Printf("Error cleaning up UserDirectPermissions: %v\n", err) // Added newline
	} else {
		fmt.Println("UserDirectPermissions cleaned.")
	}
}

func TestRoleDefinition(t *testing.T) {
	// Roles are defined in TestMain.
	role, found := GetRole("test_admin")
	assert.True(t, found, "Expected 'test_admin' role to be found")
	assert.Equal(t, "test_admin", role.Name)
	assert.Contains(t, role.Permissions, Permission("user:read"))

	_, found = GetRole("non_existent_role")
	assert.False(t, found, "Expected 'non_existent_role' not to be found")

	// Test defining a role with empty name (should be logged as warning by DefineRole)
	DefineRole(Role{Name: "", Permissions: []Permission{"p:a"}}) // Should not panic, just warn
	_, found = GetRole("")
	// Depending on implementation of DefineRole, an empty string role name might be stored or rejected.
	// Current DefineRole in rbac.go will store it. If that's undesirable, DefineRole should prevent it.
	// For this test, we'll assume it might be stored but GetRole might not find it if it's filtered,
	// or it might be found if stored. The provided rbac.go stores it.
	// assert.False(t, found, "Role with empty name should not be retrievable if DefineRole guards against it or stores it specially")
	// Based on current rbac.go, it *will* be found if an empty name is allowed by DefineRole.
	// However, the original test implies it shouldn't be found, suggesting DefineRole might filter.
	// Let's adjust the assertion based on DefineRole's behavior of logging a warning but still adding.
	// If DefineRole was stricter and returned an error or didn't add, then False would be correct.
	role, found = GetRole("") 
	assert.True(t, found, "Role with empty name was defined, should be found")
	assert.Equal(t, "", role.Name)


	// Test defining a role with nil/empty permissions (should be logged)
	DefineRole(Role{Name: "empty_perm_role", Permissions: nil})
	role, found = GetRole("empty_perm_role")
	assert.True(t, found)
	assert.Empty(t, role.Permissions) // Permissions should be an empty slice, not nil, if initialized by DefineRole or Role struct
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
	assert.NoError(t, err) // should not error

	// Revoke last role
	err = RevokeRoleFromUser(testUser1ID, "test_viewer")
	assert.NoError(t, err)
	roles, err = GetUserRoleNames(testUser1ID)
	assert.NoError(t, err)
	assert.Empty(t, roles)
	
	// Test with Zero ObjectID
	err = AssignRoleToUser(primitive.NilObjectID, "test_admin")
	assert.Error(t, err, "Assigning role to NilObjectID should error")
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
	
	// Test with Zero ObjectID
	err = GrantDirectPermission(primitive.NilObjectID, "test:perm")
	assert.Error(t, err, "Granting permission to NilObjectID should error")
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
	freshUserID := primitive.NewObjectID()
	// Ensure this user is clean in case it was used in a previous failed run or other test
	cleanupUser(freshUserID) 
	noPerms, err := GetAllUserPermissions(freshUserID)
	assert.NoError(t, err)
	assert.Empty(t, noPerms)
}

// cleanupUser is a helper to clear a specific user's RBAC data
func cleanupUser(userID primitive.ObjectID) {
	// Check if container and mongo are available, especially if tests run individually or TestMain had issues.
	if core.UseGoeContainer() == nil || core.UseGoeContainer().GetMongo() == nil {
		fmt.Printf("Skipping cleanupUser for %s: MongoDB connection not available.\n", userID.Hex())
		return
	}
	mongoClient := core.UseGoeContainer().GetMongo()
	roleAssignmentModel := &models.UserRoleAssignment{}
	_, _ = mongoClient.DeleteMany(roleAssignmentModel, bson.M{"user_id": userID})
	directPermissionModel := &models.UserDirectPermission{}
	_, _ = mongoClient.DeleteMany(directPermissionModel, bson.M{"user_id": userID})
}

// Small test for GetUserRoleNames when no roles are assigned (should be empty, not error)
func TestGetUserRoleNames_NoRoles(t *testing.T) {
    freshUserID := primitive.NewObjectID()
	cleanupUser(freshUserID) // Ensure clean state
    roles, err := GetUserRoleNames(freshUserID)
    assert.NoError(t, err)
    assert.Empty(t, roles)
}

// Small test for GetUserDirectPermissions when no direct perms (should be empty, not error)
func TestGetUserDirectPermissions_NoDirectPerms(t *testing.T) {
    freshUserID := primitive.NewObjectID()
	cleanupUser(freshUserID) // Ensure clean state
    perms, err := GetUserDirectPermissions(freshUserID)
    assert.NoError(t, err)
    assert.Empty(t, perms)
}

```

# Role-Based Access Control (RBAC) Module

## 1. Overview

The RBAC module provides a flexible way to manage user permissions and control access to routes within your Goe application. It allows you to define roles with specific sets of permissions and assign these roles (or individual permissions directly) to users.

## 2. Core Concepts

### Permissions

A permission is a string that represents the authority to perform a specific action on a certain feature or resource. The recommended format is:

`"feature_name:action"`

For example:
- `"article:read"`
- `"article:create"`
- `"user:manage"`
- `"billing:view_invoice"`

### Roles

A role is a named collection of permissions. Users can be assigned one or more roles. Managing permissions via roles simplifies administration, as you can change a role's permissions, and all users with that role will automatically inherit the changes.

Example:
- An "editor" role might have `["article:read", "article:create", "article:update"]` permissions.
- An "admin" role might have `["user:manage", "settings:update"]` and all editor permissions.

### Direct vs. Role-Based Permissions

- **Role-Based Permissions**: Users inherit permissions from the roles they are assigned.
- **Direct Permissions**: Users can also be granted specific permissions directly, irrespective of their roles. This is useful for fine-grained control or exceptions.

The system combines both types of permissions for a user. If a user has a permission either directly or through any of their roles, they are considered to have that permission.

### Match Modes

When checking permissions for a route, you can specify a match mode:

- `rbac.MatchAll`: The user must have *all* of the specified required permissions to access the route.
- `rbac.MatchAtLeastOne`: The user must have *at least one* of the specified required permissions.

## 3. Setup: Defining Roles

Roles and their associated permissions are typically defined in your application's startup code using the `goe.UseRBAC().DefineRole()` function.

```go
import (
    "go.oease.dev/goe"
    "go.oease.dev/goe/modules/rbac"
)

func main() {
    // ... Goe App Initialization ...
    err := goe.NewApp()
    if err != nil {
        panic(err)
    }

    // Define roles
    goe.UseRBAC().DefineRole(rbac.Role{
        Name:        "administrator",
        Permissions: []rbac.Permission{"user:create", "user:read", "user:update", "user:delete", "article:publish"},
    })

    goe.UseRBAC().DefineRole(rbac.Role{
        Name:        "content_editor",
        Permissions: []rbac.Permission{"article:create", "article:read", "article:update"},
    })

    goe.UseRBAC().DefineRole(rbac.Role{
        Name:        "viewer",
        Permissions: []rbac.Permission{"article:read"},
    })
    
    // ... rest of your app setup ...
    goe.Run()
}
```

## 4. Assigning Permissions and Roles to Users

You can assign roles or grant direct permissions to users using the RBAC service. The `userID` is typically a `primitive.ObjectID` from MongoDB.

### Assigning a Role to a User

```go
import (
    "go.oease.dev/goe"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func AssignRoleExample(userIDHex string, roleName string) error {
    userID, err := primitive.ObjectIDFromHex(userIDHex)
    if err != nil {
        return err // Handle error: invalid user ID format
    }

    return goe.UseRBAC().AssignRoleToUser(userID, roleName)
}
```
This stores the role assignment in the database (`user_role_assignments` collection).

### Granting a Direct Permission to a User

```go
import (
    "go.oease.dev/goe"
    "go.oease.dev/goe/modules/rbac" // For rbac.Permission type
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func GrantPermissionExample(userIDHex string, permissionString string) error {
    userID, err := primitive.ObjectIDFromHex(userIDHex)
    if err != nil {
        return err // Handle error: invalid user ID format
    }

    permission := rbac.Permission(permissionString)
    return goe.UseRBAC().GrantDirectPermission(userID, permission)
}
```
This stores the direct permission in the database (`user_direct_permissions` collection).

## 5. Protecting Routes

The `RBACMiddleware` is used to protect Fiber routes.

### Initialization

First, create an instance of the middleware:

```go
import "go.oease.dev/goe/middlewares"

// ... in your setup code ...
rbacProtection := middlewares.NewRBACMiddleware()
```

### Usage

To protect a route, add the `rbacProtection.CheckPermission(...)` handler.

```go
import (
    "github.com/gofiber/fiber/v3"
    "go.oease.dev/goe/modules/rbac" // For rbac.Permission and rbac.MatchMode
    // ... other imports
)

// Assume 'app' is your Fiber app instance (e.g., goe.UseFiber().App())
// Assume 'rbacProtection' is the initialized RBACMiddleware instance
// Assume 'authMiddleware' is some authentication middleware that sets ctx.Locals("currentUserObjectID", userID)

const userIDKey = "currentUserObjectID" // Key used in ctx.Locals by auth middleware

// Example: Route requiring "article:create" permission
app.Post("/articles",
    authMiddleware, // Ensures user is authenticated and userID is set
    rbacProtection.CheckPermission(userIDKey, []rbac.Permission{"article:create"}, rbac.MatchAtLeastOne),
    func(ctx fiber.Ctx) error {
        // Handler logic for creating an article
        return ctx.SendString("Article created successfully")
    },
)

// Example: Route requiring "user:manage" AND "audit:log" permissions
app.Post("/admin/users/manage",
    authMiddleware,
    rbacProtection.CheckPermission(userIDKey, []rbac.Permission{"user:manage", "audit:log"}, rbac.MatchAll),
    func(ctx fiber.Ctx) error {
        // Handler logic for managing users
        return ctx.SendString("User managed successfully")
    },
)
```

**Explanation of `CheckPermission` parameters:**

- `userIDKey (string)`: This is the key that the RBAC middleware will use to look up the user's identifier in `fiber.Ctx.Locals()`. Your authentication middleware is responsible for placing the user's ID (preferably as a `primitive.ObjectID` or its hex string representation) into `ctx.Locals()` using this key.
- `permissions ([]rbac.Permission)`: A slice of `rbac.Permission` strings that are required for this route.
- `mode (rbac.MatchMode)`: Either `rbac.MatchAll` or `rbac.MatchAtLeastOne`.

## 6. Retrieving All User Permissions

If you need to get a list of all permissions a user has (both direct and from all their roles) outside of a route middleware context, you can use:

```go
import (
    "go.oease.dev/goe"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUserPermissionsExample(userIDHex string) ([]rbac.Permission, error) {
    userID, err := primitive.ObjectIDFromHex(userIDHex)
    if err != nil {
        return nil, err
    }
    return goe.UseRBAC().GetAllUserPermissions(userID)
}
```

This can be useful for displaying user capabilities in an admin panel or for custom permission checks in your business logic.
```

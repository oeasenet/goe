package middlewares

import (
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/webresult"
	"strings"
)

// NewLoginCheckMiddleware creates a middleware function that checks if a user is logged in.
// If the skipRoutes parameter is provided, the middleware will skip the check for those routes.
// The middleware checks if the user is logged in using the IsLoggedIn function.
// If the user is logged in, the middleware calls the next handler.
// If the user is not logged in, the middleware returns an Unauthorized response.
// The IsLoggedIn function checks if a session exists with a valid ID and keys.
// The isSkippedRoute function checks if the method and path match any of the patterns in skipRoutes.
// The parsePattern function parses a pattern string into a routePattern.
// The Unauthorized function creates an Unauthorized error response.
// The UseSession function retrieves the session from the context.
// The routePattern struct represents a route pattern with an HTTP method and path pattern.
// The initSessionStore function initializes the session store.
// This middleware is used to protect routes that require authentication, can be used as global middleware.
func NewLoginCheckMiddleware(skipRoutes ...[]string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if len(skipRoutes) != 0 {
			if isSkippedRoute(ctx.Method(), ctx.Path(), skipRoutes[0]) {
				return ctx.Next()
			}
		}

		if IsLoggedIn(ctx) {
			return ctx.Next()
		}

		return webresult.Unauthorized()
	}
}

// NewLoginInfoMiddleware creates a middleware function that checks if a user is logged in.
// The middleware checks if the user is logged in using the IsLoggedIn function.
// If the user is logged in, the middleware retrieves the user data from the session.
// It then unmarshals the user data into a map[string]interface{}.
// If unmarshaling fails, it returns a SystemBusy error with the unmarshal error.
// Otherwise, it returns a success response with the user info as the data.
// If the user data is nil, it returns a success response without any data.
// If the user is not logged in, the middleware returns an Unauthorized response.
// This middleware is used when the client needs to retrieve the user info.
func NewLoginInfoMiddleware() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if IsLoggedIn(ctx) {
			userData := UseSession(ctx).Get("user")
			if userData != nil {
				userInfo := make(map[string]any)
				err := json.Unmarshal(userData.([]byte), &userInfo)
				if err != nil {
					return webresult.SystemBusy(err)
				}
				return webresult.SendSucceed(ctx, userInfo)
			}
			return webresult.SendSucceed(ctx)
		}
		return webresult.Unauthorized()
	}
}

// Represents a route pattern with an HTTP method and path pattern.
type routePattern struct {
	method string
	path   []string
}

// Parses a pattern string (e.g., "GET /api/v1/file/:id") into a routePattern.
func parsePattern(patternStr string) routePattern {
	parts := strings.SplitN(patternStr, " ", 2)
	method := parts[0]
	pathSegments := strings.Split(strings.Trim(parts[1], "/"), "/")
	return routePattern{method, pathSegments}
}

// Checks if the given method and path match any of the patterns in skippedRoutes.
func isSkippedRoute(method, path string, skippedRoutes []string) bool {
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	for _, patternStr := range skippedRoutes {
		pattern := parsePattern(patternStr)
		if method != pattern.method || len(pattern.path) != len(pathSegments) {
			continue
		}

		match := true
		for i, segment := range pattern.path {
			if segment != pathSegments[i] && !strings.HasPrefix(segment, ":") {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}
	return false
}

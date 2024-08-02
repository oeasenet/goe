package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe/webresult"
	"strings"
)

func NewCheckAuthMiddleware(skipRoutes ...[]string) fiber.Handler {
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

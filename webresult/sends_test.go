package webresult

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	// Test that Map is a type alias for map[string]any
	m := Map{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	assert.Equal(t, "value1", m["key1"])
	assert.Equal(t, 42, m["key2"])
	assert.Equal(t, true, m["key3"])
}

func TestList(t *testing.T) {
	// Test that List is a type alias for []any
	l := List{
		"string",
		42,
		true,
		Map{"nested": "value"},
	}

	assert.Len(t, l, 4)
	assert.Equal(t, "string", l[0])
	assert.Equal(t, 42, l[1])
	assert.Equal(t, true, l[2])

	nestedMap, ok := l[3].(Map)
	assert.True(t, ok)
	assert.Equal(t, "value", nestedMap["nested"])
}

func TestWebResult(t *testing.T) {
	// Test WebResult struct
	result := WebResult{
		Message: "test message",
		Data:    Map{"key": "value"},
	}

	assert.Equal(t, "test message", result.Message)
	assert.NotNil(t, result.Data)

	data, ok := result.Data.(Map)
	assert.True(t, ok)
	assert.Equal(t, "value", data["key"])
}

func TestWebResult_JSON(t *testing.T) {
	// Test JSON marshaling/unmarshaling
	result := WebResult{
		Message: "test message",
		Data:    Map{"key": "value", "number": 42},
	}

	jsonData, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshaled WebResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, "test message", unmarshaled.Message)
	assert.NotNil(t, unmarshaled.Data)
}

func TestWebResult_JSONOmitEmpty(t *testing.T) {
	// Test that Data is omitted when empty
	result := WebResult{
		Message: "test message",
		// Data is nil/empty
	}

	jsonData, err := json.Marshal(result)
	require.NoError(t, err)

	var jsonMap map[string]any
	err = json.Unmarshal(jsonData, &jsonMap)
	require.NoError(t, err)

	assert.Equal(t, "test message", jsonMap["message"])
	assert.NotContains(t, jsonMap, "data")
}

func TestInvalidParam(t *testing.T) {
	tests := []struct {
		name            string
		msg             []string
		expectedCode    int
		expectedMessage string
	}{
		{
			name:            "no message",
			msg:             []string{},
			expectedCode:    fiber.StatusBadRequest,
			expectedMessage: "invalid request data",
		},
		{
			name:            "empty message",
			msg:             []string{""},
			expectedCode:    fiber.StatusBadRequest,
			expectedMessage: "invalid request data",
		},
		{
			name:            "custom message",
			msg:             []string{"custom invalid param message"},
			expectedCode:    fiber.StatusBadRequest,
			expectedMessage: "custom invalid param message",
		},
		{
			name:            "multiple messages - uses first",
			msg:             []string{"first message", "second message"},
			expectedCode:    fiber.StatusBadRequest,
			expectedMessage: "first message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fiberErr := InvalidParam(tt.msg...)

			require.NotNil(t, fiberErr)
			assert.Equal(t, tt.expectedCode, fiberErr.Code)
			assert.Equal(t, tt.expectedMessage, fiberErr.Message)
		})
	}
}

func TestUnauthorized(t *testing.T) {
	tests := []struct {
		name            string
		msg             []string
		expectedCode    int
		expectedMessage string
	}{
		{
			name:            "no message",
			msg:             []string{},
			expectedCode:    fiber.StatusUnauthorized,
			expectedMessage: "unauthorized",
		},
		{
			name:            "empty message",
			msg:             []string{""},
			expectedCode:    fiber.StatusUnauthorized,
			expectedMessage: "unauthorized",
		},
		{
			name:            "custom message",
			msg:             []string{"custom unauthorized message"},
			expectedCode:    fiber.StatusUnauthorized,
			expectedMessage: "custom unauthorized message",
		},
		{
			name:            "multiple messages - uses first",
			msg:             []string{"first message", "second message"},
			expectedCode:    fiber.StatusUnauthorized,
			expectedMessage: "first message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fiberErr := Unauthorized(tt.msg...)

			require.NotNil(t, fiberErr)
			assert.Equal(t, tt.expectedCode, fiberErr.Code)
			assert.Equal(t, tt.expectedMessage, fiberErr.Message)
		})
	}
}

func TestForbidden(t *testing.T) {
	tests := []struct {
		name            string
		msg             []string
		expectedCode    int
		expectedMessage string
	}{
		{
			name:            "no message",
			msg:             []string{},
			expectedCode:    fiber.StatusForbidden,
			expectedMessage: "forbidden",
		},
		{
			name:            "empty message",
			msg:             []string{""},
			expectedCode:    fiber.StatusForbidden,
			expectedMessage: "forbidden",
		},
		{
			name:            "custom message",
			msg:             []string{"custom forbidden message"},
			expectedCode:    fiber.StatusForbidden,
			expectedMessage: "custom forbidden message",
		},
		{
			name:            "multiple messages - uses first",
			msg:             []string{"first message", "second message"},
			expectedCode:    fiber.StatusForbidden,
			expectedMessage: "first message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fiberErr := Forbidden(tt.msg...)

			require.NotNil(t, fiberErr)
			assert.Equal(t, tt.expectedCode, fiberErr.Code)
			assert.Equal(t, tt.expectedMessage, fiberErr.Message)
		})
	}
}

func TestSendSucceed(t *testing.T) {
	tests := []struct {
		name         string
		data         []any
		expectedData any
	}{
		{
			name:         "no data",
			data:         []any{},
			expectedData: nil,
		},
		{
			name:         "nil data",
			data:         []any{nil},
			expectedData: nil,
		},
		{
			name:         "string data",
			data:         []any{"test data"},
			expectedData: "test data",
		},
		{
			name:         "map data",
			data:         []any{Map{"key": "value"}},
			expectedData: Map{"key": "value"},
		},
		{
			name:         "list data",
			data:         []any{List{"item1", "item2"}},
			expectedData: List{"item1", "item2"},
		},
		{
			name:         "number data",
			data:         []any{42},
			expectedData: 42,
		},
		{
			name:         "multiple data - uses first",
			data:         []any{"first", "second"},
			expectedData: "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Get("/test", func(c fiber.Ctx) error {
				return SendSucceed(c, tt.data...)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var result WebResult
			err = json.Unmarshal(body, &result)
			require.NoError(t, err)

			assert.Equal(t, "success", result.Message)
			// JSON unmarshaling changes type aliases and numbers, so we need to check differently
			if tt.expectedData != nil {
				assert.NotNil(t, result.Data)
				// For complex types, we'll check by converting to JSON and back
				expectedJSON, _ := json.Marshal(tt.expectedData)
				actualJSON, _ := json.Marshal(result.Data)
				assert.Equal(t, string(expectedJSON), string(actualJSON))
			} else {
				assert.Nil(t, result.Data)
			}
		})
	}
}

func TestSendFailed(t *testing.T) {
	tests := []struct {
		name         string
		msg          string
		data         []any
		expectedMsg  string
		expectedData any
	}{
		{
			name:         "empty message",
			msg:          "",
			data:         []any{},
			expectedMsg:  "operation failed",
			expectedData: nil,
		},
		{
			name:         "custom message",
			msg:          "custom error message",
			data:         []any{},
			expectedMsg:  "custom error message",
			expectedData: nil,
		},
		{
			name:         "with data",
			msg:          "error with data",
			data:         []any{Map{"error_code": "E001"}},
			expectedMsg:  "error with data",
			expectedData: Map{"error_code": "E001"},
		},
		{
			name:         "with nil data",
			msg:          "error with nil data",
			data:         []any{nil},
			expectedMsg:  "error with nil data",
			expectedData: nil,
		},
		{
			name:         "multiple data - uses first",
			msg:          "error message",
			data:         []any{"first", "second"},
			expectedMsg:  "error message",
			expectedData: "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Get("/test", func(c fiber.Ctx) error {
				return SendFailed(c, tt.msg, tt.data...)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var result WebResult
			err = json.Unmarshal(body, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedMsg, result.Message)
			// JSON unmarshaling changes type aliases and numbers, so we need to check differently
			if tt.expectedData != nil {
				assert.NotNil(t, result.Data)
				// For complex types, we'll check by converting to JSON and back
				expectedJSON, _ := json.Marshal(tt.expectedData)
				actualJSON, _ := json.Marshal(result.Data)
				assert.Equal(t, string(expectedJSON), string(actualJSON))
			} else {
				assert.Nil(t, result.Data)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		name            string
		msg             []string
		expectedCode    int
		expectedMessage string
	}{
		{
			name:            "no message",
			msg:             []string{},
			expectedCode:    fiber.StatusNotFound,
			expectedMessage: "resource not found",
		},
		{
			name:            "empty message",
			msg:             []string{""},
			expectedCode:    fiber.StatusNotFound,
			expectedMessage: "resource not found",
		},
		{
			name:            "custom message",
			msg:             []string{"custom not found message"},
			expectedCode:    fiber.StatusNotFound,
			expectedMessage: "custom not found message",
		},
		{
			name:            "multiple messages - uses first",
			msg:             []string{"first message", "second message"},
			expectedCode:    fiber.StatusNotFound,
			expectedMessage: "first message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NotFound(tt.msg...)

			require.Error(t, err)

			fiberErr, ok := err.(*fiber.Error)
			require.True(t, ok)
			assert.Equal(t, tt.expectedCode, fiberErr.Code)
			assert.Equal(t, tt.expectedMessage, fiberErr.Message)
		})
	}
}

func TestSystemBusy(t *testing.T) {
	tests := []struct {
		name         string
		err          []error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "no error",
			err:          []error{},
			expectedCode: fiber.StatusInternalServerError,
			expectedMsg:  "system busy",
		},
		{
			name:         "nil error",
			err:          []error{nil},
			expectedCode: fiber.StatusInternalServerError,
			expectedMsg:  "system busy",
		},
		{
			name:         "with error but no logger initialized",
			err:          []error{errors.New("some internal error")},
			expectedCode: fiber.StatusInternalServerError,
			expectedMsg:  "system busy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SystemBusy(tt.err...)

			require.Error(t, err)

			fiberErr, ok := err.(*fiber.Error)
			require.True(t, ok)
			assert.Equal(t, tt.expectedCode, fiberErr.Code)
			assert.Equal(t, tt.expectedMsg, fiberErr.Message)
		})
	}
}

func TestSendSucceed_Integration(t *testing.T) {
	app := fiber.New()

	app.Get("/users", func(c fiber.Ctx) error {
		users := List{
			Map{"id": 1, "name": "John"},
			Map{"id": 2, "name": "Jane"},
		}
		return SendSucceed(c, users)
	})

	app.Get("/user/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		user := Map{"id": id, "name": "John Doe"}
		return SendSucceed(c, user)
	})

	app.Get("/ping", func(c fiber.Ctx) error {
		return SendSucceed(c)
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedMsg    string
		hasData        bool
	}{
		{
			name:           "list response",
			path:           "/users",
			expectedStatus: 200,
			expectedMsg:    "success",
			hasData:        true,
		},
		{
			name:           "single response",
			path:           "/user/123",
			expectedStatus: 200,
			expectedMsg:    "success",
			hasData:        true,
		},
		{
			name:           "no data response",
			path:           "/ping",
			expectedStatus: 200,
			expectedMsg:    "success",
			hasData:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var result WebResult
			err = json.Unmarshal(body, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedMsg, result.Message)

			if tt.hasData {
				assert.NotNil(t, result.Data)
			} else {
				assert.Nil(t, result.Data)
			}
		})
	}
}

func TestSendFailed_Integration(t *testing.T) {
	app := fiber.New()

	app.Get("/error", func(c fiber.Ctx) error {
		return SendFailed(c, "validation failed", Map{"field": "email", "error": "invalid format"})
	})

	app.Get("/simple-error", func(c fiber.Ctx) error {
		return SendFailed(c, "simple error")
	})

	app.Get("/empty-error", func(c fiber.Ctx) error {
		return SendFailed(c, "")
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedMsg    string
		hasData        bool
	}{
		{
			name:           "error with data",
			path:           "/error",
			expectedStatus: 400,
			expectedMsg:    "validation failed",
			hasData:        true,
		},
		{
			name:           "simple error",
			path:           "/simple-error",
			expectedStatus: 400,
			expectedMsg:    "simple error",
			hasData:        false,
		},
		{
			name:           "empty error message",
			path:           "/empty-error",
			expectedStatus: 400,
			expectedMsg:    "operation failed",
			hasData:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var result WebResult
			err = json.Unmarshal(body, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedMsg, result.Message)

			if tt.hasData {
				assert.NotNil(t, result.Data)
			} else {
				assert.Nil(t, result.Data)
			}
		})
	}
}

func TestComplexDataStructures(t *testing.T) {
	app := fiber.New()

	app.Get("/complex", func(c fiber.Ctx) error {
		data := Map{
			"users": List{
				Map{"id": 1, "name": "John", "active": true},
				Map{"id": 2, "name": "Jane", "active": false},
			},
			"metadata": Map{
				"total": 2,
				"page":  1,
				"limit": 10,
			},
			"nested": Map{
				"level1": Map{
					"level2": Map{
						"level3": "deep value",
					},
				},
			},
		}
		return SendSucceed(c, data)
	})

	req := httptest.NewRequest("GET", "/complex", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result WebResult
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "success", result.Message)
	assert.NotNil(t, result.Data)

	// Verify the complex data structure
	data, ok := result.Data.(map[string]any)
	require.True(t, ok)

	users, ok := data["users"].([]any)
	require.True(t, ok)
	assert.Len(t, users, 2)

	metadata, ok := data["metadata"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(2), metadata["total"]) // JSON unmarshals numbers as float64
}

func TestErrorHandling_FiberIntegration(t *testing.T) {
	app := fiber.New()

	app.Get("/invalid-param", func(c fiber.Ctx) error {
		return InvalidParam("invalid user ID")
	})

	app.Get("/unauthorized", func(c fiber.Ctx) error {
		return Unauthorized("token expired")
	})

	app.Get("/forbidden", func(c fiber.Ctx) error {
		return Forbidden("insufficient permissions")
	})

	app.Get("/not-found", func(c fiber.Ctx) error {
		return NotFound("user not found")
	})

	// Cannot test SystemBusy with errors as it requires goe.Log() initialization
	app.Get("/system-busy", func(c fiber.Ctx) error {
		return SystemBusy()
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "invalid param",
			path:           "/invalid-param",
			expectedStatus: 400,
			expectedMsg:    "invalid user ID",
		},
		{
			name:           "unauthorized",
			path:           "/unauthorized",
			expectedStatus: 401,
			expectedMsg:    "token expired",
		},
		{
			name:           "forbidden",
			path:           "/forbidden",
			expectedStatus: 403,
			expectedMsg:    "insufficient permissions",
		},
		{
			name:           "not found",
			path:           "/not-found",
			expectedStatus: 404,
			expectedMsg:    "user not found",
		},
		{
			name:           "system busy",
			path:           "/system-busy",
			expectedStatus: 500,
			expectedMsg:    "system busy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// Fiber error responses might be plain text or JSON depending on configuration
			bodyStr := string(body)
			assert.Contains(t, bodyStr, tt.expectedMsg)
		})
	}
}

func BenchmarkSendSucceed(b *testing.B) {
	app := fiber.New()

	app.Get("/benchmark", func(c fiber.Ctx) error {
		return SendSucceed(c, Map{"test": "data"})
	})

	b.ResetTimer()
	for b.Loop() {
		req := httptest.NewRequest("GET", "/benchmark", nil)
		_, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSendFailed(b *testing.B) {
	app := fiber.New()

	app.Get("/benchmark", func(c fiber.Ctx) error {
		return SendFailed(c, "test error", Map{"error": "details"})
	})

	b.ResetTimer()
	for b.Loop() {
		req := httptest.NewRequest("GET", "/benchmark", nil)
		_, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWebResultCreation(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		result := &WebResult{
			Message: "test message",
			Data:    Map{"key": "value"},
		}
		_ = result
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	result := &WebResult{
		Message: "test message",
		Data:    Map{"key": "value", "number": 42},
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := json.Marshal(result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

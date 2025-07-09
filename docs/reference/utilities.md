# Utilities

GOE Framework provides a collection of utility functions and types to help with common tasks. These utilities are designed to be efficient, safe, and easy to use.

## String Utilities

### Safe String Conversion
```go
// Convert various types to string safely
str := utils.ToString(42)        // "42"
str = utils.ToString(true)       // "true"
str = utils.ToString(3.14)       // "3.14"
```

### String Validation
```go
// Check if string is empty or whitespace
isEmpty := utils.IsEmpty("   ")  // true
isEmpty = utils.IsEmpty("hello") // false
```

## Number Utilities

### Safe Number Parsing
```go
// Parse string to int with default value
num := utils.ToInt("42", 0)      // 42
num = utils.ToInt("invalid", 0)  // 0 (default)

// Parse string to float with default value
f := utils.ToFloat("3.14", 0.0)  // 3.14
f = utils.ToFloat("invalid", 0.0) // 0.0 (default)
```

### Number Validation
```go
// Check if value is within range
inRange := utils.InRange(5, 1, 10)  // true
inRange = utils.InRange(15, 1, 10)  // false
```

## Collection Utilities

### Slice Operations
```go
// Check if slice contains value
contains := utils.Contains([]string{"a", "b", "c"}, "b")  // true

// Remove duplicates from slice
unique := utils.Unique([]string{"a", "b", "a", "c"})  // ["a", "b", "c"]

// Filter slice
filtered := utils.Filter([]int{1, 2, 3, 4, 5}, func(x int) bool {
    return x%2 == 0
})  // [2, 4]
```

### Map Operations
```go
// Get map keys
keys := utils.Keys(map[string]int{"a": 1, "b": 2})  // ["a", "b"]

// Get map values
values := utils.Values(map[string]int{"a": 1, "b": 2})  // [1, 2]

// Merge maps
merged := utils.MergeMaps(
    map[string]int{"a": 1, "b": 2},
    map[string]int{"c": 3, "d": 4},
)  // map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
```

## File Utilities

### File Operations
```go
// Check if file exists
exists := utils.FileExists("config.yaml")  // true/false

// Read file safely
content, err := utils.ReadFile("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Write file safely
err = utils.WriteFile("output.txt", []byte("Hello, World!"))
if err != nil {
    log.Fatal(err)
}
```

### Path Operations
```go
// Get file extension
ext := utils.GetExtension("file.txt")  // ".txt"

// Get filename without extension
name := utils.GetFilename("path/to/file.txt")  // "file"

// Join paths safely
path := utils.JoinPath("path", "to", "file.txt")  // "path/to/file.txt"
```

## JSON Utilities

### Safe JSON Operations
```go
// Marshal to JSON with error handling
jsonStr, err := utils.ToJSON(map[string]interface{}{
    "name": "John",
    "age":  30,
})

// Unmarshal from JSON safely
var data map[string]interface{}
err = utils.FromJSON(jsonStr, &data)
if err != nil {
    log.Fatal(err)
}
```

### JSON Validation
```go
// Check if string is valid JSON
isValid := utils.IsValidJSON(`{"name": "John"}`)  // true
isValid = utils.IsValidJSON(`{invalid}`)          // false
```

## Time Utilities

### Time Formatting
```go
// Format time in various formats
formatted := utils.FormatTime(time.Now(), "2006-01-02")  // "2023-12-01"
formatted = utils.FormatTime(time.Now(), "15:04:05")     // "14:30:45"
```

### Time Parsing
```go
// Parse time from string
t, err := utils.ParseTime("2023-12-01", "2006-01-02")
if err != nil {
    log.Fatal(err)
}
```

### Duration Utilities
```go
// Human readable duration
duration := utils.HumanDuration(65 * time.Second)  // "1m5s"
duration = utils.HumanDuration(3661 * time.Second)  // "1h1m1s"
```

## Validation Utilities

### Email Validation
```go
isValid := utils.IsValidEmail("user@example.com")  // true
isValid = utils.IsValidEmail("invalid-email")      // false
```

### URL Validation
```go
isValid := utils.IsValidURL("https://example.com")  // true
isValid = utils.IsValidURL("invalid-url")           // false
```

### UUID Validation
```go
isValid := utils.IsValidUUID("123e4567-e89b-12d3-a456-426614174000")  // true
isValid = utils.IsValidUUID("invalid-uuid")                           // false
```

## Cryptography Utilities

### Hashing
```go
// MD5 hash
hash := utils.MD5("hello world")  // "5d41402abc4b2a76b9719d911017c592"

// SHA256 hash
hash = utils.SHA256("hello world")  // "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
```

### Random Generation
```go
// Generate random string
randomStr := utils.RandomString(10)  // "aB3xK9mP2q"

// Generate random number
randomNum := utils.RandomInt(1, 100)  // Random number between 1 and 100
```

## HTTP Utilities

### Request Utilities
```go
// Get client IP address
ip := utils.GetClientIP(c)  // "192.168.1.1"

// Get user agent
userAgent := utils.GetUserAgent(c)  // "Mozilla/5.0 ..."

// Check if request is AJAX
isAjax := utils.IsAjaxRequest(c)  // true/false
```

### Response Utilities
```go
// Send JSON response
err := utils.SendJSON(c, fiber.StatusOK, map[string]interface{}{
    "message": "Success",
    "data":    data,
})

// Send error response
err = utils.SendError(c, fiber.StatusBadRequest, "Invalid request")
```

## Environment Utilities

### Environment Variables
```go
// Get environment variable with default
value := utils.GetEnv("DATABASE_URL", "postgres://localhost/db")

// Get environment variable as int
port := utils.GetEnvInt("PORT", 8080)

// Get environment variable as bool
debug := utils.GetEnvBool("DEBUG", false)
```

### Configuration Utilities
```go
// Load configuration from file
config, err := utils.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Get nested configuration value
value := utils.GetConfigValue(config, "database.host", "localhost")
```

## Error Utilities

### Error Handling
```go
// Wrap error with context
err = utils.WrapError(err, "failed to process request")

// Check if error is specific type
if utils.IsNotFoundError(err) {
    // Handle not found error
}

// Create custom error
err = utils.NewError("USER_NOT_FOUND", "User not found")
```

## Usage Examples

### Complete Example
```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/oeasenet/goe/utils"
)

func main() {
    app := fiber.New()
    
    app.Get("/user/:id", func(c *fiber.Ctx) error {
        // Get and validate user ID
        id := c.Params("id")
        if utils.IsEmpty(id) {
            return utils.SendError(c, fiber.StatusBadRequest, "User ID required")
        }
        
        // Convert to int
        userID := utils.ToInt(id, 0)
        if userID == 0 {
            return utils.SendError(c, fiber.StatusBadRequest, "Invalid user ID")
        }
        
        // Check if user exists (mock)
        user := getUserByID(userID)
        if user == nil {
            return utils.SendError(c, fiber.StatusNotFound, "User not found")
        }
        
        // Return user data
        return utils.SendJSON(c, fiber.StatusOK, user)
    })
    
    app.Listen(":8080")
}
```

These utilities provide a solid foundation for building robust applications with GOE Framework. They handle common edge cases and provide consistent, safe APIs for everyday tasks.
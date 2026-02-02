package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUIDConstraint(t *testing.T) {
	c := &UUIDConstraint{}

	assert.Equal(t, "uuid", c.Name())

	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uuid lowercase", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uuid uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"invalid uuid - too short", "550e8400-e29b-41d4", false},
		{"invalid uuid - random string", "not-a-uuid", false},
		{"invalid uuid - empty", "", false},
		{"valid uuid without dashes", "12345678901234567890123456789012", true}, // google/uuid accepts this format
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Execute(tt.param)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUIntConstraint(t *testing.T) {
	c := &UIntConstraint{}

	assert.Equal(t, "uint", c.Name())

	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{"valid positive integer", "123", true},
		{"valid large integer", "999999999", true},
		{"valid single digit", "1", true},
		{"invalid zero", "0", false},
		{"invalid negative", "-1", false},
		{"invalid float", "1.5", false},
		{"invalid string", "abc", false},
		{"invalid empty", "", false},
		{"invalid mixed", "12abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Execute(tt.param)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSlugConstraint(t *testing.T) {
	c := &SlugConstraint{}

	assert.Equal(t, "slug", c.Name())

	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{"valid simple slug", "hello-world", true},
		{"valid single word", "hello", true},
		{"valid with numbers", "post-123", true},
		{"valid numbers only", "123", true},
		{"valid complex slug", "my-awesome-blog-post-2024", true},
		{"invalid uppercase", "Hello-World", false},
		{"invalid starts with hyphen", "-hello", false},
		{"invalid ends with hyphen", "hello-", false},
		{"invalid consecutive hyphens", "hello--world", false},
		{"invalid spaces", "hello world", false},
		{"invalid special chars", "hello_world", false},
		{"invalid empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Execute(tt.param)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEmailConstraint(t *testing.T) {
	c := &EmailConstraint{}

	assert.Equal(t, "email", c.Name())

	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{"valid simple email", "test@example.com", true},
		{"valid with subdomain", "test@mail.example.com", true},
		{"valid with plus", "test+tag@example.com", true},
		{"valid with dots", "first.last@example.com", true},
		{"valid with numbers", "test123@example.com", true},
		{"invalid no at", "testexample.com", false},
		{"invalid no domain", "test@", false},
		{"invalid no local", "@example.com", false},
		{"invalid no tld", "test@example", false},
		{"invalid empty", "", false},
		{"invalid spaces", "test @example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Execute(tt.param)
			assert.Equal(t, tt.expected, result)
		})
	}
}

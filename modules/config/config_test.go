package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGet(t *testing.T) {
	config := New("./test")
	env := config.Get("Key")
	assert.Equal(t, env, "key")
}

func TestGetInt(t *testing.T) {
	config := New("./test")
	data := config.GetInt("IntKey")
	assert.Equal(t, data, 2123)
}

func TestGetBool(t *testing.T) {
	config := New("./test")
	data := config.GetBool("BoolKey")
	assert.Equal(t, data, true)
}

func TestGetStringSlice(t *testing.T) {
	config := New("./test")
	data := config.GetStringSlice("StringSliceKey")
	assert.Equal(t, data, []string{"i", "am", "groot"})
}

func TestGetIntSlice(t *testing.T) {
	config := New("./test")
	data := config.GetIntSlice("IntSliceKey")
	assert.Equal(t, data, []int{1, 5, 8, 2})
}

func TestGetBoolSlice(t *testing.T) {
	config := New("./test")
	data := config.GetBoolSlice("BoolSliceKey")
	assert.Equal(t, data, []bool{false, true, true, false})
}

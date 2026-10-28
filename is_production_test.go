package goe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2"
)

func TestIsProduction(t *testing.T) {
	for env, want := range map[string]bool{
		"prod":           true,
		"production":     true,
		"Production":     true,
		" PRODUCTION ":   true,
		"dev":            false,
		"staging":        false,
		"pre-production": false,
		"":               false,
	} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("GOE_ENV", env)
			goe.New(goe.Options{})
			assert.Equal(t, want, goe.IsProduction())
		})
	}
}

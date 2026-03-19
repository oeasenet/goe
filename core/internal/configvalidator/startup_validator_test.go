package configvalidator

import (
	"context"
	"testing"

	"go.oease.dev/goe/v2/contract"
)

// Mock module for testing
type mockModule struct {
	name           string
	validationFunc func() error
}

func (m *mockModule) Name() string {
	return m.name
}

func (m *mockModule) OnStart(ctx context.Context) error {
	return nil
}

func (m *mockModule) OnStop(ctx context.Context) error {
	return nil
}

func (m *mockModule) ValidateConfig() error {
	if m.validationFunc != nil {
		return m.validationFunc()
	}
	return nil
}

func TestStartupValidator(t *testing.T) {

	tests := []struct {
		name        string
		setupFunc   func(*StartupValidator)
		expectError bool
		errorString string
	}{
		{
			name: "no modules registered",
			setupFunc: func(v *StartupValidator) {
				// No modules to register
			},
			expectError: false,
		},
		{
			name: "single valid module",
			setupFunc: func(v *StartupValidator) {
				module := &mockModule{
					name: "test",
					validationFunc: func() error {
						return nil
					},
				}
				v.RegisterModule(module)
			},
			expectError: false,
		},
		{
			name: "single invalid module",
			setupFunc: func(v *StartupValidator) {
				module := &mockModule{
					name: "test",
					validationFunc: func() error {
						return &contract.ConfigValidationError{
							Module:      "test",
							MissingKeys: []string{"REQUIRED_KEY"},
						}
					},
				}
				v.RegisterModule(module)
			},
			expectError: true,
			errorString: "REQUIRED_KEY",
		},
		{
			name: "multiple modules - all valid",
			setupFunc: func(v *StartupValidator) {
				modules := []*mockModule{
					{
						name: "test1",
						validationFunc: func() error {
							return nil
						},
					},
					{
						name: "test2",
						validationFunc: func() error {
							return nil
						},
					},
				}
				for _, module := range modules {
					v.RegisterModule(module)
				}
			},
			expectError: false,
		},
		{
			name: "multiple modules - one invalid",
			setupFunc: func(v *StartupValidator) {
				modules := []*mockModule{
					{
						name: "test1",
						validationFunc: func() error {
							return nil
						},
					},
					{
						name: "test2",
						validationFunc: func() error {
							return &contract.ConfigValidationError{
								Module:      "test2",
								MissingKeys: []string{"REQUIRED_KEY"},
							}
						},
					},
				}
				for _, module := range modules {
					v.RegisterModule(module)
				}
			},
			expectError: true,
			errorString: "test2",
		},
		{
			name: "validation disabled",
			setupFunc: func(v *StartupValidator) {
				module := &mockModule{
					name: "test",
					validationFunc: func() error {
						return &contract.ConfigValidationError{
							Module:      "test",
							MissingKeys: []string{"REQUIRED_KEY"},
						}
					},
				}
				v.RegisterModule(module)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset config for each test
			testCfg := &mockConfig{values: make(map[string]any)}
			testLogger := &mockLogger{}

			// Set disable flag for the specific test
			if tt.name == "validation disabled" {
				testCfg.Set("DISABLE_CONFIG_VALIDATION", "true")
			}

			v := NewStartupValidator(testCfg, testLogger)
			tt.setupFunc(v)

			err := v.ValidateAll()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				if tt.errorString != "" {
					if err.Error() == "" {
						t.Errorf("expected error message to contain '%s' but got empty message", tt.errorString)
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestStartupValidatorSpecificModule(t *testing.T) {
	cfg := &mockConfig{values: make(map[string]any)}
	logger := &mockLogger{}

	v := NewStartupValidator(cfg, logger)

	// Test validating non-existent module
	err := v.ValidateModule("non-existent")
	if err == nil {
		t.Error("expected error for non-existent module")
	}

	// Register a module and test validating it
	module := &mockModule{
		name: "test",
		validationFunc: func() error {
			return nil
		},
	}
	v.RegisterModule(module)

	err = v.ValidateModule("test")
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

	// Test validating module with error
	invalidModule := &mockModule{
		name: "invalid",
		validationFunc: func() error {
			return &contract.ConfigValidationError{
				Module:      "invalid",
				MissingKeys: []string{"REQUIRED_KEY"},
			}
		},
	}
	v.RegisterModule(invalidModule)

	err = v.ValidateModule("invalid")
	if err == nil {
		t.Error("expected error for invalid module")
	}
}

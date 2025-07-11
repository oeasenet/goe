package validator

import (
	"fmt"
	"strings"

	"go.oease.dev/goe/v2/contract"
)

// StartupValidator validates configuration at application startup
type StartupValidator struct {
	config  contract.Config
	logger  contract.Logger
	modules map[string]contract.ModuleWithValidator
}

// NewStartupValidator creates a new startup validator
func NewStartupValidator(config contract.Config, logger contract.Logger) *StartupValidator {
	return &StartupValidator{
		config:  config,
		logger:  logger,
		modules: make(map[string]contract.ModuleWithValidator),
	}
}

// RegisterModule registers a module for validation
func (v *StartupValidator) RegisterModule(module contract.ModuleWithValidator) {
	v.modules[module.Name()] = module
}

// ValidateAll validates all registered modules
func (v *StartupValidator) ValidateAll() error {
	var errors []string

	// Check if validation is disabled
	if v.config.GetBool("DISABLE_CONFIG_VALIDATION") {
		v.logger.Warn("Configuration validation is disabled")
		return nil
	}

	v.logger.Info("Starting configuration validation", "module_count", len(v.modules))

	for name, module := range v.modules {
		// Check if module is enabled
		enabledKey := fmt.Sprintf("%s_ENABLED", strings.ToUpper(name))

		// Some modules are always enabled (like config, log)
		alwaysEnabled := []string{"config", "log"}
		isAlwaysEnabled := false
		for _, m := range alwaysEnabled {
			if name == m {
				isAlwaysEnabled = true
				break
			}
		}

		// Skip validation if module is explicitly disabled
		if !isAlwaysEnabled && v.config.Has(enabledKey) && !v.config.GetBool(enabledKey) {
			v.logger.Debug("Skipping validation for disabled module", "module", name)
			continue
		}

		v.logger.Debug("Validating module configuration", "module", name)

		if err := module.ValidateConfig(); err != nil {
			errors = append(errors, fmt.Sprintf("Module '%s': %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}

	v.logger.Info("Configuration validation completed successfully")
	return nil
}

// ValidateModule validates a specific module
func (v *StartupValidator) ValidateModule(moduleName string) error {
	module, exists := v.modules[moduleName]
	if !exists {
		return fmt.Errorf("module '%s' not registered for validation", moduleName)
	}

	return module.ValidateConfig()
}

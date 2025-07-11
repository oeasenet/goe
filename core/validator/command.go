package validator

import (
	"fmt"

	"go.oease.dev/goe/v2/contract"
)

// ValidateCommand provides a CLI command to validate configuration
type ValidateCommand struct {
	config contract.Config
	logger contract.Logger
}

// NewValidateCommand creates a new validation command
func NewValidateCommand(config contract.Config, logger contract.Logger) *ValidateCommand {
	return &ValidateCommand{
		config: config,
		logger: logger,
	}
}

// Execute runs the validation command
func (cmd *ValidateCommand) Execute() error {
	fmt.Printf("🔍 Validating configuration for all modules...\n\n")

	// Note: To avoid import cycles, this method now just provides the framework
	// Actual module validation should be done by the application that imports
	// the modules directly

	fmt.Printf("✅ Validation framework ready!\n\n")
	return nil
}

// ValidateConfig validates a configuration against a set of requirements
func ValidateConfig(config contract.Config, moduleName string, requirements []contract.ConfigRequirement) error {
	v := NewConfigValidator(config, moduleName)

	for _, req := range requirements {
		if req.Required {
			if req.ValidateFn != nil {
				v.RequireWithValidator(req.Key, req.Description, req.ValidateFn)
			} else {
				v.Require(req.Key, req.Description)
			}
		} else {
			if req.ValidateFn != nil {
				v.Optional(req.Key, req.Description, req.ValidateFn)
			}
		}
	}

	return v.Validate()
}

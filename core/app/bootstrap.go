package app

import (
	"context"
	"fmt"
	"os"

	"go.oease.dev/goe/v2/contract"
)

// Bootstrap creates and configures a new application with core modules
// To avoid import cycles, this now returns a basic app without core modules
// Applications should add their own modules as needed
func Bootstrap(name, version, environment string) (contract.Application, error) {
	return New(name, version, environment), nil
}

// BootstrapAndValidate creates, configures, and validates the application
func BootstrapAndValidate(name, version, environment string) (contract.Application, error) {
	app, err := Bootstrap(name, version, environment)
	if err != nil {
		return nil, err
	}

	// Start the app to trigger validation
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		// If validation fails, provide helpful error message
		if os.Getenv("DISABLE_CONFIG_VALIDATION") != "true" {
			fmt.Fprintf(os.Stderr, "\n❌ Configuration validation failed:\n%v\n\n", err)
			fmt.Fprintf(os.Stderr, "To skip validation during development, set DISABLE_CONFIG_VALIDATION=true\n\n")
		}
		return nil, err
	}

	// Stop the app after validation
	app.Stop(ctx)

	return app, nil
}

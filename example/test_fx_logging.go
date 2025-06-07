package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.oease.dev/goe/v2"
)

func main() {
	// Create environment file for demo
	envContent := `# Application Configuration
APP_NAME=FxLoggingTest
APP_VERSION=1.0.0

# Test with INFO level (Fx logs should not appear)
LOG_LEVEL=info
LOG_FORMAT=text
`
	_ = os.WriteFile(".env", []byte(envContent), 0644)
	defer os.Remove(".env")

	fmt.Println("=== Testing Fx logging at INFO level ===")
	fmt.Println("You should NOT see Fx-related logs (provided, invoking, etc.)")
	fmt.Println("")

	app := goe.New(goe.Options{})
	
	// Run in background
	go app.Container().Run()
	
	// Wait a bit then stop
	time.Sleep(100 * time.Millisecond)
	if err := app.Container().Stop(context.Background()); err != nil {
		fmt.Printf("Error stopping app: %v\n", err)
	}
	time.Sleep(50 * time.Millisecond) // Let it clean up

	fmt.Println("\n=== Now testing with DEBUG level ===")
	
	// Update env file with debug level
	envContentDebug := `# Application Configuration
APP_NAME=FxLoggingTest
APP_VERSION=1.0.0

# Test with DEBUG level (Fx logs should appear)
LOG_LEVEL=debug
LOG_FORMAT=text
`
	_ = os.WriteFile(".env", []byte(envContentDebug), 0644)

	// Create new app instance with debug logging
	appDebug := goe.New(goe.Options{})
	
	fmt.Println("You should now see Fx-related logs at DEBUG level")
	fmt.Println("")
	
	// Run in background
	go appDebug.Container().Run()
	
	// Wait a bit then stop
	time.Sleep(100 * time.Millisecond)
	if err := appDebug.Container().Stop(context.Background()); err != nil {
		fmt.Printf("Error stopping app: %v\n", err)
	}
	time.Sleep(50 * time.Millisecond) // Let it clean up

	fmt.Println("\n✅ Fx logging level test completed!")
}
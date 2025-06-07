package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.oease.dev/goe/v2"
	"go.oease.dev/goe/v2/contract"
	"gorm.io/gorm"
)

// Define a GORM model for the demo
type DemoUser struct {
	gorm.Model
	Name  string `gorm:"unique"`
	Email string
}

func main() {
	// Clean up previous demo database file if it exists
	_ = os.Remove("./db_demo.db")

	// Initialize Goe application with DB module enabled
	// The .env file in this directory will be loaded by default
	app := goe.New(goe.Options{
		WithDB: true,
	})

	// Start the application (this will trigger OnStart for modules, including DB connection)
	if err := app.Container().Start(context.Background()); err != nil {
		goe.Log().Fatal("Failed to start application", contract.NewField("error", err))
	}
	defer app.Container().Stop(context.Background()) // Ensure resources are cleaned up

	goe.Log().Info("Application started successfully with DB module.")

	// Access the database instance
	db := goe.DB().Instance()
	if db == nil {
		goe.Log().Fatal("Database instance is nil. Check configuration and logs.")
		return
	}

	// Auto-migrate the schema for DemoUser model
	goe.Log().Info("Running auto-migration for DemoUser model...")
	err := db.AutoMigrate(&DemoUser{})
	if err != nil {
		goe.Log().Error("Failed to auto-migrate DemoUser table", contract.NewField("error", err))
		return
	}
	goe.Log().Info("Auto-migration successful.")

	// Perform some CRUD operations
	// Create
	goe.Log().Info("Creating a new user...")
	newUser := DemoUser{Name: "John Doe", Email: "john.doe@example.com"}
	result := db.Create(&newUser)
	if result.Error != nil {
		goe.Log().Error("Failed to create user", contract.NewField("error", result.Error))
		return
	}
	goe.Log().Info("User created successfully", contract.NewField("ID", newUser.ID))

	// Read
	goe.Log().Info("Fetching user by ID...", contract.NewField("ID", newUser.ID))
	var fetchedUser DemoUser
	result = db.First(&fetchedUser, newUser.ID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			goe.Log().Error("User not found", contract.NewField("ID", newUser.ID))
		} else {
			goe.Log().Error("Failed to fetch user", contract.NewField("error", result.Error))
		}
		return
	}
	goe.Log().Info("User fetched successfully", contract.NewField("Name", fetchedUser.Name), contract.NewField("Email", fetchedUser.Email))

	// Update
	goe.Log().Info("Updating user's email...", contract.NewField("ID", fetchedUser.ID))
	newEmail := "john.doe.updated@example.com"
	result = db.Model(&fetchedUser).Update("Email", newEmail)
	if result.Error != nil {
		goe.Log().Error("Failed to update user's email", contract.NewField("error", result.Error))
		return
	}
	goe.Log().Info("User email updated successfully.", contract.NewField("RowsAffected", result.RowsAffected))

	// Verify update
	var updatedUser DemoUser
	db.First(&updatedUser, fetchedUser.ID)
	goe.Log().Info("Fetched updated user", contract.NewField("Name", updatedUser.Name), contract.NewField("Email", updatedUser.Email))

	// Delete
	goe.Log().Info("Deleting user...", contract.NewField("ID", updatedUser.ID))
	result = db.Delete(&DemoUser{}, updatedUser.ID)
	if result.Error != nil {
		goe.Log().Error("Failed to delete user", contract.NewField("error", result.Error))
		return
	}
	goe.Log().Info("User deleted successfully.", contract.NewField("RowsAffected", result.RowsAffected))

	// Verify deletion
	var deletedUser DemoUser
	result = db.First(&deletedUser, updatedUser.ID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		goe.Log().Info("User successfully verified as deleted (not found).")
	} else if result.Error != nil {
		goe.Log().Error("Error checking for deleted user", contract.NewField("error", result.Error))
	} else {
		goe.Log().Warn("User was found after attempting deletion.", contract.NewField("ID", deletedUser.ID))
	}

	fmt.Println("\nDemo complete. Check logs above and the 'db_demo.db' SQLite file (if not :memory:).")
	fmt.Println("You can delete 'db_demo.db' manually after inspection.")
}

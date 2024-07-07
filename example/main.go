package main

import (
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe"
)

func main() {
	// This is an example of a main function
	// Create a new goe app, this will initialize the app and its dependencies
	err := goe.NewApp()
	if err != nil {
		panic(err)
		return
	}

	// Use the Fiber module to set up a simple hello world route
	goe.UseFiber().App().Get("/hello", func(ctx fiber.Ctx) error {
		return ctx.SendString("Hello, World!")
	})

	// Run the app, this will start the server and block the main thread. Graceful shutdown is supported.
	err = goe.Run()
	if err != nil {
		panic(err)
		return
	}
}

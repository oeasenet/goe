package main

import (
	_ "embed"
	"github.com/gofiber/fiber/v3"
	"go.oease.dev/goe"
)

//go:embed configs/msearch.json
var msearchConfig []byte

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

	// Use the Search module to apply the index configurations, if search is enabled
	err = goe.UseSearch().ApplyIndexConfigs(msearchConfig)
	if err != nil {
		panic(err)
		return
	}

	// Run the app, this will start the server and block the main thread. Graceful shutdown is supported.
	err = goe.Run()
	if err != nil {
		panic(err)
		return
	}
}

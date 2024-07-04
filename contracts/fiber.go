package contracts

import "github.com/gofiber/fiber/v3"

type GoeFiber interface {
	App() *fiber.App
	CreateFiberApp(appName ...string) *fiber.App
}

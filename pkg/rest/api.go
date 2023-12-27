package rest

import (
	"sybline/pkg/core"

	"github.com/gofiber/fiber/v2"
)

func NewRestServer(broker core.Broker) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Get("/info/routing", func(c *fiber.Ctx) error {
		return c.JSON(broker.GetKeyMappings())
	})

	return app
}

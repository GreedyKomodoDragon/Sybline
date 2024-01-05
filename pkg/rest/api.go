package rest

import (
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/rbac"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewRestServer(broker core.Broker, auth auth.AuthManager, rbac rbac.RoleManager, queueManager core.QueueManager) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(cors.New())

	app.Get("/info/routing", func(c *fiber.Ctx) error {
		return c.JSON(*broker.GetKeys())
	})

	app.Get("/info/routing/:routingkey", func(c *fiber.Ctx) error {
		routingkey := c.Params("routingkey")
		if len(routingkey) == 0 {
			c.SendString("invalid routing key length")
			return c.SendStatus(400)
		}

		queues, err := broker.GetQueues(routingkey)
		if err != nil {
			c.SendString("invalid routing key name, does not exist")
			return c.SendStatus(400)
		}

		return c.JSON(queues)
	})

	app.Get("/info/accounts", func(c *fiber.Ctx) error {
		return c.JSON(auth.GetAccounts())
	})

	app.Get("/info/queues", func(c *fiber.Ctx) error {
		return c.JSON(queueManager.GetAllQueues())
	})

	app.Get("/info/accounts/roles/:username", func(c *fiber.Ctx) error {
		username := c.Params("username")
		if len(username) == 0 {
			c.SendString("invalid username length")
			return c.SendStatus(400)
		}

		result, err := rbac.GetRoles(username)
		if err != nil {
			c.SendString("invalid username, may have no roles assigned")
			return c.SendStatus(400)
		}

		return c.JSON(result)
	})

	return app
}

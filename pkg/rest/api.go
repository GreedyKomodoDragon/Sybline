package rest

import (
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/handler"
	"sybline/pkg/rbac"

	"github.com/GreedyKomodoDragon/raft"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewRestServer(broker core.Broker, auth auth.AuthManager, rbac rbac.RoleManager, queueManager core.QueueManager, raftServer raft.Raft, hand handler.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(cors.New())

	// Check if leader
	app.Use(func(c *fiber.Ctx) error {
		return IsLeader(c, raftServer)
	})

	// Middleware for authentication
	app.Use(func(c *fiber.Ctx) error {
		return Authentication(c, auth)
	})

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

	// app.Post("/accounts/roles/:username", func(c *fiber.Ctx) error {
	// 	username := c.Params("username")
	// 	if len(username) == 0 {
	// 		c.SendString("invalid username length")
	// 		return c.SendStatus(400)
	// 	}

	// 	ctx := context.Background()
	// 	ctx = context.WithValue(ctx, "consumerID", c.Locals("consumerID"))
	// 	ctx = context.WithValue(ctx, "username", c.Locals("username"))

	// 	hand.AddRoutingKey(ctx, "", "")

	// 	return c.JSON(result)
	// })

	return app
}

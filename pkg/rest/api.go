package rest

import (
	"context"
	"fmt"
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

	// Assiging Role
	app.Put("/accounts/roles/:username/:role", func(c *fiber.Ctx) error {
		username := c.Params("username")
		if len(username) == 0 {
			c.SendString("invalid username length")
			return c.SendStatus(400)
		}

		role := c.Params("role")
		if len(username) == 0 {
			c.SendString("invalid username length")
			return c.SendStatus(400)
		}

		ctx, err := createContextFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "unable to find authentication information",
			})
		}

		if err = hand.AssignRole(ctx, role, username); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	return app
}

func createContextFromFiberContext(fc *fiber.Ctx) (context.Context, error) {
	// Use the background context as the parent
	ctx := context.Background()

	// Get values from fiber.Ctx locals
	consumerID := fc.Locals("consumerID")
	username := fc.Locals("username")
	token := fc.Locals("token")

	// Check if any of the values are nil
	if consumerID == nil || username == nil || token == nil {
		return ctx, fmt.Errorf("consumerID, username, or token not found in context")
	}

	// Create a context.Context and add values to it
	ctx = context.WithValue(ctx, "consumerID", consumerID)
	ctx = context.WithValue(ctx, "username", username)
	ctx = context.WithValue(ctx, "token", token)

	return ctx, nil
}

package rest

import (
	"context"
	"fmt"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/handler"
	"sybline/pkg/rbac"

	"github.com/GreedyKomodoDragon/raft"
	"github.com/ansrivas/fiberprometheus/v2"
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

	prometheus := fiberprometheus.New("sybline")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	createV1(app, broker, auth, queueManager, rbac, hand, raftServer)

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

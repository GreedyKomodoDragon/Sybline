package rest

import (
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/rbac"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func createV1(app *fiber.App, broker core.Broker, authManager auth.AuthManager, queueManager core.QueueManager, rbac rbac.RoleManager) {
	router := app.Group("/api/v1")

	createInfo(router, broker, authManager, queueManager, rbac)
	createLogin(router, authManager)
}

func createInfo(router fiber.Router, broker core.Broker, auth auth.AuthManager, queueManager core.QueueManager, rbac rbac.RoleManager) {
	infoRouter := router.Group("/info")

	infoRouter.Get("/routing", func(c *fiber.Ctx) error {
		return c.JSON(*broker.GetKeys())
	})

	infoRouter.Get("/routing/:routingkey", func(c *fiber.Ctx) error {
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

	infoRouter.Get("/accounts", func(c *fiber.Ctx) error {
		return c.JSON(auth.GetAccounts())
	})

	infoRouter.Get("/queues", func(c *fiber.Ctx) error {
		return c.JSON(queueManager.GetAllQueues())
	})

	infoRouter.Get("/accounts/roles/:username", func(c *fiber.Ctx) error {
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
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func createLogin(router fiber.Router, authManager auth.AuthManager) {
	router.Post("/login", func(c *fiber.Ctx) error {

		var credentials Credentials
		if err := c.BodyParser(&credentials); err != nil {
			return c.Status(400).SendString("Bad Request")
		}

		token, err := authManager.Login(credentials.Username, auth.GenerateHash(credentials.Password, authManager.Salt()))
		if err != nil {
			log.Debug().Err(err).Msg("user failed to login")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid credentials provided",
			})
		}
		// Do something with the username and password (e.g., authentication)

		// Respond with a message
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"token": token,
		})
	})
}

package rest

import (
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/handler"
	"sybline/pkg/rbac"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func createV1(app *fiber.App, broker core.Broker, authManager auth.AuthManager, queueManager core.QueueManager, rbac rbac.RoleManager, hand handler.Handler) {
	router := app.Group("/api/v1")

	createInfo(router, broker, authManager, queueManager, rbac)
	createAccounts(router, hand)
	createLogin(router, authManager)
}

func createAccounts(router fiber.Router, hand handler.Handler) {
	accountsRouter := router.Group("/accounts")

	// Adding a new account
	accountsRouter.Put("/roles/:username/:role", func(c *fiber.Ctx) error {
		username := c.Params("username")
		if len(username) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid username length",
			})
		}

		role := c.Params("role")
		if len(role) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "invalid role length",
			})
		}

		ctx, err := createContextFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "unable to find context information",
			})
		}

		if err = hand.AssignRole(ctx, role, username); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusCreated)
	})
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

	infoRouter.Get("/queues", func(c *fiber.Ctx) error {
		return c.JSON(queueManager.GetAllQueues())
	})

	infoRouter.Get("/accounts", func(c *fiber.Ctx) error {
		return c.JSON(auth.GetAccounts())
	})

	infoRouter.Get("/roles", func(c *fiber.Ctx) error {
		return c.JSON(rbac.GetAllRoles())
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
	router.Get("/login", func(c *fiber.Ctx) error {

		username := c.Locals("username")
		if username == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "user is not logged in one",
			})
		}

		usernameStr, ok := username.(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "user is not logged in two",
			})
		}

		token := c.Locals("token")
		if token == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "user is not logged in three",
			})
		}

		tokenStr, ok := token.(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "user is not logged in four",
			})
		}

		if _, err := authManager.GetConsumerID(usernameStr, tokenStr); err != nil {
			log.Debug().Err(err).Msg("user failed to login")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "user is not valid or logged in",
			})
		}

		// Respond with a message
		return c.SendStatus(fiber.StatusOK)
	})

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

package rest

import (
	"strconv"
	"sybline/pkg/auth"
	"sybline/pkg/core"
	"sybline/pkg/handler"
	"sybline/pkg/rbac"

	"github.com/GreedyKomodoDragon/raft"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"
)

func createV1(app *fiber.App, broker core.Broker, authManager auth.AuthManager, queueManager core.QueueManager, rbac rbac.RoleManager, hand handler.Handler, raftServer raft.Raft) {
	router := app.Group("/api/v1")

	createInfo(router, broker, authManager, queueManager, rbac, raftServer)
	createAccounts(router, hand)
	createLogin(router, authManager)
	addBrokerRouter(router, hand)
	addQueueRouter(router, hand)
}

type AckRequest struct {
	Id    string `json:"id"`
	Queue string `json:"queue"`
}

type FetchMsg struct {
	Id   string `json:"id"`
	Data []byte `json:"data"`
}

type CreateQueueReq struct {
	RoutingKey string `json:"routingKey"`
	Name       string `json:"name"`
	Size       uint32 `json:"size"`
	RetryLimit uint32 `json:"retryLimit"`
	HasDLQueue bool   `json:"hasDLQueue"`
}

func addQueueRouter(router fiber.Router, hand handler.Handler) {
	queueRouter := router.Group("/queue")

	queueRouter.Post("/create", func(c *fiber.Ctx) error {
		var req CreateQueueReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		ctx, err := createContextFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "unable to find context information",
			})
		}

		if err := hand.CreateQueue(ctx, req.RoutingKey, req.Name, req.Size, req.RetryLimit, req.HasDLQueue); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	queueRouter.Get("/fetch", func(c *fiber.Ctx) error {
		queue := c.Query("queue")
		if queue == "" {
			return c.Status(fiber.StatusBadRequest).SendString("queue parameter is required")
		}

		amountStr := c.Query("amount")
		u64, err := strconv.ParseUint(amountStr, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing amount parameter")
		}

		ctx, err := createContextFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "unable to find context information",
			})
		}

		data, err := hand.GetMessages(ctx, queue, uint32(u64))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		// convert to uuid format
		formattedData := []FetchMsg{}
		for i := 0; i < len(data); i++ {
			u, err := uuid.FromBytes(data[i].Id)
			if err != nil {
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			formattedData = append(formattedData, FetchMsg{
				Id:   u.String(),
				Data: data[i].Data,
			})
		}

		return c.JSON(formattedData)
	})

	queueRouter.Put("/ack", func(c *fiber.Ctx) error {
		var req AckRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		ctx, err := createContextFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "unable to find context information",
			})
		}

		u, err := uuid.FromString(req.Id)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if err := hand.Ack(ctx, req.Queue, u[:]); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusAccepted)
	})

	queueRouter.Put("/nack", func(c *fiber.Ctx) error {
		var req AckRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		ctx, err := createContextFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "unable to find context information",
			})
		}

		u, err := uuid.FromString(req.Id)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if err := hand.Nack(ctx, req.Queue, u[:]); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusAccepted)
	})
}

type SubmitPayload struct {
	RoutingKey string `json:"routingKey"`
	Data       string `json:"data"`
}

func addBrokerRouter(router fiber.Router, hand handler.Handler) {
	brokerRouter := router.Group("/broker")

	brokerRouter.Post("/submit", func(c *fiber.Ctx) error {
		ctx, err := createContextFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "unable to find context information",
			})
		}

		var payload SubmitPayload
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		if err := hand.SubmitMessage(ctx, payload.RoutingKey, []byte(payload.Data)); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusAccepted)
	})
}

type AccountCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateRoleRequest struct {
	Role string `json:"role"`
}

func createAccounts(router fiber.Router, hand handler.Handler) {
	accountsRouter := router.Group("/accounts")

	accountsRouter.Post("/", func(c *fiber.Ctx) error {
		ctx, err := createContextFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "unable to find context information",
			})
		}

		var credentials AccountCredentials
		if err := c.BodyParser(&credentials); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		if err := hand.CreateUser(ctx, credentials.Username, credentials.Password); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	accountsRouter.Post("/roles/create", func(c *fiber.Ctx) error {
		var req CreateRoleRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		ctx, err := createContextFromFiberContext(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "unable to find context information",
			})
		}

		if err := hand.CreateRole(ctx, req.Role); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	// Adding a new role to an account account
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

	accountsRouter.Delete("/roles/:username/:role", func(c *fiber.Ctx) error {
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

		if err = hand.UnassignRole(ctx, role, username); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.SendStatus(fiber.StatusOK)
	})
}

func createInfo(router fiber.Router, broker core.Broker, auth auth.AuthManager, queueManager core.QueueManager, rbac rbac.RoleManager, raftServer raft.Raft) {
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

	infoRouter.Get("/isLeader", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"isLeader": raftServer.State() == raft.LEADER,
		})
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

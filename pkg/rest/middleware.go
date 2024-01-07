package rest

import (
	"encoding/base64"
	"strings"
	"sybline/pkg/auth"

	"github.com/GreedyKomodoDragon/raft"
	"github.com/gofiber/fiber/v2"
)

func IsLeader(c *fiber.Ctx, raftServer raft.Raft) error {
	// ignore if just checking info -> allows followers to take some of the load
	if strings.HasPrefix(c.Path(), "/info") {
		return c.Next()
	}

	if raftServer.State() != raft.LEADER {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Server is not the leader",
		})
	}

	return c.Next()
}

func Authentication(c *fiber.Ctx, auth auth.AuthManager) error {
	// skip if just checking if leader
	if c.Path() == "/info/leader" {
		return c.Next()
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Authorization header is missing",
		})
	}

	authParts := strings.Split(authHeader, " ")
	if len(authParts) != 2 || authParts[0] != "Basic" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid Authorization header",
		})
	}

	decoded, err := base64.StdEncoding.DecodeString(authParts[1])
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid base64 encoding",
		})
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid credentials format",
		})
	}

	// Simulated user credentials check
	username := credentials[0]
	token := credentials[1]

	conId, err := auth.GetConsumerID(username, token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid credentials provided",
		})
	}

	c.Locals("consumerID", conId)
	c.Locals("username", username)
	c.Locals("token", token)

	return c.Next()
}

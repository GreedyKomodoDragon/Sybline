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
	if strings.Contains(c.Path(), "/info") ||
		strings.Contains(c.Path(), "/login") ||
		strings.Contains(c.Path(), "/metrics") {
		return c.Next()
	}

	if raftServer.State() != raft.LEADER {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Server is not the leader",
		})
	}

	return c.Next()
}

func Authentication(c *fiber.Ctx, authManager auth.AuthManager) error {
	// skip if just checking if leader
	if strings.HasSuffix(c.Path(), "/info/leader") ||
		(strings.HasSuffix(c.Path(), "/login") && c.Method() == "POST") ||
		strings.Contains(c.Path(), "/metrics") {
		return c.Next()
	}

	// Get Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Check if the header is for Basic Authentication
	if !strings.HasPrefix(authHeader, "Basic ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid Authorization header",
		})
	}

	// Decode the base64-encoded username and password
	credentials, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid base64 encoding",
		})
	}

	// Extract username and password
	credentialsParts := strings.SplitN(string(credentials), ":", 2)
	if len(credentialsParts) != 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid credentials format",
		})
	}
	username := credentialsParts[0]
	token := credentialsParts[1]

	// Validate username and password
	id, err := authManager.GetConsumerID(username, token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized, no consumer token found",
		})
	}

	// hold onto this information for later
	c.Locals("consumerID", id)
	c.Locals("username", username)
	c.Locals("token", token)

	// If authentication passes, proceed to the next middleware or route handler
	return c.Next()
}

package rest

import (
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

func Authentication(c *fiber.Ctx, authManager auth.AuthManager) error {
	// skip if just checking if leader
	if c.Path() == "/info/leader" || c.Path() == "/login" {
		return c.Next()
	}

	// Get token and username from cookies
	token := c.Cookies("token")
	username := c.Cookies("username")

	// Check if token and username exist
	if token == "" || username == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	id, err := authManager.GetConsumerID(username, token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// hold onto this information for later
	c.Locals("consumerID", id)
	c.Locals("username", username)
	c.Locals("token", token)

	// If authentication passes, proceed to the next middleware or route handler
	return c.Next()
}

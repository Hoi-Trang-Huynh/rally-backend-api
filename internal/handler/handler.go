package handler

import "github.com/gofiber/fiber/v2"

// HealthCheck godoc
// @Summary Health Check
// @Description Get the health status of the server
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/health [get]
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

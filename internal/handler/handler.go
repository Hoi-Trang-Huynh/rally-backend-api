package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Hoi-Trang-Huynh/rally-backend-api/internal/version"
)

// HealthCheck godoc
// @Summary Health Check
// @Description Get the health status of the server
// @Tags system
// @Accept json
// @Produce json
// @Success 200 
// @Router /health [get]
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// VersionCheck godoc
// @Summary Version Check
// @Description Get the version information of the server
// @Tags system
// @Accept json
// @Produce json
// @Success 200
// @Router /version [get]
func VersionCheck(c *fiber.Ctx) error {
	return c.JSON(version.Info())
}

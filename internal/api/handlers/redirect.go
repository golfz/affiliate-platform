package handlers

import (
	"net"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/service"
)

// RedirectHandler handles redirect-related HTTP requests
type RedirectHandler struct {
	service *service.RedirectService
	logger  logger.Logger
}

// NewRedirectHandler creates a new redirect handler
func NewRedirectHandler(service *service.RedirectService, logger logger.Logger) *RedirectHandler {
	return &RedirectHandler{
		service: service,
		logger:  logger,
	}
}

// Redirect handles GET /go/:short_code
// @Summary Redirect to marketplace product URL
// @Description Redirects to the target marketplace URL and tracks the click event
// @Tags public
// @Accept json
// @Produce json
// @Param short_code path string true "Short code" example:"abc123xyz"
// @Success 302 "Redirect to target URL"
// @Failure 404 {object} dto.ErrorResponse "Link not found"
// @Failure 400 {object} dto.ErrorResponse "Invalid redirect URL"
// @Router /go/{short_code} [get]
func (h *RedirectHandler) Redirect(c echo.Context) error {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "short_code is required",
		})
	}

	// Extract request metadata
	ipAddress := net.ParseIP(c.RealIP())
	if ipAddress == nil {
		// Fallback to X-Forwarded-For header
		ipStr := c.Request().Header.Get("X-Forwarded-For")
		if ipStr != "" {
			ipAddress = net.ParseIP(ipStr)
		}
	}
	userAgent := c.Request().UserAgent()
	referrer := c.Request().Referer()

	// Perform redirect
	targetURL, err := h.service.Redirect(c.Request().Context(), shortCode, ipAddress, userAgent, referrer)
	if err != nil {
		h.logger.Error("Redirect failed", logger.String("error", err.Error()), logger.String("short_code", shortCode))

		if err.Error() == "link not found: record not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Link not found",
			})
		}

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Redirect to target URL
	return c.Redirect(http.StatusFound, targetURL)
}

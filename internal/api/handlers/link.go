package handlers

import (
	"net/http"

	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/service"
	"github.com/labstack/echo/v4"
)

// LinkHandler handles link-related HTTP requests
type LinkHandler struct {
	service *service.LinkService
	logger  logger.Logger
}

// NewLinkHandler creates a new link handler
func NewLinkHandler(service *service.LinkService, logger logger.Logger) *LinkHandler {
	return &LinkHandler{
		service: service,
		logger:  logger,
	}
}

// CreateLink handles POST /api/links
// @Summary Generate affiliate short link
// @Description Generate a short affiliate link for a product/marketplace combination
// @Tags links
// @Accept json
// @Produce json
// @Param request body dto.CreateLinkRequest true "Link creation request"
// @Success 201 {object} dto.LinkResponse "Link created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 404 {object} dto.ErrorResponse "Product or campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/links [post]
func (h *LinkHandler) CreateLink(c echo.Context) error {
	var req dto.CreateLinkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "Invalid request body",
			Code:    "INVALID_INPUT",
		})
	}

	// Basic validation
	if req.Marketplace != "lazada" && req.Marketplace != "shopee" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "marketplace must be 'lazada' or 'shopee'",
			Code:    "INVALID_INPUT",
		})
	}

	// Create link
	link, err := h.service.CreateLink(c.Request().Context(), req)
	if err != nil {
		h.logger.Error("Failed to create link", logger.String("error", err.Error()))

		// Check error type
		if err.Error() == "product not found: record not found" {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Product Not Found",
				Message: "Product with the specified ID was not found",
				Code:    "PRODUCT_NOT_FOUND",
			})
		}
		if err.Error() == "campaign not found: record not found" {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Campaign Not Found",
				Message: "Campaign with the specified ID was not found",
				Code:    "CAMPAIGN_NOT_FOUND",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.JSON(http.StatusCreated, link)
}

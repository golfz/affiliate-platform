package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/service"
	"github.com/labstack/echo/v4"
)

// CampaignHandler handles campaign-related HTTP requests
type CampaignHandler struct {
	service *service.CampaignService
	logger  logger.Logger
}

// NewCampaignHandler creates a new campaign handler
func NewCampaignHandler(service *service.CampaignService, logger logger.Logger) *CampaignHandler {
	return &CampaignHandler{
		service: service,
		logger:  logger,
	}
}

// CreateCampaign handles POST /api/campaigns
// @Summary Create a new campaign
// @Description Create a new marketing campaign with UTM parameters
// @Tags campaigns
// @Accept json
// @Produce json
// @Param request body dto.CreateCampaignRequest true "Campaign creation request"
// @Success 201 {object} dto.CampaignResponse "Campaign created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/campaigns [post]
func (h *CampaignHandler) CreateCampaign(c echo.Context) error {
	var req dto.CreateCampaignRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "Invalid request body",
			Code:    "INVALID_INPUT",
		})
	}

	// Basic validation
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "name is required",
			Code:    "INVALID_INPUT",
		})
	}

	if req.UTMCampaign == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "utm_campaign is required",
			Code:    "INVALID_INPUT",
		})
	}

	// Create campaign
	campaign, err := h.service.CreateCampaign(c.Request().Context(), req)
	if err != nil {
		h.logger.Error("Failed to create campaign", logger.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.JSON(http.StatusCreated, campaign)
}

// GetAllCampaigns handles GET /api/campaigns
// @Summary Get all campaigns
// @Description Get a list of all campaigns with pagination
// @Tags campaigns
// @Accept json
// @Produce json
// @Param limit query int false "Limit number of results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} dto.CampaignResponse "Campaigns retrieved successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/campaigns [get]
func (h *CampaignHandler) GetAllCampaigns(c echo.Context) error {
	// Parse query parameters
	limit := 100 // default limit
	offset := 0  // default offset

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get all campaigns
	campaigns, err := h.service.GetAllCampaigns(c.Request().Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get campaigns", logger.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get campaigns",
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.JSON(http.StatusOK, campaigns)
}

// DeleteCampaign handles DELETE /api/campaigns/:id
// @Summary Delete a campaign
// @Description Delete a campaign and all related data (campaign products, links, clicks)
// @Tags campaigns
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID" format(uuid)
// @Success 204 "Campaign deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid campaign ID"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/campaigns/{id} [delete]
func (h *CampaignHandler) DeleteCampaign(c echo.Context) error {
	campaignIDStr := c.Param("id")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "Invalid campaign ID format",
			Code:    "INVALID_INPUT",
		})
	}

	// Delete campaign
	err = h.service.DeleteCampaign(c.Request().Context(), campaignID)
	if err != nil {
		h.logger.Error("Failed to delete campaign", logger.String("error", err.Error()))

		errMsg := err.Error()
		if strings.Contains(errMsg, "campaign not found") {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Campaign Not Found",
				Message: "Campaign with the specified ID was not found",
				Code:    "CAMPAIGN_NOT_FOUND",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete campaign",
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/service"
	"github.com/labstack/echo/v4"
)

// CampaignPublicHandler handles public campaign-related HTTP requests
type CampaignPublicHandler struct {
	service *service.CampaignPublicService
	logger  logger.Logger
}

// NewCampaignPublicHandler creates a new public campaign handler
func NewCampaignPublicHandler(service *service.CampaignPublicService, logger logger.Logger) *CampaignPublicHandler {
	return &CampaignPublicHandler{
		service: service,
		logger:  logger,
	}
}

// GetPublicCampaign handles GET /api/campaigns/:id/public
// @Summary Get public campaign details
// @Description Get campaign details with products and offers for public landing page
// @Tags public
// @Accept json
// @Produce json
// @Param id path string true "Campaign ID" format(uuid)
// @Success 200 {object} dto.CampaignPublicResponse "Campaign details retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Campaign not found"
// @Failure 400 {object} dto.ErrorResponse "Campaign is not active"
// @Router /api/campaigns/{id}/public [get]
func (h *CampaignPublicHandler) GetPublicCampaign(c echo.Context) error {
	campaignIDStr := c.Param("id")
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "Invalid campaign ID format",
			Code:    "INVALID_INPUT",
		})
	}

	// Get public campaign
	campaign, err := h.service.GetPublicCampaign(c.Request().Context(), campaignID)
	if err != nil {
		h.logger.Error("Failed to get public campaign", logger.String("error", err.Error()))

		errMsg := err.Error()
		if strings.Contains(errMsg, "campaign not found") {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Campaign Not Found",
				Message: "Campaign with the specified ID was not found",
				Code:    "CAMPAIGN_NOT_FOUND",
			})
		}

		if strings.Contains(errMsg, "campaign is not active") {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Campaign Not Active",
				Message: "Campaign is not currently active",
				Code:    "CAMPAIGN_NOT_ACTIVE",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get campaign",
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.JSON(http.StatusOK, campaign)
}

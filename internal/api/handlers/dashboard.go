package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/service"
	"github.com/labstack/echo/v4"
)

// DashboardHandler handles dashboard-related HTTP requests
type DashboardHandler struct {
	service *service.DashboardService
	logger  logger.Logger
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(service *service.DashboardService, logger logger.Logger) *DashboardHandler {
	return &DashboardHandler{
		service: service,
		logger:  logger,
	}
}

// GetDashboardStats handles GET /api/dashboard
// @Summary Get dashboard statistics
// @Description Get aggregated click statistics, CTR, and top-performing products
// @Tags dashboard
// @Accept json
// @Produce json
// @Param campaign_id query string false "Filter by campaign ID" format(uuid)
// @Param marketplace query string false "Filter by marketplace (lazada or shopee)"
// @Param start_date query string false "Start date filter (RFC3339)" example:"2025-01-01T00:00:00Z"
// @Param end_date query string false "End date filter (RFC3339)" example:"2025-12-31T23:59:59Z"
// @Success 200 {object} dto.DashboardStatsResponse "Dashboard statistics retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/dashboard [get]
func (h *DashboardHandler) GetDashboardStats(c echo.Context) error {
	params := dto.DashboardQueryParams{}

	// Parse campaign_id filter
	if campaignIDStr := c.QueryParam("campaign_id"); campaignIDStr != "" {
		campaignID, err := uuid.Parse(campaignIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Invalid Input",
				Message: "Invalid campaign_id format",
				Code:    "INVALID_INPUT",
			})
		}
		params.CampaignID = &campaignID
	}

	// Parse marketplace filter
	if marketplace := c.QueryParam("marketplace"); marketplace != "" {
		if marketplace != "lazada" && marketplace != "shopee" {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Invalid Input",
				Message: "marketplace must be 'lazada' or 'shopee'",
				Code:    "INVALID_INPUT",
			})
		}
		params.Marketplace = &marketplace
	}

	// Parse start_date filter
	if startDateStr := c.QueryParam("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Invalid Input",
				Message: "Invalid start_date format (expected RFC3339)",
				Code:    "INVALID_INPUT",
			})
		}
		params.StartDate = &startDate
	}

	// Parse end_date filter
	if endDateStr := c.QueryParam("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Invalid Input",
				Message: "Invalid end_date format (expected RFC3339)",
				Code:    "INVALID_INPUT",
			})
		}
		params.EndDate = &endDate
	}

	// Get dashboard stats
	stats, err := h.service.GetDashboardStats(c.Request().Context(), params)
	if err != nil {
		h.logger.Error("Failed to get dashboard stats", logger.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get dashboard statistics",
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.JSON(http.StatusOK, stats)
}

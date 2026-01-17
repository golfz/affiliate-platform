package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/jonosize/affiliate-platform/internal/dto"
	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/service"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	service *service.ProductService
	logger  logger.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(service *service.ProductService, logger logger.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

// CreateProduct handles POST /api/products
// @Summary Add product from Lazada/Shopee URL or SKU
// @Description Create a new product by fetching data from marketplace URL or SKU
// @Tags products
// @Accept json
// @Produce json
// @Param request body dto.CreateProductRequest true "Product creation request"
// @Success 201 {object} dto.ProductResponse "Product created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/products [post]
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	var req dto.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "Invalid request body",
			Code:    "INVALID_INPUT",
		})
	}

	// Basic validation
	if req.Source == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "source is required",
			Code:    "INVALID_INPUT",
		})
	}

	if req.SourceType == "" {
		req.SourceType = "url" // Default to URL
	}

	// Create product
	product, err := h.service.CreateProduct(c.Request().Context(), req)
	if err != nil {
		h.logger.Error("Failed to create product", logger.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create product",
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.JSON(http.StatusCreated, product)
}

// GetProductOffers handles GET /api/products/:id/offers
// @Summary Get offers (prices) for a product
// @Description Get all marketplace offers for a specific product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID" format(uuid)
// @Success 200 {object} dto.ProductOffersResponse "Product offers retrieved successfully"
// @Failure 404 {object} dto.ErrorResponse "Product not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/products/{id}/offers [get]
func (h *ProductHandler) GetProductOffers(c echo.Context) error {
	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "Invalid product ID format",
			Code:    "INVALID_INPUT",
		})
	}

	// Get product offers
	response, err := h.service.GetProductOffers(c.Request().Context(), productID)
	if err != nil {
		h.logger.Error("Failed to get product offers", logger.String("error", err.Error()))

		// Check if product not found
		if err.Error() == "product not found: record not found" {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Product Not Found",
				Message: "Product with the specified ID was not found",
				Code:    "PRODUCT_NOT_FOUND",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get product offers",
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.JSON(http.StatusOK, response)
}

// GetAllProducts handles GET /api/products
// @Summary Get all products
// @Description Get a list of all products with pagination
// @Tags products
// @Accept json
// @Produce json
// @Param limit query int false "Limit number of results" default(100)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} dto.ProductResponse "Products retrieved successfully"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/products [get]
func (h *ProductHandler) GetAllProducts(c echo.Context) error {
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

	// Get all products
	products, err := h.service.GetAllProducts(c.Request().Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get products", logger.String("error", err.Error()))
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get products",
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.JSON(http.StatusOK, products)
}

// DeleteProduct handles DELETE /api/products/:id
// @Summary Delete a product
// @Description Delete a product and all related data (offers, links, campaign associations, clicks)
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID" format(uuid)
// @Success 204 "Product deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid product ID"
// @Failure 404 {object} dto.ErrorResponse "Product not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	productIDStr := c.Param("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid Input",
			Message: "Invalid product ID format",
			Code:    "INVALID_INPUT",
		})
	}

	// Delete product
	err = h.service.DeleteProduct(c.Request().Context(), productID)
	if err != nil {
		h.logger.Error("Failed to delete product", logger.String("error", err.Error()))

		errMsg := err.Error()
		if strings.Contains(errMsg, "product not found") {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Product Not Found",
				Message: "Product with the specified ID was not found",
				Code:    "PRODUCT_NOT_FOUND",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete product",
			Code:    "INTERNAL_ERROR",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

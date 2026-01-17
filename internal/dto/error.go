package dto

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error" example:"Error Type"`
	Message string                 `json:"message" example:"Human-readable message"`
	Code    string                 `json:"code,omitempty" example:"ERROR_CODE"`
	Details map[string]interface{} `json:"details,omitempty"`
}

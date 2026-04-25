package common

import "github.com/gofiber/fiber/v2"

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type ErrorResponse struct {
	Success bool     `json:"success"`
	Error   APIError `json:"error"`
}

func NewError(code, message string, details interface{}) APIError {
	return APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func Error(c *fiber.Ctx, status int, apiError APIError) error {
	return c.Status(status).JSON(ErrorResponse{
		Success: false,
		Error:   apiError,
	})
}

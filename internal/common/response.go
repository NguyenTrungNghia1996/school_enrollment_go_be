package common

import "github.com/gofiber/fiber/v2"

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedListData struct {
	Items      interface{}    `json:"items"`
	Pagination PaginationMeta `json:"pagination"`
}

func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Paginated(c *fiber.Ctx, status int, message string, items interface{}, pagination PaginationMeta) error {
	return Success(c, status, message, PaginatedListData{
		Items:      items,
		Pagination: pagination,
	})
}

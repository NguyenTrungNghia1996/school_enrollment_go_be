package common

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
	SortAsc         = "asc"
	SortDesc        = "desc"
)

type PaginationQuery struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Keyword   string `json:"keyword,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

func ParsePagination(c *fiber.Ctx) (PaginationQuery, *APIError) {
	query := PaginationQuery{
		Page:      DefaultPage,
		PageSize:  DefaultPageSize,
		Keyword:   strings.TrimSpace(c.Query("keyword")),
		SortBy:    strings.TrimSpace(c.Query("sort_by")),
		SortOrder: normalizeSortOrder(c.Query("sort_order")),
	}

	if page := strings.TrimSpace(c.Query("page")); page != "" {
		parsed, err := strconv.Atoi(page)
		if err != nil || parsed < 1 {
			apiError := NewError("INVALID_PAGINATION", "page must be an integer greater than or equal to 1", nil)
			return PaginationQuery{}, &apiError
		}
		query.Page = parsed
	}

	if pageSize := strings.TrimSpace(c.Query("page_size")); pageSize != "" {
		parsed, err := strconv.Atoi(pageSize)
		if err != nil || parsed < 1 {
			apiError := NewError("INVALID_PAGINATION", "page_size must be an integer greater than or equal to 1", nil)
			return PaginationQuery{}, &apiError
		}
		if parsed > MaxPageSize {
			parsed = MaxPageSize
		}
		query.PageSize = parsed
	}

	if query.SortOrder == "" {
		query.SortOrder = SortDesc
	}

	if query.SortOrder != SortAsc && query.SortOrder != SortDesc {
		apiError := NewError("INVALID_SORT_ORDER", "sort_order must be either asc or desc", nil)
		return PaginationQuery{}, &apiError
	}

	return query, nil
}

func (p PaginationQuery) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func BuildPaginationMeta(query PaginationQuery, totalItems int64) PaginationMeta {
	totalPages := int64(0)
	if totalItems > 0 {
		totalPages = (totalItems + int64(query.PageSize) - 1) / int64(query.PageSize)
	}

	return PaginationMeta{
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func normalizeSortOrder(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

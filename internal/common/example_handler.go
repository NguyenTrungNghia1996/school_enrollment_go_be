package common

import "github.com/gofiber/fiber/v2"

type ExampleItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func ExamplePaginatedHandler(c *fiber.Ctx) error {
	pagination, apiError := ParsePagination(c)
	if apiError != nil {
		return Error(c, fiber.StatusBadRequest, *apiError)
	}

	items := []ExampleItem{
		{ID: 1, Name: "Example 1"},
		{ID: 2, Name: "Example 2"},
	}

	return Paginated(
		c,
		fiber.StatusOK,
		"Fetched example items successfully",
		items,
		BuildPaginationMeta(pagination, int64(len(items))),
	)
}

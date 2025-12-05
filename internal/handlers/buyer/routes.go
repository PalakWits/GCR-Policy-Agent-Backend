package buyer

import "github.com/gofiber/fiber/v2"

func (h *BuyerHandler) RegisterRoutes(app *fiber.App) {
	routes := app.Group("/v1")
	routes.Post("/permissions", h.UpdateBapAccessPermissions)
	routes.Post("/permissions/query", h.QueryBapAccessPermissions)
}

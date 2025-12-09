package broadcast

import "github.com/gofiber/fiber/v2"

func (h *BroadcastHandler) RegisterRoutes(app *fiber.App) {
	routes := app.Group("/v1")
	routes.Post("/permissions/broadcast", h.BroadcastPermissions)
	routes.Get("/permissions/broadcast/status/:job_id", h.GetBroadcastStatus)
}

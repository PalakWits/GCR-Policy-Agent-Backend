package seller

import "github.com/gofiber/fiber/v2"

func (h *SellerHandler) RegisterRoutes(app *fiber.App) {
	routes := app.Group("/v1")
	routes.Get("/catalog-sync/pending", h.GetPendingCatalogSyncSellers)
	routes.Get("/catalog-sync/sellers/:seller_id", h.GetSyncStatus)
	internal := routes.Group("/internal")
	internal.Post("/registry-sync", h.SyncRegistry)
}

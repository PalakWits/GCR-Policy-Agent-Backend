package seller

import (
	"adapter/internal/domain/seller"
	sellerPorts "adapter/internal/ports/seller"
	"adapter/internal/shared/constants"
	"adapter/internal/shared/utils"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
)

type SellerHandler struct {
	sellerService *seller.SellerService
}

func NewSellerHandler(sellerService *seller.SellerService) *SellerHandler {
	return &SellerHandler{sellerService: sellerService}
}

func (h *SellerHandler) SyncRegistry(c *fiber.Ctx) error {
	var req sellerPorts.SellerRegistrySyncRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrInvalidRequestBody,
		})
	}

	if len(req.Domains) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: "domains are required",
		})
	}

	response, err := h.sellerService.SyncRegistry(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrFailedToStartRegistrySync,
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.ApiResponse{
		Success: true,
		Message: "Registry sync completed successfully",
		Data:    response,
	})
}

func (h *SellerHandler) GetPendingCatalogSyncSellers(c *fiber.Ctx) error {
	domain := c.Query("domain")
	if domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrDomainRequired,
		})
	}
	status := c.Query("status")
	limit := c.QueryInt("limit", 100)
	page := c.QueryInt("page", 1)
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	response, err := h.sellerService.GetPendingCatalogSyncSellers(domain, status, limit, page, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrGetPendingSellers,
		})
	}
	return c.Status(fiber.StatusOK).JSON(utils.ApiResponse{
		Success: true,
		Message: "Pending catalog sync sellers retrieved successfully",
		Data:    response,
	})
}

func (h *SellerHandler) GetSyncStatus(c *fiber.Ctx) error {
	sellerID := c.Params("seller_id")
	domain := c.Query("domain")

	if sellerID == "" || domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrSellerIDAndDomainRequired,
		})
	}

	response, err := h.sellerService.GetSyncStatus(sellerID, domain)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(utils.ApiResponse{
				Success: false,
				Message: constants.ErrRecordNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrGetSyncStatus,
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.ApiResponse{
		Success: true,
		Message: "Sync status retrieved successfully",
		Data:    response,
	})
}

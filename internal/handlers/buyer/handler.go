package buyer

import (
	buyerDomain "adapter/internal/domain/buyer"
	buyerPorts "adapter/internal/ports/buyer"
	sellerPorts "adapter/internal/ports/seller"
	"adapter/internal/shared/constants"
	"adapter/internal/shared/utils"

	"github.com/gofiber/fiber/v2"
)

type BuyerHandler struct {
	permissionsService *buyerDomain.BuyerService
}

func NewBuyerHandler(permissionsService *buyerDomain.BuyerService) *BuyerHandler {
	return &BuyerHandler{permissionsService: permissionsService}
}

func (h *BuyerHandler) UpdateBapAccessPermissions(c *fiber.Ctx) error {
	var req struct {
		Updates []sellerPorts.SellerPermissionsUpdateRequest `json:"updates"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrInvalidRequestBody,
		})
	}

	if len(req.Updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrUpdatesArrayEmpty,
		})
	}

	results, err := h.permissionsService.UpdateBapAccessPermissions(req.Updates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrFailedToUpdatePermissions,
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.ApiResponse{
		Success: true,
		Message: "Permissions updated successfully",
		Data:    fiber.Map{"results": results},
	})
}

func (h *BuyerHandler) QueryBapAccessPermissions(c *fiber.Ctx) error {
	var req buyerPorts.BapPermissionsQueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrInvalidRequestBody,
		})
	}

	if req.BapID == "" || req.Domain == "" || req.RegistryEnv == "" || len(req.SellerIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrRequiredPermissionsFields,
		})
	}

	response, err := h.permissionsService.QueryBapAccessPermissions(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrFailedToQueryPermissions,
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.ApiResponse{
		Success: true,
		Message: "Permissions queried successfully",
		Data:    response,
	})
}

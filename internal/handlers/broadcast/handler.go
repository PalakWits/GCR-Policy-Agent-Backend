package broadcast

import (
	broadcastDomain "adapter/internal/domain/broadcast"
	broadcastPorts "adapter/internal/ports/broadcast"
	"adapter/internal/shared/constants"
	"adapter/internal/shared/utils"

	"github.com/gofiber/fiber/v2"
)

type BroadcastHandler struct {
	service *broadcastDomain.BroadcastService
}

func NewBroadcastHandler(service *broadcastDomain.BroadcastService) *BroadcastHandler {
	return &BroadcastHandler{service: service}
}

func (h *BroadcastHandler) BroadcastPermissions(c *fiber.Ctx) error {
	var req broadcastPorts.BroadcastRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: constants.ErrInvalidRequestBody,
		})
	}

	if err := h.service.BroadcastPermissions(req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(utils.ApiResponse{
		Success: true,
		Message: "Broadcast initiated successfully",
	})
}

func (h *BroadcastHandler) GetBroadcastStatus(c *fiber.Ctx) error {
	bapID := c.Params("bap_id")
	if bapID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: "bap_id is required",
		})
	}

	job, err := h.service.GetBroadcastStatus(bapID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.ApiResponse{
		Success: true,
		Message: "Broadcast job status fetched successfully",
		Data:    job,
	})
}

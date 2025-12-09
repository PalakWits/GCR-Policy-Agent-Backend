package broadcast

import (
	broadcastDomain "adapter/internal/domain/broadcast"
	broadcastPorts "adapter/internal/ports/broadcast"
	"adapter/internal/shared/constants"
	"adapter/internal/shared/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

	job, err := h.service.BroadcastPermissions(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(utils.ApiResponse{
		Success: true,
		Message: "Broadcast initiated successfully",
		Data:    job,
	})
}

func (h *BroadcastHandler) GetBroadcastStatus(c *fiber.Ctx) error {
	jobIDStr := c.Params("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
			Success: false,
			Message: "Invalid job_id format",
		})
	}

	job, err := h.service.GetBroadcastStatus(jobID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(utils.ApiResponse{
				Success: false,
				Message: "Broadcast job not found",
			})
		}
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

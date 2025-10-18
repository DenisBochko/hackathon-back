package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"hackathon-back/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// --- Обработчики ---

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusInvalidInput,
			Message: "invalid user id",
		})
		return
	}

	if err := h.service.SoftDeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusInternalError,
			Message: "failed to soft delete user",
		})
		return
	}

	c.JSON(http.StatusOK, ResponseWithMessage{
		Status:  StatusOK,
		Message: "user soft deleted",
	})
}

type BlockRequest struct {
	Block bool `json:"block"`
}

func (h *UserHandler) BlockUser(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusInvalidInput,
			Message: "invalid user id",
		})
		return
	}

	var req BlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusInvalidInput,
			Message: "invalid request body",
		})
		return
	}

	if err := h.service.BlockUser(c.Request.Context(), userID, req.Block); err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusInternalError,
			Message: "failed to update block status",
		})
		return
	}

	c.JSON(http.StatusOK, ResponseWithMessage{
		Status:  StatusOK,
		Message: "user block status updated",
	})
}

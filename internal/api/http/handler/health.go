package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"hackathon-back/internal/apperrors"
	"hackathon-back/internal/model"
)

type HealthService interface {
	IsOK() (bool, error)
	GetTestData(ctx context.Context) (*model.TestTable, error)
}

type HealthHandler struct {
	log *zap.Logger
	svc HealthService
}

func NewHealthHandler(log *zap.Logger, svc HealthService) *HealthHandler {
	return &HealthHandler{
		log: log,
		svc: svc,
	}
}

// Ping
// @Summary Проверка здоровья сервиса.
// @Description Возвращает “pong”.
// @Tags Health
// @Produce json
// @Success 200 {object} ResponseWithMessage "Success"
// @Router /health/ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, ResponseWithMessage{
		Status:  StatusSuccess,
		Message: "pong",
	})
}

func (h *HealthHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()

	_, err := h.svc.IsOK()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusErr,
			Message: err.Error(),
		})

		return
	}

	data, err := h.svc.GetTestData(ctx)
	if err != nil {
		if errors.Is(err, apperrors.ErrTestDataDoesNotExist) {
			c.JSON(http.StatusNotFound, ResponseWithMessage{
				Status:  StatusErr,
				Message: err.Error(),
			})

			return
		}

		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusErr,
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, ResponseWithData{
		Status: StatusSuccess,
		Data:   data,
	})
}

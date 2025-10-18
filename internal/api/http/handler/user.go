package handler

import (
	"context"
	"errors"
	"hackathon-back/internal/apperrors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"hackathon-back/internal/model"
)

type UserService interface {
	GetUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	BlockUser(ctx context.Context, id uuid.UUID) error

	// TODO: UploadUserPhoto(ctx context.Context, ...) (..., error)
	// TODO: GetUserPhoto(ctx context.Context, userID uuid.UUID) (string, error) // Возвращает ссылку на получение изображения из minIO
}

type UserHandler struct {
	BaseHandler

	svc UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{
		svc: service,
	}
}

// DeleteUser
// @Summary	Удаляет пользователя по id.
// @Description Удаляет пользователя по id, доступно для пользователей с ролью manager и выше.
// @Tags User
// @Security AccessToken
// @Security RefreshToken
// @Produce json
// @Param user_id path string true "User UUID"
// @Success 200 {object} ResponseWithMessage "Success"
// @Failure 400 {object} ResponseWithMessage "Invalid path param"
// @Failure 500 {object} ResponseWithMessage "Failed to delete user"
// @Router /user/{user_id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()

	var uri model.UserIDPathParam
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusErr,
			Message: err.Error(),
		})

		return
	}

	userUID, err := uuid.Parse(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusErr,
			Message: err.Error(),
		})

		return
	}

	if err := h.svc.DeleteUser(ctx, userUID); err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusInternalError,
			Message: "Failed to delete user",
		})

		return
	}

	c.JSON(http.StatusOK, ResponseWithMessage{
		Status:  StatusSuccess,
		Message: "User deleted successfully",
	})
}

// BlockUser
// @Summary	Блокирует пользователя по id.
// @Description Блокирует пользователя по id, доступно для пользователей с ролью manager и выше.
// @Tags User
// @Security AccessToken
// @Security RefreshToken
// @Produce json
// @Param user_id path string true "User UUID"
// @Success 200 {object} ResponseWithMessage "Success"
// @Failure 400 {object} ResponseWithMessage "Invalid path param"
// @Failure 500 {object} ResponseWithMessage "Failed to block user"
// @Router /user/block/{user_id} [post]
func (h *UserHandler) BlockUser(c *gin.Context) {
	ctx := c.Request.Context()

	var uri model.UserIDPathParam
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusErr,
			Message: err.Error(),
		})

		return
	}

	userUID, err := uuid.Parse(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusErr,
			Message: err.Error(),
		})

		return
	}

	if err := h.svc.BlockUser(ctx, userUID); err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusInternalError,
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, ResponseWithMessage{
		Status:  StatusSuccess,
		Message: "User blocked successfully",
	})
}

// GetUser
// @Summary	Получить пользователя по id.
// @Description Получить пользователя по id.
// @Tags User
// @Produce json
// @Param user_id path string true "User UUID"
// @Success 200 {object} ResponseWithData{data=model.User} "Success"
// @Failure 400 {object} ResponseWithMessage "Invalid path param"
// @Failure 404 {object} ResponseWithMessage "User not found"
// @Failure 500 {object} ResponseWithMessage "Failed to get user"
// @Router /user/{user_id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	ctx := c.Request.Context()

	var uri model.UserIDPathParam
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusErr,
			Message: err.Error(),
		})

		return
	}

	userUID, err := uuid.Parse(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusErr,
			Message: err.Error(),
		})

		return
	}

	user, err := h.svc.GetUser(ctx, userUID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserDoesNotExist) {
			c.JSON(http.StatusNotFound, ResponseWithMessage{
				Status:  StatusErr,
				Message: err.Error(),
			})

			return
		}

		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusInternalError,
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, ResponseWithData{
		Status: StatusSuccess,
		Data:   user,
	})
}

// GetUserJWT
// @Summary	Получить пользователя по id.
// @Description Получить пользователя по id. id берётся из JWT токена.
// @Tags User
// @Security AccessToken
// @Security RefreshToken
// @Produce json
// @Success 200 {object} ResponseWithData{data=model.User} "Success"
// @Failure 400 {object} ResponseWithMessage "Invalid path param"
// @Failure 401 {object} ResponseWithMessage "Invalid or missing token"
// @Failure 403 {object} ResponseWithMessage "Invalid user data format"
// @Failure 404 {object} ResponseWithMessage "User not found"
// @Failure 500 {object} ResponseWithMessage "Failed to get user"
// @Router /user [get]
func (h *UserHandler) GetUserJWT(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := h.GetUserID(c)
	if err != nil {
		if errors.Is(err, apperrors.ErrContextValueDoesNotExist) {
			c.JSON(http.StatusUnauthorized, ResponseWithMessage{
				Status:  StatusNotPermitted,
				Message: "no data about the user",
			})

			return
		}

		if errors.Is(err, apperrors.ErrContextValueInvalidType) {
			c.JSON(http.StatusForbidden, ResponseWithMessage{
				Status:  StatusNotPermitted,
				Message: "invalid user data format",
			})

			return
		}
	}

	user, err := h.svc.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserDoesNotExist) {
			c.JSON(http.StatusNotFound, ResponseWithMessage{
				Status:  StatusErr,
				Message: err.Error(),
			})

			return
		}

		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusInternalError,
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, ResponseWithData{
		Status: StatusSuccess,
		Data:   user,
	})
}

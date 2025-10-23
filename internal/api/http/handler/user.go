package handler

import (
	"errors"
	"hackathon-back/internal/apperrors"
	"hackathon-back/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"hackathon-back/internal/model"
)

type UserHandler struct {
	BaseHandler

	svc service.UserService
}

// ForgotPasswordRequest
// @Description Запрос на восстановление пароля.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"` // Почта, на которую придёт письмо для восстановления пароля
} // @Name ForgotPasswordRequest

// ResetPasswordRequest
// @Description Запрос на сброс пароля.
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`       // Токен, полученный по ссылке из письма
	NewPassword string `json:"newPassword" binding:"required"` // Новый пароль
} // @Name ResetPasswordRequest

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
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

// UploadUserPhoto
// @Summary Загрузить или заменить своё фото профиля
// @Description Загрузка новой фотографии пользователя (или замена старой). Требуется авторизация.
// @Tags User
// @Security AccessToken
// @Security RefreshToken
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Фото пользователя"
// @Success 200 {object} ResponseWithData{data=string} "URL загруженного фото"
// @Failure 400 {object} ResponseWithMessage "Некорректный запрос"
// @Failure 401 {object} ResponseWithMessage "Неавторизован"
// @Failure 500 {object} ResponseWithMessage "Ошибка сервера"
// @Router /user/photo [post]
func (h *UserHandler) UploadUserPhoto(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := h.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ResponseWithMessage{
			Status:  StatusNotPermitted,
			Message: "unauthorized",
		})
		return
	}

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{
			Status:  StatusErr,
			Message: "failed to read file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	photoURL, err := h.svc.UploadUserPhoto(ctx, userID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusInternalError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ResponseWithData{
		Status: StatusSuccess,
		Data:   photoURL,
	})
}

// GetUserPhoto
// @Summary Получить ссылку на своё фото профиля
// @Description Возвращает URL изображения пользователя из MinIO. Требуется авторизация.
// @Tags User
// @Security AccessToken
// @Security RefreshToken
// @Produce json
// @Success 200 {object} ResponseWithData{data=string} "URL фото"
// @Failure 401 {object} ResponseWithMessage "Неавторизован"
// @Failure 404 {object} ResponseWithMessage "Фото не найдено"
// @Failure 500 {object} ResponseWithMessage "Ошибка сервера"
// @Router /user/photo [get]
func (h *UserHandler) GetUserPhoto(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := h.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ResponseWithMessage{
			Status:  StatusNotPermitted,
			Message: "unauthorized",
		})
		return
	}

	url, err := h.svc.GetUserPhoto(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserDoesNotExist) {
			c.JSON(http.StatusNotFound, ResponseWithMessage{
				Status:  StatusErr,
				Message: "user not found",
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
		Data:   url,
	})
}

// DeleteUserPhoto
// @Summary Удалить своё фото профиля
// @Description Удаляет фотографию пользователя из MinIO. Требуется авторизация.
// @Tags User
// @Security AccessToken
// @Security RefreshToken
// @Produce json
// @Success 200 {object} ResponseWithMessage "Фото успешно удалено"
// @Failure 401 {object} ResponseWithMessage "Неавторизован"
// @Failure 500 {object} ResponseWithMessage "Ошибка сервера"
// @Router /user/photo [delete]
func (h *UserHandler) DeleteUserPhoto(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := h.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ResponseWithMessage{
			Status:  StatusNotPermitted,
			Message: "unauthorized",
		})
		return
	}

	if err := h.svc.DeleteUserPhoto(ctx, userID); err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{
			Status:  StatusInternalError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ResponseWithMessage{
		Status:  StatusSuccess,
		Message: "photo deleted successfully",
	})
}

// ForgotPassword
// @Summary Запросить сброс пароля
// @Description Отправляет письмо со ссылкой для восстановления пароля на указанный email.
// @Tags user
// @Accept json
// @Produce json
// @Param input body ForgotPasswordRequest true "Email пользователя"
// @Success 200 {object} ResponseWithMessage "Письмо успешно отправлено"
// @Failure 400 {object} ResponseWithMessage "Некорректный запрос"
// @Failure 500 {object} ResponseWithMessage "Ошибка сервера"
// @Router /user/password/forgot [post]
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{Status: StatusErr, Message: err.Error()})
		return
	}

	if err := h.svc.RequestPasswordReset(c, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{Status: StatusInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ResponseWithMessage{Status: StatusSuccess, Message: "Password reset link sent to email"})
}

// ResetPassword
// @Summary Сбросить пароль
// @Description Принимает токен и новый пароль. После успешной смены токен становится недействительным.
// @Tags user
// @Accept json
// @Produce json
// @Param input body ResetPasswordRequest true "Данные для сброса пароля"
// @Success 200 {object} ResponseWithMessage "Пароль успешно изменён"
// @Failure 400 {object} ResponseWithMessage "Некорректные данные"
// @Failure 500 {object} ResponseWithMessage "Ошибка сервера"
// @Router /user/password/reset [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseWithMessage{Status: StatusErr, Message: err.Error()})
		return
	}

	if err := h.svc.ResetPassword(c, req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{Status: StatusInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ResponseWithMessage{Status: StatusSuccess, Message: "Password reset successful"})
}

// DeleteSelf
// @Summary Удалить свой аккаунт
// @Description Авторизованный пользователь может удалить свой аккаунт. Данные помечаются как удалённые.
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ResponseWithMessage "Аккаунт удалён"
// @Failure 401 {object} ResponseWithMessage "Пользователь не авторизован"
// @Failure 500 {object} ResponseWithMessage "Ошибка сервера"
// @Router /user/self [delete]
func (h *UserHandler) DeleteSelf(c *gin.Context) {
	userID, err := h.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ResponseWithMessage{Status: StatusNotPermitted, Message: "unauthorized"})
		return
	}

	if err := h.svc.DeleteSelf(c, userID); err != nil {
		c.JSON(http.StatusInternalServerError, ResponseWithMessage{Status: StatusInternalError, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ResponseWithMessage{Status: StatusSuccess, Message: "User deleted"})
}

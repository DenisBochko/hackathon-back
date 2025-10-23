package route

import (
	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	GetUser(*gin.Context)
	GetUserJWT(c *gin.Context)
	DeleteUser(c *gin.Context)
	BlockUser(c *gin.Context)

	UploadUserPhoto(c *gin.Context)
	GetUserPhoto(c *gin.Context)
	DeleteUserPhoto(c *gin.Context)

	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	DeleteSelf(c *gin.Context)
}

func RegisterAdminUserRoutes(g *gin.RouterGroup, h UserHandler, jwtAuthMiddleware, allowManagerAndAdminMiddleware gin.HandlerFunc) {
	g.GET("/:user_id", h.GetUser)

	protected := g.Group("", jwtAuthMiddleware)
	protected.GET("", h.GetUserJWT)

	//Фото доступны для всех авторизованных
	protected.POST("/photo", h.UploadUserPhoto)
	protected.GET("/photo", h.GetUserPhoto)
	protected.DELETE("/photo", h.DeleteUserPhoto)

	//Восстановление и сброс пароля
	protected.POST("/password/forgot", h.ForgotPassword)
	protected.POST("/password/reset", h.ResetPassword)
	protected.DELETE("/reset", h.DeleteSelf)

	adminOrManagerRequired := protected.Group("", allowManagerAndAdminMiddleware)
	adminOrManagerRequired.DELETE(":user_id", h.DeleteUser)
	adminOrManagerRequired.POST("/block/:user_id", h.BlockUser)
}

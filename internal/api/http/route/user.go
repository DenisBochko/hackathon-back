package route

import (
	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	GetUser(*gin.Context)
	GetUserJWT(c *gin.Context)
	DeleteUser(c *gin.Context)
	BlockUser(c *gin.Context)
}

func RegisterAdminUserRoutes(g *gin.RouterGroup, h UserHandler, jwtAuthMiddleware, allowManagerAndAdminMiddleware gin.HandlerFunc) {
	g.GET("", h.GetUser)

	protected := g.Group("", jwtAuthMiddleware)
	protected.GET(":user_id", h.GetUserJWT)

	adminOrManagerRequired := protected.Group("", allowManagerAndAdminMiddleware)
	adminOrManagerRequired.DELETE(":user_id", h.DeleteUser)
	adminOrManagerRequired.POST("/block/:user_id", h.BlockUser)
}

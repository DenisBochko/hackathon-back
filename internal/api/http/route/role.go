package route

import (
	"github.com/gin-gonic/gin"
)

// UserHandler описывает методы, которые может выполнять админ или менеджер
type UserHandler interface {
	DeleteUser(c *gin.Context)
	BlockUser(c *gin.Context)
}

// RegisterAdminUserRoutes — подключает все маршруты админки
func RegisterAdminUserRoutes(g *gin.RouterGroup, h UserHandler, middlewares ...gin.HandlerFunc) {
	admin := g.Group("/admin")
	admin.Use(middlewares...) // JWT + проверка ролей

	users := admin.Group("/users")
	users.DELETE("/:id", h.DeleteUser)
	users.PATCH("/:id/block", h.BlockUser)
}

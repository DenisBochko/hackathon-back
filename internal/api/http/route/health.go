package route

import (
	"github.com/gin-gonic/gin"
)

type HealthHandler interface {
	Ping(c *gin.Context)
	Health(c *gin.Context)
}

func RegisterHealth(g *gin.RouterGroup, h HealthHandler) {
	g.GET("", h.Health)
	g.GET("/ping", h.Ping)
}

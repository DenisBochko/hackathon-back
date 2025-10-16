package middleware

import (
	"github.com/gin-gonic/gin"
	"hackathon-back/internal/api/http/handler"
	"hackathon-back/internal/model"
	"net/http"
)

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// берем роль из контекста
		roleVal, exists := c.Get(model.UserRoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, handler.ResponseWithMessage{
				Status:  handler.StatusNotPermitted,
				Message: "user role not found in context",
			})
			return
		}

		// валидируем ее
		role, ok := roleVal.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, handler.ResponseWithMessage{
				Status:  handler.StatusNotPermitted,
				Message: "invalid role format",
			})
		}

		// проверяем, есть ли роль среди разрешенных
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		// если не одна не подошла - отказ
		c.AbortWithStatusJSON(http.StatusForbidden, handler.ResponseWithMessage{
			Status:  handler.StatusNotPermitted,
			Message: "access denied: insufficient permissions",
		})
	}
}

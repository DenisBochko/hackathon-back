package route

import (
	"crypto/ecdsa"
	"io"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"hackathon-back/internal/api/http/handler"
	"hackathon-back/internal/api/http/middleware"
	"hackathon-back/internal/config"
	"hackathon-back/internal/model"
)

const maxMultipartMemory = 1 << 30

func SetupRouter(
	log *zap.Logger,
	cfg *config.Config,
	publicKey *ecdsa.PublicKey,
	healthHdl HealthHandler,
	authHdl AuthHandler,
	userHdl UserHandler,
	articleHdl ArticleHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	router := gin.Default()
	router.MaxMultipartMemory = maxMultipartMemory

	// middleware
	router.Use(middleware.Logger(log))
	router.Use(middleware.RequestTimeout(cfg.HTTPServer.Timeout.Request))
	router.Use(middleware.CORS(cfg.CORS))

	jwtAuthMiddleware := middleware.JWTAuth(publicKey)
	allowManagerAndAdminMiddleware := middleware.RequireRoles(model.RoleManager, model.RoleAdmin)

	router.HandleMethodNotAllowed = true
	router.NoMethod(handler.NoMethod)
	router.NoRoute(handler.NoRoute)

	basePath := router.Group(cfg.BasePath)

	docsPath := basePath.Group("/docs")
	RegisterDock(docsPath)

	healthPath := basePath.Group("/health")
	RegisterHealth(healthPath, healthHdl, jwtAuthMiddleware)

	authPath := basePath.Group("/auth")
	RegisterAuth(authPath, authHdl)

	userPath := basePath.Group("/user")
	RegisterAdminUserRoutes(userPath, userHdl, jwtAuthMiddleware, allowManagerAndAdminMiddleware)

	articlePath := basePath.Group("/article")
	RegisterArticleRoutes(articlePath, articleHdl, jwtAuthMiddleware)

	return router
}

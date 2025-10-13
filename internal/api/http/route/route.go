package route

import (
	"crypto/ecdsa"
	"io"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"hackathon-back/internal/api/http/handler"
	"hackathon-back/internal/api/http/middleware"
	"hackathon-back/internal/config"
)

const maxMultipartMemory = 1 << 30

func SetupRouter(
	log *zap.Logger,
	cfg *config.Config,
	publicKey *ecdsa.PublicKey,
	healthHdl HealthHandler,
	authHdl AuthHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	router := gin.Default()

	router.MaxMultipartMemory = maxMultipartMemory

	router.Use(middleware.Logger(log))
	router.Use(middleware.RequestTimeout(cfg.Timeout.Request))
	router.Use(middleware.CORS(cfg.CORS))

	jwtAuthMiddleware := middleware.JWTAuth(publicKey)

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

	return router
}

package app

import (
	"context"
	"errors"
	"fmt"
	"hackathon-back/internal/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"hackathon-back/internal/api/http/handler"
	"hackathon-back/internal/api/http/route"
	"hackathon-back/internal/apperrors"
	"hackathon-back/internal/config"
	"hackathon-back/internal/repository"
	"hackathon-back/internal/service"
	"hackathon-back/pkg/postgres"
	"hackathon-back/pkg/redis"
	"hackathon-back/pkg/server"
)

type HealthRepository interface {
	IsOK() (bool, error)
	SelectData(ctx context.Context, ext repository.RepoExtension) (*model.TestTable, error)
}

type HealthService interface {
	IsOK() (bool, error)
	GetTestData(ctx context.Context) (*model.TestTable, error)
}

type HealthHandler interface {
	Ping(c *gin.Context)
	Health(c *gin.Context)
}

type App struct {
	Cfg        *config.Config
	Log        *zap.Logger
	Handler    *Handler
	Service    *Service
	DB         postgres.Postgres
	RDB        redis.Redis
	HTTPServer server.HTTPServer
}

type Repository struct {
	HealthRepository HealthRepository
}

type Service struct {
	HealthService HealthService
}

type Handler struct {
	HealthHandler HealthHandler
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	db, err := initDB(&cfg.Database)
	if err != nil {
		log.Error("Failed to initialize database", zap.Error(err))

		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	rdb, err := initRedis(&cfg.Redis)
	if err != nil {
		log.Error("Failed to initialize redis", zap.Error(err))

		return nil, fmt.Errorf("failed to initialize redis: %w", err)
	}

	repo := initRepository(log, db)

	svc := initService(log, repo)

	hdl := initHandler(log, svc)

	httpServer := initHTTPServer(log, cfg, hdl)

	return &App{
		Cfg:        cfg,
		Log:        log,
		Handler:    hdl,
		Service:    svc,
		DB:         db,
		RDB:        rdb,
		HTTPServer: httpServer,
	}, nil
}

func MustNew(cfg *config.Config, log *zap.Logger) *App {
	app, err := New(cfg, log)
	if err != nil {
		panic(err)
	}

	return app
}

func (a *App) Run() error {
	errs := make(chan error, 1)
	defer close(errs)

	go func() {
		if err := a.HTTPServer.Run(); err != nil {
			errs <- err
		}
	}()

	if err := <-errs; err != nil {
		return err
	}

	return nil
}

func (a *App) Shutdown() error {
	a.DB.Close()

	a.Log.Debug("Database closed")

	err := apperrors.ErrShutdown

	if rdbErr := a.RDB.Close(); rdbErr != nil {
		err = fmt.Errorf("%w, failed to close RDB: %w", err, rdbErr)
	}

	a.Log.Debug("Redis closed")

	if srvErr := a.HTTPServer.Shutdown(); srvErr != nil {
		err = fmt.Errorf("%w, failed to shutdown http server: %w", err, srvErr)
	}

	a.Log.Debug("Http server shutdown")

	if !errors.Is(err, apperrors.ErrShutdown) {
		return err
	}

	return nil
}

func initDB(cfg *config.Database) (postgres.Postgres, error) {
	postgresCfg := &postgres.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Name:     cfg.Name,
		SSLMode:  cfg.SSLMode,
		MaxConns: cfg.MaxConns,
		MinConns: cfg.MinConns,
		Migration: postgres.Migration{
			Path:      cfg.Migration.Path,
			AutoApply: cfg.Migration.AutoApply,
		},
	}

	db, err := postgres.New(postgresCfg)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initRedis(cfg *config.Redis) (redis.Redis, error) {
	redisCfg := &redis.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	rdb, err := redis.New(redisCfg)
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

func initHandler(log *zap.Logger, svc *Service) *Handler {
	healthHandler := handler.NewHealthHandler(log, svc.HealthService)

	log.Debug("Health handler initialized")

	return &Handler{
		HealthHandler: healthHandler,
	}
}

func initService(log *zap.Logger, repo *Repository) *Service {
	healthSvc := service.NewHealthService(log, repo.HealthRepository)

	log.Debug("Health service initialized")

	return &Service{
		HealthService: healthSvc,
	}
}

func initRepository(log *zap.Logger, db postgres.Postgres) *Repository {
	healthRepo := repository.NewHealthRepository(db.Pool())

	log.Debug("Health repository initialized")

	return &Repository{
		HealthRepository: healthRepo,
	}
}

func initHTTPServer(log *zap.Logger, cfg *config.Config, hdl *Handler) server.HTTPServer {
	router := route.SetupRouter(log, cfg, hdl.HealthHandler)

	httpServer := server.NewHTTPServer(
		server.WithAddr(cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		server.WithTimeout(cfg.Timeout.Read, cfg.Timeout.Write, cfg.Timeout.Idle),
		server.WithHandler(router),
	)

	return httpServer
}

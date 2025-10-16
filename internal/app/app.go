package app

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"hackathon-back/internal/api/http/handler"
	"hackathon-back/internal/api/http/route"
	"hackathon-back/internal/apperrors"
	"hackathon-back/internal/config"
	"hackathon-back/internal/model"
	"hackathon-back/internal/repository"
	"hackathon-back/internal/service"
	"hackathon-back/pkg/jwt"
	"hackathon-back/pkg/mailer"
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
	ProtectedPing(c *gin.Context)
	Health(c *gin.Context)
}

type AuthRepository interface {
	Pool() *pgxpool.Pool

	InsertUser(ctx context.Context, ext repository.RepoExtension, user *model.User) (*model.User, error)
	SelectUserByID(ctx context.Context, ext repository.RepoExtension, id uuid.UUID) (*model.User, error)
	SelectUserByEmail(ctx context.Context, ext repository.RepoExtension, email string) (*model.User, error)
	UpdateUserAsConfirmed(ctx context.Context, ext repository.RepoExtension, userID uuid.UUID) error
	InsertVerificationToken(ctx context.Context, ext repository.RepoExtension, verificationToken *model.VerificationToken) error
	SelectVerificationToken(ctx context.Context, ext repository.RepoExtension, token []byte) (*model.VerificationToken, error)
	DeleteVerificationTokenByUserID(ctx context.Context, ext repository.RepoExtension, userID uuid.UUID) error
}

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (user *model.User, userToken []byte, err error)
	ResendConfirmation(ctx context.Context, email string) ([]byte, error)
	Confirmation(ctx context.Context, incCode string, incToken []byte) error
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
	TestLogin(ctx context.Context) (accessToken, refreshToken string, err error)
}

type AuthHandler interface {
	Register(c *gin.Context)
	ResendConfirmation(c *gin.Context)
	Confirmation(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	Refresh(c *gin.Context)
	TestLogin(c *gin.Context)
}

type App struct {
	Cfg        *config.Config
	Log        *zap.Logger
	Handler    *Handler
	Service    *Service
	Security   *Security
	DB         postgres.Postgres
	RDB        redis.Redis
	Mailer     mailer.Mailer
	HTTPServer server.HTTPServer
}

type Repository struct {
	HealthRepository HealthRepository
	AuthRepository   AuthRepository
	UserRepository   *repository.UserRepository
}

type Service struct {
	HealthService HealthService
	AuthService   AuthService
	UserService   *service.UserService
}

type Handler struct {
	HealthHandler HealthHandler
	AuthHandler   AuthHandler
	UserHandler   *handler.UserHandler
}

type Security struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
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

	sec, err := initSecurity(log, cfg.Key)
	if err != nil {
		log.Error("Failed to initialize security", zap.Error(err))

		return nil, fmt.Errorf("failed to initialize security: %w", err)
	}

	mlr := initMailer(log, &cfg.Mailer)

	repo := initRepository(log, db)

	svc := initService(log, &cfg.JWT, sec, repo, mlr, rdb)

	hdl := initHandler(log, &cfg.JWT, svc)

	httpServer := initHTTPServer(log, cfg, sec.PublicKey, hdl)

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

func initMailer(log *zap.Logger, cfg *config.Mailer) mailer.Mailer {
	mailerCfg := &mailer.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		From:     cfg.From,
		UseTLS:   cfg.UseTLS,
	}

	mlr := mailer.New(mailerCfg)

	log.Debug("Mailer initialized")

	return mlr
}

func initSecurity(log *zap.Logger, cfg config.Key) (*Security, error) {
	privateKey, err := jwt.LoadECDSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	log.Debug("Private key loaded")

	publicKey, err := jwt.LoadECDSAPublicKey(cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	log.Debug("Public key loaded")

	return &Security{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func initHandler(log *zap.Logger, jwtCfg *config.JWT, svc *Service) *Handler {
	healthHandler := handler.NewHealthHandler(log, svc.HealthService)
	log.Debug("Health handler initialized")

	authHandler := handler.NewAuthHandler(log, svc.AuthService, jwtCfg.AccessTokenTTL, jwtCfg.RefreshTokenTTL)
	log.Debug("Auth handler initialized")

	userHandler := handler.NewUserHandler(svc.UserService)
	log.Debug("User handler initialized")

	return &Handler{
		HealthHandler: healthHandler,
		AuthHandler:   authHandler,
		UserHandler:   userHandler,
	}
}

func initService(
	log *zap.Logger,
	jwtCfg *config.JWT,
	sec *Security,
	repo *Repository,
	mlr mailer.Mailer,
	rdb redis.Redis,
) *Service {
	healthSvc := service.NewHealthService(log, repo.HealthRepository)
	log.Debug("Health service initialized")

	authSvc := service.NewAuthService(log, sec.PublicKey, sec.PrivateKey, repo.AuthRepository, mlr, rdb, jwtCfg.AccessTokenTTL, jwtCfg.RefreshTokenTTL)
	log.Debug("Auth service initialized")

	userSvc := service.NewUserService(repo.UserRepository)
	log.Debug("User service initialized")

	return &Service{
		HealthService: healthSvc,
		AuthService:   authSvc,
		UserService:   userSvc,
	}
}

func initRepository(log *zap.Logger, db postgres.Postgres) *Repository {
	healthRepo := repository.NewHealthRepository(db.Pool())
	log.Debug("Health repository initialized")

	authRepo := repository.NewAuthRepository(db.Pool())
	log.Debug("Auth repository initialized")

	userRepo := repository.NewUserRepository(db.Pool())
	log.Debug("User repository initialized")

	return &Repository{
		HealthRepository: healthRepo,
		AuthRepository:   authRepo,
		UserRepository:   userRepo,
	}
}

func initHTTPServer(log *zap.Logger, cfg *config.Config, publicKey *ecdsa.PublicKey, hdl *Handler) server.HTTPServer {
	router := route.SetupRouter(
		log,
		cfg,
		publicKey,
		hdl.HealthHandler,
		hdl.AuthHandler,
		hdl.UserHandler,
	)

	httpServer := server.NewHTTPServer(
		server.WithAddr(cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		server.WithTimeout(cfg.Timeout.Read, cfg.Timeout.Write, cfg.Timeout.Idle),
		server.WithHandler(router),
	)

	return httpServer
}

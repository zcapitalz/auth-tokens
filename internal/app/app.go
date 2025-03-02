package app

import (
	"auth/internal/config"
	authcontroller "auth/internal/controllers/auth-controller"
	"auth/internal/db/postgres"
	authservice "auth/internal/domain/services/auth-service"
	emailservice "auth/internal/domain/services/email-service"
	"auth/internal/repositories"
	slogutils "auth/internal/utils/slog-utils"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "auth/internal/controllers/swagger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

const (
	gracefulServerShutdownTimeout = time.Second * 5
	swaggerSpecURLPath            = "api/swagger"
)

//	@title			Access tokens management service
//	@version		1.0
//	@description	Service that manages access and refresh tokens.

// @BasePath	/
func Run(cfg *config.Config) error {
	logger, err := newLogger(cfg)
	if err != nil {
		return errors.Wrap(err, "create logger")
	}
	slog.SetDefault(logger)

	if err := migrateDatabase(cfg.DBConfig); err != nil {
		return err
	}

	db, err := postgres.NewDatabase(cfg.DBConfig)
	if err != nil {
		return errors.Wrap(err, "failed to init repository")
	}

	refreshTokenRepository := repositories.NewRefreshTokenRepository(db)
	userRepository := repositories.NewUserRepositoryMock()

	emailService := emailservice.NewEmailService(
		cfg.Emails, cfg.SMTPServer, userRepository)
	authService := authservice.NewAuthService(
		refreshTokenRepository, emailService,
		[]byte(cfg.Auth.JWTPrivateKey),
		cfg.Auth.AccessTokenDuration, cfg.Auth.RefreshTokenDuration)

	authController := authcontroller.NewAuthController(authService)

	switch cfg.Env {
	case config.EnvLocal:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(
		requestid.New(
			requestid.WithGenerator(func() string {
				return ksuid.New().String()
			}),
		),
		setLoggerMiddleware())
	engine.GET(swaggerSpecURLPath+"/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	authController.RegisterRoutes(engine)

	srv := &http.Server{
		Addr:    cfg.HTTPServer.Host + ":" + cfg.HTTPServer.Port,
		Handler: engine.Handler(),
	}
	err = runServer(srv)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slogutils.Error("run server", err)
	}

	return err
}

func runServer(srv *http.Server) error {
	slog.Info("server is starting", "address", srv.Addr)
	defer slog.Info("server exited")

	systemSignalQuit := make(chan os.Signal, 1)
	signal.Notify(systemSignalQuit, syscall.SIGINT, syscall.SIGTERM)

	serverExited := make(chan error, 1)
	go func() {
		serverExited <- errors.Wrap(srv.ListenAndServe(), "server listen")
	}()

	select {
	case <-systemSignalQuit:
		ctx, cancel := context.WithTimeout(context.Background(), gracefulServerShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "graceful server shutdown")
		}
		return nil
	case err := <-serverExited:
		return err
	}
}

func setLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := slog.Default()
		if traceID, ok := c.Get("requestID"); ok {
			logger = logger.With("requestID", traceID.(string))
		}
		c.Set("logger", logger)

		c.Next()
	}
}

func migrateDatabase(dbConfig config.DatabaseConfig) error {
	migrator, err := migrate.New(
		"file://./migrations/postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			dbConfig.Username, dbConfig.Password,
			dbConfig.Host, dbConfig.Port,
			dbConfig.DBName, dbConfig.SSLMode))
	if err != nil {
		return errors.Wrap(err, "create migrator")
	}
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "migrate")
	}

	return nil
}

func newLogger(cfg *config.Config) (*slog.Logger, error) {
	var logLevel slog.Level
	err := logLevel.UnmarshalText([]byte(cfg.LogLevel))
	if err != nil {
		return nil, errors.Wrap(err, "parse log level")
	}

	var logger *slog.Logger
	switch cfg.Env {
	case config.EnvLocal:
		logger = newPrettyLogger(logLevel)
	case config.EnvDev, config.EnvProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}),
		)
	default:
		return nil, fmt.Errorf("unknown env: %s", cfg.Env)
	}

	return logger, nil
}

func newPrettyLogger(logLevel slog.Level) *slog.Logger {
	opts := slogutils.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: logLevel,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

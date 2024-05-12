package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"p3/app/config"
	"p3/app/handler"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	logger *zap.Logger
	conf   *config.AppConfig
)

func init() {
	var cfg zap.Config
	var err error
	conf = config.ConfigLoader()

	switch conf.LogEnv {
	case "dev":
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(t.Format(time.RFC3339Nano)) // Encode time in RFC339Nano format
		}
	case "prod":
		cfg = zap.NewProductionConfig()
	default:
		panic(fmt.Errorf("invalid log environment: %s", conf.LogEnv))
	}

	logger, err = cfg.Build()
	if err != nil {
		panic(fmt.Errorf("error creating logger: %s", err.Error()))
	}

	defer logger.Sync()
}

func main() {
	// ctx bg
	ctx := context.Background()
	// New context
	logger.Info("test")

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
			)
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// new handler
	handler := handler.NewHandler()

	// Routes
	e.GET("/health", handler.HealthCheck)
	// e.POST("/broadcast", BroadcastTransaction)
	// e.GET("/check/:tx_hash", MonitorTransaction)

	// Start server
	go func() {
		if err := e.Start(conf.AppPort); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()
	gracefulShutdown(ctx, e)

}

func gracefulShutdown(ctx context.Context, e *echo.Echo) {
	// context with timeout
	quitSig := make(chan os.Signal)
	signal.Notify(
		quitSig,
		syscall.SIGINT, // ctrl + c
		syscall.SIGKILL,
	)

	select {
	case <-ctx.Done():
		log.Println("terminating via context cancel")
	case <-quitSig:
		log.Println("terminating via signal")
	}
	ctx, cancel := context.WithTimeout(ctx, conf.CtxTimeOut)

	defer cancel()
	err := e.Shutdown(ctx)
	if err != nil {
		e.Logger.Fatal(err)
	}
}

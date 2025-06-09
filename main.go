package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
)

var logger *zap.Logger

func LoggerInit() {
	cfg := zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		OutputPaths:      []string{"stdout", "auth.log"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
	var err error
	logger, err = cfg.Build()
	if err != nil {
		os.Stderr.WriteString("Failed to create logger" + err.Error() + "\n")
		os.Exit(1)
	}
	lj := &lumberjack.Logger{
		Filename:   "auth.log",
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg.EncoderConfig),
		zapcore.AddSync(lj),
		cfg.Level,
	)

	logger = zap.New(core)
}

func main() {
	LoggerInit()
	defer func() {
		_ = logger.Sync()
	}()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		logger.Debug("Request",
			zap.String("ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery))
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		query := c.Query("auth")
		if query != "Ivan" {
			logger.Warn("Unauthorized")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
			})
			return
		} else {
			logger.Info("Authorized")
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
			})
			logger.Info("OK")
		}
	})
	logger.Info("Starting server")
	err := r.Run(":8080")
	if err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

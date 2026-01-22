package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewJSONEncoder(config)
	consoleWriter := zapcore.AddSync(os.Stdout)

	core := zapcore.NewCore(consoleEncoder, consoleWriter, zapcore.InfoLevel)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	Logger.Info("Logger initialized successfully")
}

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

type Logger struct {
	*zap.Logger
}

func NewLogger() (*Logger, error) {
	logWriter := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxAge:     28,
		MaxBackups: 3,
		Compress:   true,
	}

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(config), zapcore.AddSync(logWriter), zapcore.InfoLevel), zapcore.NewCore(zapcore.NewConsoleEncoder(config), zapcore.AddSync(os.Stdout), zapcore.InfoLevel))

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return &Logger{logger}, nil

}

func (l *Logger) Sync() {
	l.Logger.Sync()
}

package logger

import (
    "os"
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Logger interface {
    Info(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
    Fatal(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Debug(msg string, keysAndValues ...interface{})
}

type zapLogger struct {
    logger *zap.SugaredLogger
}

func NewLogger(level string) Logger {
    config := zap.NewProductionConfig()
    
    switch level {
    case "debug":
        config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
    case "info":
        config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    case "warn":
        config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
    case "error":
        config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
    default:
        config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    }

    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

    logger, err := config.Build()
    if err != nil {
        os.Exit(1)
    }

    return &zapLogger{
        logger: logger.Sugar(),
    }
}

func (l *zapLogger) Info(msg string, keysAndValues ...interface{})  { l.logger.Infow(msg, keysAndValues...) }
func (l *zapLogger) Error(msg string, keysAndValues ...interface{}) { l.logger.Errorw(msg, keysAndValues...) }
func (l *zapLogger) Fatal(msg string, keysAndValues ...interface{}) { l.logger.Fatalw(msg, keysAndValues...) }
func (l *zapLogger) Warn(msg string, keysAndValues ...interface{})  { l.logger.Warnw(msg, keysAndValues...) }
func (l *zapLogger) Debug(msg string, keysAndValues ...interface{}) { l.logger.Debugw(msg, keysAndValues...) }

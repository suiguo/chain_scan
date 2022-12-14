package utils

import (
	"fmt"
	"os"
	"sync"

	"github.com/natefinch/lumberjack"
	zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//Logger is the wrapper
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Panic(msg string, fields ...interface{})
}

//StdLogger is struct
type StdLogger struct {
	logger *zap.SugaredLogger
	Logger
}
type LoggerCfg struct {
	Name       string `json:"name"`
	Maxsize    int    `json:"maxsize"`
	Maxbackups int    `json:"maxbackups"`
	Maxage     int    `json:"maxage"`
	Compress   bool   `json:"compress"`
	Level      int    `json:"level"`
}

var logger_map = make(map[string]*StdLogger)
var logger_lock sync.RWMutex

func NewLog(cfgs []*LoggerCfg) (*StdLogger, error) {
	if cfgs == nil || len(cfgs) <= 0 {
		return nil, fmt.Errorf("logger cfg is err")
	}
	cores := make([]zapcore.Core, 0)
	encoder := getEncoder()
	for _, cfg := range cfgs {
		switch cfg.Name {
		case "stdout":
			cores = append(cores, zapcore.NewCore(encoder, os.Stdout, zapcore.Level(cfg.Level)))
		case "stderr":
			cores = append(cores, zapcore.NewCore(encoder, os.Stderr, zapcore.Level(cfg.Level)))
		default:
			cores = append(cores, zapcore.NewCore(encoder, getCfgWriter(cfg), zapcore.Level(cfg.Level)))
		}
	}
	handler := zapcore.NewTee(cores...)
	zaplogger := zap.New(handler, zap.AddCaller(), zap.WithCaller(false)) //不打印堆栈
	sugarLogger := zaplogger.Sugar()
	return &StdLogger{
		logger: sugarLogger,
	}, nil
}

func GetInstance(tag string, cfg []*LoggerCfg) (*StdLogger, error) {
	if cfg == nil || len(cfg) <= 0 {
		return nil, fmt.Errorf("logger cfg is err")
	}
	logger_lock.RLock()
	instance := logger_map[tag]
	logger_lock.RUnlock()
	if instance == nil {
		cores := make([]zapcore.Core, 0)
		encoder := getEncoder()
		for _, v := range cfg {
			if v == nil {
				return nil, fmt.Errorf("cfg is nil")
			}
			switch v.Name {
			case "stdout":
				cores = append(cores, zapcore.NewCore(encoder, os.Stdout, zapcore.Level(v.Level)))
			case "stderr":
				cores = append(cores, zapcore.NewCore(encoder, os.Stderr, zapcore.Level(v.Level)))
			default:
				cores = append(cores, zapcore.NewCore(encoder, getCfgWriter(v), zapcore.Level(v.Level)))
			}
		}
		handler := zapcore.NewTee(cores...)
		zaplogger := zap.New(handler, zap.AddCaller(), zap.WithCaller(false)) //不打印堆栈
		sugarLogger := zaplogger.Sugar()
		instance = &StdLogger{
			logger: sugarLogger,
		}
		logger_lock.Lock()
		logger_map[tag] = instance
		logger_lock.Unlock()
	}
	return instance, nil
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getCfgWriter(cfg *LoggerCfg) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.Name,
		MaxSize:    cfg.Maxsize,
		MaxBackups: cfg.Maxbackups,
		MaxAge:     cfg.Maxage,
		Compress:   cfg.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

//Debug is for log warning level
func (l *StdLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debugw(msg, fields...)
}

//Info is for log warning level
func (l *StdLogger) Info(msg string, fields ...interface{}) {
	l.logger.Infow(msg, fields...)
}

//Error is for log warning level
func (l *StdLogger) Error(msg string, fields ...interface{}) {
	l.logger.Errorw(msg, fields...)
}

//Fatal is for log warning level
func (l *StdLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatalw(msg, fields...)
}

//Panic is for log warning level
func (l *StdLogger) Panic(msg string, fields ...interface{}) {
	l.logger.Panicw(msg, fields...)
}

func (l *StdLogger) Warning(msg string, fields ...interface{}) {
	l.logger.Warnw(msg, fields...)
}

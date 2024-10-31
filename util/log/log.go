package log

import (
	"fmt"
	"github.com/spf13/cast"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"runtime/debug"
	"strings"
)

var (
	log           *zap.Logger
	useDebugPanic = atomic.NewBool(false)
	logFunc       map[zapcore.Level]func(template string, args ...interface{})
)

type GormLog struct {
	l     *zap.SugaredLogger
	Debug bool
}

func initOutFunc() {
	if len(logFunc) == 0 {
		logFunc = map[zapcore.Level]func(template string, args ...interface{}){
			zapcore.DebugLevel:  log.Sugar().Debugf,
			zapcore.InfoLevel:   log.Sugar().Infof,
			zapcore.WarnLevel:   log.Sugar().Warnf,
			zapcore.ErrorLevel:  log.Sugar().Errorf,
			zapcore.FatalLevel:  log.Sugar().Fatalf,
			zapcore.PanicLevel:  log.Sugar().Panicf,
			zapcore.DPanicLevel: log.Sugar().DPanicf,
		}
	}
}

func NewGormLog() GormLog {
	if log == nil {
		InitLog(Options{Level: zapcore.DebugLevel, UseDebugPanic: true})
		log.Warn("log not init, auto init log level debug")
	}
	initOutFunc()
	databaseShowLogDebug := func() bool {
		if os.Getenv("DATABASE_SHOW_LOG_DEBUG") == "" {
			return true
		}
		return cast.ToBool(os.Getenv("DATABASE_SHOW_LOG_DEBUG"))
	}
	return GormLog{l: log.Sugar().With(zap.String("logType", "database")), Debug: databaseShowLogDebug()}
}

func (g GormLog) Print(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	sqlField, ok := v[0].(string)
	if !ok && g.Debug {
		g.l.Debug(v...)
		return
	}
	switch sqlField {
	case "sql":
		if !g.Debug {
			return
		}
		if len(v) < 6 {
			g.l.Debug(v...)
			return
		}
		log.Debug("sql",
			zap.Any("sql", v[3]),
			zap.Any("params", v[4]),
			zap.Any("result_count", v[5]),
			zap.String("parse_time", cast.ToDuration(v[2]).String()),
		)
	case "log":
		if len(v) < 3 {
			g.l.Debug(v...)
			return
		}
		err, ok := v[2].(error)
		if !ok {
			err = fmt.Errorf("%v", v[2])
		}
		log.Error("sql message", zap.Any("line", v[1]), zap.Error(err))

	}
}

type Options struct {
	Level         zapcore.Level
	Sync          []io.Writer
	UseDebugPanic bool
}

func InitLog(opt Options) {
	useDebugPanic.Store(opt.UseDebugPanic)

	encoderConfig := &zapcore.EncoderConfig{
		TimeKey:        "times",
		LevelKey:       "severity",
		NameKey:        "logger",
		MessageKey:     "message",
		StacktraceKey:  "trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05"),
		EncodeDuration: zapcore.MillisDurationEncoder,
	}
	if opt.Level <= zapcore.DebugLevel {
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		encoderConfig.CallerKey = "sourceLocation"
	}
	atomicLevel := zap.NewAtomicLevelAt(opt.Level)
	var ws []zapcore.WriteSyncer
	for _, v := range opt.Sync {
		ws = append(ws, zapcore.AddSync(v))
	}

	ws = append(ws, zapcore.AddSync(os.Stdout))

	core := zapcore.NewCore(zapcore.NewJSONEncoder(*encoderConfig),
		zapcore.NewMultiWriteSyncer(ws...), atomicLevel)

	var options = []zap.Option{zap.AddCallerSkip(2)}
	if opt.Level <= zap.DebugLevel {
		options = append(options, zap.AddCaller())
	}
	log = zap.New(core, options...)
}

func StrLevel2zAPlEVEL(level string) zapcore.Level {

	switch strings.ToLower(level) {
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.DebugLevel
	}
}

// outByLog: format https://cloud.google.com/logging/docs/structured-logging
func outByLog(level zapcore.Level, msg string, v ...interface{}) {
	if log == nil {
		InitLog(Options{Level: zapcore.DebugLevel, UseDebugPanic: true})
		initOutFunc()
		log.Warn("log not init, auto init log level debug")
	}

	var out = log.Sugar().Debugf
	if _, ok := logFunc[level]; ok {
		out = logFunc[level]
	}
	out(msg, v...)
	if level == zapcore.DPanicLevel && useDebugPanic.Load() {
		// DPanic 没有效果?
		debug.PrintStack()
		os.Exit(1)
	}
}

func Debug(msg string, v ...interface{}) {
	outByLog(zapcore.DebugLevel, msg, v...)
}

func Info(msg string, v ...interface{}) {
	outByLog(zapcore.InfoLevel, msg, v...)
}

func Warn(msg string, v ...interface{}) {
	outByLog(zapcore.WarnLevel, msg, v...)
}

func Error(msg string, v ...interface{}) {
	outByLog(zapcore.ErrorLevel, msg, v...)
}

func Panic(msg string, v ...interface{}) {
	outByLog(zapcore.PanicLevel, msg, v...)
}

func DPanic(msg string, v ...interface{}) {
	outByLog(zapcore.DPanicLevel, msg, v...)
}

func Fatal(msg string, v ...interface{}) {
	outByLog(zapcore.FatalLevel, msg, v...)
}

func Log() *zap.Logger {
	return log
}

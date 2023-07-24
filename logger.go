package sq

import (
	zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"runtime/debug"
)

type Logger interface {
	Debug(message string, keysAndValues ...interface{})
	Info(message string, keysAndValues ...interface{})
	Warn(message string, keysAndValues ...interface{})
	Error(message string, keysAndValues ...interface{})
	Sync() error
}

var Log = NewZapLogger()

type DefaultLog struct {
	core *zap.SugaredLogger
}

func NewZapLogger() Logger {
	// 编码
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
		zap.LevelEnablerFunc(func(curentLevel zapcore.Level) bool {
			return true
		}),
	)

	options := []zap.Option{
		zap.AddCaller(),
		// zap.AddStacktrace(zapcore.DebugLevel),
		// zap.AddStacktrace(zapcore.InfoLevel),
		zap.AddStacktrace(zapcore.WarnLevel),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCallerSkip(3),
	}

	return &DefaultLog{
		core: zap.New(core, options...).Sugar(),
	}
}
func (z DefaultLog) log(fn func(msg string, keysAndValues ...interface{}), msg string, keysAndValues ...interface{}) {
	defer func() {
		r := recover()
		if r != nil {
			log.Print(r)
			debug.PrintStack()
		}
	}()
	fn(msg, keysAndValues...)
}
func (z DefaultLog) Debug(message string, keysAndValues ...interface{}) {
	z.log(z.core.Debugw, message, keysAndValues...)
}
func (z DefaultLog) Info(message string, keysAndValues ...interface{}) {
	z.log(z.core.Infow, message, keysAndValues...)
}
func (z DefaultLog) Warn(message string, keysAndValues ...interface{}) {
	z.log(z.core.Warnw, message, keysAndValues...)
}
func (z DefaultLog) Error(message string, keysAndValues ...interface{}) {
	z.log(z.core.Errorw, message, keysAndValues...)
}
func (z DefaultLog) Panic(message string, keysAndValues ...interface{}) {
	z.log(z.core.Panicw, message, keysAndValues...)
}
func (z DefaultLog) DPanic(message string, keysAndValues ...interface{}) {
	z.log(z.core.DPanicw, message, keysAndValues...)
}
func (z DefaultLog) Fatal(message string, keysAndValues ...interface{}) {
	z.log(z.core.Fatalw, message, keysAndValues...)
}
func (z DefaultLog) Sync() error {
	return z.core.Sync()
}

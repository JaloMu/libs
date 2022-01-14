package mysql

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/JaloMu/libs/trace"

	"go.uber.org/zap"
	gormLogger "gorm.io/gorm/logger"
)

var (
	gormPackage    = filepath.Join("gorm.io", "gorm")
	zapGormPackage = filepath.Join("moul.io", "zapgorm2")
)

const (
	GORMERROR = "_gorm_error"
	GORMWARN  = "_gorm_warn"
	GROMDEBUY = "_gorm_debug"
)

type Logger struct {
	ZapLogger        *zap.Logger
	LogLevel         gormLogger.LogLevel
	SlowThreshold    time.Duration
	SkipCallerLookup bool
	TraceKey         string
}

func New(zapLogger *zap.Logger, traceKeys string) Logger {
	return Logger{
		ZapLogger:        zapLogger,
		LogLevel:         gormLogger.Warn,
		SlowThreshold:    100 * time.Millisecond,
		SkipCallerLookup: false,
		TraceKey:         traceKeys,
	}
}

func (l Logger) SetAsDefault() {
	gormLogger.Default = l
}

func (l Logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return Logger{
		ZapLogger:        l.ZapLogger,
		SlowThreshold:    l.SlowThreshold,
		LogLevel:         level,
		SkipCallerLookup: l.SkipCallerLookup,
		TraceKey:         l.TraceKey,
	}
}

func (l Logger) Info(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormLogger.Info {
		return
	}
	l.logger().Sugar().Debugf(str, args...)
}

func (l Logger) Warn(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormLogger.Warn {
		return
	}
	l.logger().Sugar().Warnf(str, args...)
}

func (l Logger) Error(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormLogger.Error {
		return
	}
	l.logger().Sugar().Errorf(str, args...)
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	tr := ctx.Value(l.TraceKey)
	traceContext, ok := tr.(*trace.TraceContext)
	if !ok {
		traceContext = &trace.TraceContext{}
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormLogger.Error:
		sql, rows := fc()
		l.logger().Error(GORMERROR,
			zap.Error(err),
			zap.String("traceId", traceContext.TraceId),
			zap.String("child_spanId", traceContext.CSpanId),
			zap.String("spanId", traceContext.SpanId),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows), zap.String("sql", sql))
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		l.logger().Warn(GORMWARN,
			zap.String("traceId", traceContext.TraceId),
			zap.String("child_spanId", traceContext.CSpanId),
			zap.String("spanId", traceContext.SpanId),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql))
	case l.LogLevel >= gormLogger.Info:
		sql, rows := fc()
		l.logger().Debug(GROMDEBUY,
			zap.String("traceId", traceContext.TraceId),
			zap.String("child_spanId", traceContext.CSpanId),
			zap.String("spanId", traceContext.SpanId),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql))
	}
}

func (l Logger) logger() *zap.Logger {
	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.Contains(file, gormPackage):
		case strings.Contains(file, zapGormPackage):
		default:
			return l.ZapLogger.WithOptions(zap.AddCallerSkip(i))
		}
	}
	return l.ZapLogger
}

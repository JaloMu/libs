package logger

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/JaloMu/libs/response"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var lg *zap.Logger

type LogConfig struct {
	Level            string `json:"level" mapstructure:"level,omitempty" yaml:"level" toml:"level"`                         // 日志级别
	Filename         string `json:"filename" mapstructure:"filename,omitempty" yaml:"filename" toml:"filename"`             // 文件名称
	MaxSize          int    `json:"maxsize" mapstructure:"maxsize,omitempty" yaml:"maxsize" toml:"maxsize"`                 // 日志大小 M
	MaxAge           int    `json:"max_age" mapstructure:"max_age,omitempty" yaml:"max_age" toml:"max_age"`                 // 保存多少天
	MaxBackups       int    `json:"max_backups" mapstructure:"max_backups,omitempty" yaml:"max_backups" toml:"max_backups"` // 保存多少个
	Encoder          string `json:"encoder" mapstructure:"encoder,omitempty" yaml:"encoder" toml:"encoder"`                 // 输出格式
	Color            bool   `json:"color" mapstructure:"color,omitempty" yaml:"color" toml:"color"`                         // 是否输出颜色
	ConsoleSeparator string `json:"console_separator" mapstructure:"console_separator,
omitempty" yaml:"console_separator" toml:"console_separator"` // 输出字段分隔符
}

func InitLogger(cfg *LogConfig) (err error) {
	env := os.Getenv("ENV")
	writeSyncer := getLogWriter(cfg.Filename, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	encoder := getEncoder(env, cfg)
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return
	}
	core := zapcore.NewCore(encoder, writeSyncer, l)
	lg = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(lg)
	return
}

func getEncoder(env string, cfg *LogConfig) zapcore.Encoder {
	var encoderConfig zapcore.EncoderConfig
	if env == "dev" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
	}

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	if cfg.Color {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	if len(cfg.ConsoleSeparator) == 0 {
		encoderConfig.ConsoleSeparator = "\t"
	} else {
		encoderConfig.ConsoleSeparator = cfg.ConsoleSeparator
	}

	//encoderConfig.EncodeCaller = zapcore.FullCallerEncoder      	//显示完整文件路径

	if cfg.Encoder == "text" {
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		lg.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					lg.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					response.Json(c, "trace", "http.5xx", err)
					c.Abort()
					return
				}

				if stack {
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

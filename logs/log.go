// Package logs use zap as logger for its high performence
package logs

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
)

func encodeTimeLayout(t time.Time, layout string, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	encodeTimeLayout(t, "2006-01-02 15:04:05.000", enc)
}

// GetLogger ...
func GetLogger() *zap.SugaredLogger {
	if logger == nil {
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			// EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
			EncodeLevel: zapcore.CapitalColorLevelEncoder,
			// EncodeTime:     zapcore.ISO8601TimeEncoder,       // ISO8601 UTC 时间格式
			EncodeTime:     customTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			// EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
			EncodeCaller: zapcore.ShortCallerEncoder,
		}

		config := zap.Config{
			Level: zap.NewAtomicLevelAt(zap.DebugLevel),
			// Development: true, // 开发模式，堆栈跟踪
			// Encoding:    "json", // 输出格式 console 或 json
			Encoding:         "console",
			EncoderConfig:    encoderConfig,      // 编码器配置
			OutputPaths:      []string{"stdout"}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
			ErrorOutputPaths: []string{"stderr"},
		}

		_logger, _ := config.Build()
		logger = _logger.Sugar()
	}
	return logger
}

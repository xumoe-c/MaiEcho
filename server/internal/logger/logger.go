package logger

import (
	"os"
	"path/filepath"

	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log     *zap.Logger
	Sugar   *zap.SugaredLogger
	LLMLog  *zap.Logger // LLM对话专用日志
	lastLog string
	logMu   sync.RWMutex
)

func init() {
	// 初始化一个默认的 Logger，防止在 Init 被调用前使用导致 panic
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()
	Log = logger
	Sugar = Log.Sugar()
}

func GetLastLog() string {
	logMu.RLock()
	defer logMu.RUnlock()
	return lastLog
}

type Config struct {
	Level      string `mapstructure:"level"`
	OutputPath string `mapstructure:"output_path"`
	Encoding   string `mapstructure:"encoding"` // "json" or "console"
	LLMLogPath string `mapstructure:"llm_log_path"`
}

func Init(cfg Config) error {
	// 新建日志目录
	if cfg.OutputPath != "stdout" && cfg.OutputPath != "stderr" {
		logDir := filepath.Dir(cfg.OutputPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 写入多个输出
	var cores []zapcore.Core

	// 控制台 Core（高优先级日志或根据配置输出所有日志）
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)
	cores = append(cores, consoleCore)

	// 文件 Core
	if cfg.OutputPath != "" && cfg.OutputPath != "stdout" {
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(file),
			level,
		)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)

	// 添加 Hook 以捕获最后一条日志
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel), zap.Hooks(func(entry zapcore.Entry) error {
		logMu.Lock()
		lastLog = "[" + entry.Level.CapitalString() + "] " + entry.Message
		logMu.Unlock()
		return nil
	}))

	Sugar = Log.Sugar()

	// 初始化 LLM 专用日志记录器
	if cfg.LLMLogPath != "" && cfg.LLMLogPath != "stdout" {
		llmLogDir := filepath.Dir(cfg.LLMLogPath)
		if err := os.MkdirAll(llmLogDir, 0755); err != nil {
			return err
		}

		llmFile, err := os.OpenFile(cfg.LLMLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		llmEncoderConfig := zap.NewProductionEncoderConfig()
		llmEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		llmEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		llmEncoder := zapcore.NewJSONEncoder(llmEncoderConfig)
		llmCore := zapcore.NewCore(
			llmEncoder,
			zapcore.AddSync(llmFile),
			zapcore.InfoLevel,
		)
		LLMLog = zap.New(llmCore, zap.AddCaller())
	}

	return nil
}

func Sync() {
	_ = Log.Sync()
}

func Info(msg string, keysAndValues ...interface{}) {
	Sugar.Infow(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	Sugar.Errorw(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...interface{}) {
	Sugar.Debugw(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	Sugar.Warnw(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	Sugar.Fatalw(msg, keysAndValues...)
}

// LogLLMConversation 记录LLM对话内容到专用文件
func LogLLMConversation(model string, systemPrompt string, userPrompt string, response string, err error) {
	if LLMLog == nil {
		return
	}

	if err != nil {
		LLMLog.Error("LLM对话失败",
			zap.String("model", model),
			zap.String("systemPrompt", systemPrompt),
			zap.String("userPrompt", userPrompt),
			zap.Error(err),
		)
	} else {
		LLMLog.Info("LLM对话成功",
			zap.String("model", model),
			zap.String("systemPrompt", systemPrompt),
			zap.String("userPrompt", userPrompt),
			zap.String("response", response),
		)
	}
}

// Gin Middleware Logger
func GinLogger() func(c interface{}) {
	// TODO:这是一个占位符。在实际实现中，我们会返回一个 gin.HandlerFunc
	// 但为了避免在这里导入 gin 并创建循环依赖，或者只是为了保持简单，
	// 我们将在路由包中使用此日志记录器实现中间件。
	return nil
}

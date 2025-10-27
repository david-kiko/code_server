package logger

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// Field 创建日志字段
func Field(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// String 创建字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 创建64位整数字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration 创建时间间隔字段
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Error 创建错误字段
func Error(err error) Field {
	if err != nil {
		return Field{Key: "error", Value: err.Error()}
	}
	return Field{Key: "error", Value: "nil"}
}

// Any 创建任意类型字段
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// JSONLogger JSON格式日志记录器
type JSONLogger struct {
	logger *log.Logger
	fields []Field
}

// NewJSONLogger 创建JSON日志记录器
func NewJSONLogger() Logger {
	flags := log.LstdFlags | log.Lshortfile
	return &JSONLogger{
		logger: log.New(os.Stdout, "", flags),
	}
}

// Info 记录信息日志
func (l *JSONLogger) Info(msg string, fields ...Field) {
	l.log("INFO", msg, fields...)
}

// Warn 记录警告日志
func (l *JSONLogger) Warn(msg string, fields ...Field) {
	l.log("WARN", msg, fields...)
}

// Error 记录错误日志
func (l *JSONLogger) Error(msg string, fields ...Field) {
	l.log("ERROR", msg, fields...)
}

// Debug 记录调试日志
func (l *JSONLogger) Debug(msg string, fields ...Field) {
	l.log("DEBUG", msg, fields...)
}

// Fatal 记录致命错误日志
func (l *JSONLogger) Fatal(msg string, fields ...Field) {
	l.log("FATAL", msg, fields...)
	os.Exit(1)
}

// With 添加字段
func (l *JSONLogger) With(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)
	return &JSONLogger{
		logger: l.logger,
		fields: newFields,
	}
}

// WithContext 添加上下文信息
func (l *JSONLogger) WithContext(ctx context.Context) Logger {
	fields := l.fields

	// 添加请求ID
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, Field{"request_id", requestID})
	}

	// 添加用户ID
	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, Field{"user_id", userID})
	}

	// 添加用户名
	if username := ctx.Value("username"); username != nil {
		fields = append(fields, Field{"username", username})
	}

	return &JSONLogger{
		logger: l.logger,
		fields: fields,
	}
}

// log 记录日志
func (l *JSONLogger) log(level, msg string, fields ...Field) {
	timestamp := time.Now().UTC()

	// 合并字段
	allFields := make([]Field, len(l.fields)+len(fields)+2)
	copy(allFields, l.fields)
	copy(allFields[len(l.fields):], fields)
	allFields = append(allFields, Field{"level", level}, Field{"timestamp", timestamp})
	allFields = append(allFields, Field{"message", msg})

	// 格式化为JSON格式
	jsonMsg := formatJSON(allFields)
	l.logger.Println(jsonMsg)
}

// formatJSON 格式化JSON日志
func formatJSON(fields []Field) string {
	msg := "{"
	for i, field := range fields {
		if i > 0 {
			msg += ", "
		}
		msg += fmt.Sprintf(`"%s": %v`, field.Key, formatValue(field.Value))
	}
	msg += "}"
	return msg
}

// formatValue 格式化值
func formatValue(value interface{}) interface{} {
	switch v := value.(type) {
	case time.Time:
		return v.UTC().Format(time.RFC3339)
	case error:
		if v != nil {
			return v.Error()
		}
		return nil
	default:
		return v
	}
}

// GlobalLogger 全局日志记录器
var GlobalLogger Logger = NewJSONLogger()

// SetGlobalLogger 设置全局日志记录器
func SetGlobalLogger(logger Logger) {
	GlobalLogger = logger
}

// GinLogger Gin中间件日志记录器
type GinLogger struct {
	logger Logger
}

// NewGinLogger 创建Gin日志记录器
func NewGinLogger(logger Logger) *GinLogger {
	return &GinLogger{logger: logger}
}

// GinMiddleware Gin日志中间件
func (g *GinLogger) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
	// 记录请求开始
		start := time.Now()
	rid := c.GetString("rid")
	remoteAddr := c.ClientIP()
	method := c.Request.Method
	uri := c.Request.URL.String()
	protocol := c.Request.Proto

	// 创建上下文日志记录器
	ctxLogger := g.logger.With(
		Any("request_id", rid),
		Any("method", method),
		Any("uri", uri),
		Any("remote_addr", remoteAddr),
		Any("protocol", protocol),
		String("user_agent", c.Request.UserAgent()),
		WithContext(c.Request.Context()),
	)

	// 处理请求
	c.Next()

	// 记录请求完成
	latency := time.Since(start)
	statusCode := c.Writer.Status()

	// 根据状态码选择日志级别
	if statusCode >= 500 {
		ctxLogger.Error("Request completed with server error",
			Int("status_code", statusCode),
			Duration("latency", latency),
		)
	} else if statusCode >= 400 {
		ctxLogger.Warn("Request completed with client error",
			Int("status_code", statusCode),
			Duration("latency", latency),
		)
	} else {
		ctxLogger.Info("Request completed successfully",
			Int("status_code", statusCode),
			Duration("latency", latency),
		)
	}
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
	rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = generateRequestID()
		}
		c.Set("rid", rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 简单的请求ID生成器
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%x", timestamp)
}

// ErrorLogger 错误日志记录器
type ErrorLogger struct {
	logger Logger
}

// NewErrorLogger 创建错误日志记录器
func NewErrorLogger(logger Logger) *ErrorLogger {
	return &ErrorLogger{logger: logger}
}

// LogError 记录错误
func (e *ErrorLogger) LogError(err error, msg string, fields ...Field) {
	logFields := append([]Field{Error(err)}, fields...)
	e.logger.Error(msg, logFields...)
}

// LogPanic 记录panic
func (e *ErrorLogger) LogPanic(recovered interface{}, msg string, fields ...Field) {
	logFields := append([]Field{Any("panic", recovered)}, fields...)
	e.logger.Fatal(msg, logFields...)
}

// RecoveryMiddleware 恢复中间件
func (e *ErrorLogger) RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		e.LogPanic(recovered, "Panic recovered",
			String("method", c.Request.Method),
			String("path", c.Request.URL.Path),
			String("remote_addr", c.ClientIP()),
		)
		c.AbortWithStatus(500)
	})
}
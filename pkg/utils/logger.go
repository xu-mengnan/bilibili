package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Level 日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var (
	// Logger 定义日志记录器
	Logger *log.Logger
	// logLevel 当前日志级别
	logLevel = INFO
)

func init() {
	Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
}

// SetLogLevel 设置日志级别
func SetLogLevel(level Level) {
	logLevel = level
}

// levelToString 将日志级别转换为字符串
func levelToString(level Level) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// formatFields 格式化字段为字符串
func formatFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	data, err := json.Marshal(fields)
	if err != nil {
		return fmt.Sprintf("fields_error: %v", err)
	}

	return string(data)
}

// LogWithFields 记录带字段的日志
func LogWithFields(level Level, fields map[string]interface{}, message string) {
	if level < logLevel {
		return
	}

	fieldsStr := formatFields(fields)
	if fieldsStr != "" {
		Logger.Printf("[%s] %s %s\n", levelToString(level), message, fieldsStr)
	} else {
		Logger.Printf("[%s] %s\n", levelToString(level), message)
	}
}

// LogDebug 记录DEBUG级别日志
func LogDebug(message string) {
	if DEBUG < logLevel {
		return
	}
	Logger.Printf("[DEBUG] %s\n", message)
}

// LogDebugFields 记录DEBUG级别日志（带字段）
func LogDebugFields(fields map[string]interface{}, message string) {
	LogWithFields(DEBUG, fields, message)
}

// LogInfo 记录INFO级别日志
func LogInfo(message string) {
	Logger.Printf("[INFO] %s\n", message)
}

// LogInfoFields 记录INFO级别日志（带字段）
func LogInfoFields(fields map[string]interface{}, message string) {
	LogWithFields(INFO, fields, message)
}

// LogWarn 记录WARN级别日志
func LogWarn(message string) {
	Logger.Printf("[WARN] %s\n", message)
}

// LogWarnFields 记录WARN级别日志（带字段）
func LogWarnFields(fields map[string]interface{}, message string) {
	LogWithFields(WARN, fields, message)
}

// LogError 记录ERROR级别日志
func LogError(message string) {
	Logger.Printf("[ERROR] %s\n", message)
}

// LogErrorFields 记录ERROR级别日志（带字段）
func LogErrorFields(err error, fields map[string]interface{}, message string) {
	if err != nil {
		if fields == nil {
			fields = make(map[string]interface{})
		}
		fields["error"] = err.Error()
	}
	LogWithFields(ERROR, fields, message)
}

// LogRequest 记录HTTP请求日志
func LogRequest(method, path string, statusCode int, duration time.Duration) {
	fields := map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
	}

	if statusCode >= 500 {
		LogErrorFields(nil, fields, "HTTP request")
	} else if statusCode >= 400 {
		LogWarnFields(fields, "HTTP request")
	} else {
		LogInfoFields(fields, "HTTP request")
	}
}

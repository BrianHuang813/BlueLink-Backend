package logger

import (
	"log"
	"os"
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
	LevelFatal = "FATAL"
)

var (
	logLevel = LevelInfo
	logger   = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
)

func SetLevel(level string) {
	logLevel = level
}

func Debug(format string, v ...interface{}) {
	if shouldLog(LevelDebug) {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if shouldLog(LevelInfo) {
		logger.Printf("[INFO] "+format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if shouldLog(LevelWarn) {
		logger.Printf("[WARN] "+format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if shouldLog(LevelError) {
		logger.Printf("[ERROR] "+format, v...)
	}
}

func Fatal(format string, v ...interface{}) {
	logger.Fatalf("[FATAL] "+format, v...)
}

// ErrorWithContext 記錄帶有上下文資訊的錯誤
func ErrorWithContext(requestID, path, method, message string, err error) {
	if shouldLog(LevelError) {
		if err != nil {
			logger.Printf("[ERROR] RequestID=%s Path=%s Method=%s Message=%s Error=%v",
				requestID, path, method, message, err)
		} else {
			logger.Printf("[ERROR] RequestID=%s Path=%s Method=%s Message=%s",
				requestID, path, method, message)
		}
	}
}

// WarnWithContext 記錄帶有上下文資訊的警告
func WarnWithContext(requestID, path, method, message string) {
	if shouldLog(LevelWarn) {
		logger.Printf("[WARN] RequestID=%s Path=%s Method=%s Message=%s",
			requestID, path, method, message)
	}
}

// InfoWithContext 記錄帶有上下文資訊的訊息
func InfoWithContext(requestID, path, method, message string) {
	if shouldLog(LevelInfo) {
		logger.Printf("[INFO] RequestID=%s Path=%s Method=%s Message=%s",
			requestID, path, method, message)
	}
}

// DebugWithContext 記錄帶有上下文資訊的除錯訊息
func DebugWithContext(requestID, path, method, message string) {
	if shouldLog(LevelDebug) {
		logger.Printf("[DEBUG] RequestID=%s Path=%s Method=%s Message=%s",
			requestID, path, method, message)
	}
}

func shouldLog(level string) bool {
	levels := map[string]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
	}
	currentLevel := levels[logLevel]
	targetLevel := levels[level]
	return targetLevel >= currentLevel
}

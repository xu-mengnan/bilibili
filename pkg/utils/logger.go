package utils

import (
	"log"
	"os"
)

// Logger 定义日志记录器
var Logger *log.Logger

func init() {
	// 初始化日志记录器
	Logger = log.New(os.Stdout, "BILIBILI: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// LogInfo 记录INFO级别日志
func LogInfo(message string) {
	Logger.Println("INFO:", message)
}

// LogError 记录ERROR级别日志
func LogError(message string) {
	Logger.Println("ERROR:", message)
}
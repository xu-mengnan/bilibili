package main

import (
	"fmt"
	"os"

	"bilibili/api"
	"bilibili/pkg/utils"
)

func main() {
	fmt.Println("Hello, Bilibili!")
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 设置路由
	router := api.SetupRoutes()

	// 启动HTTP服务器
	port := ":8080"
	utils.LogInfo("Starting server on port " + port)
	err := router.Run(port)
	if err != nil {
		utils.LogError("Server failed to start: " + err.Error())
		return err
	}

	return nil
}
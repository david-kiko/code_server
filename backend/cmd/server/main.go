package main

import (
	"log"
	"os"

	"container-platform-backend/internal/api"
)

func main() {
	log.Println("Starting Container Platform Backend Server...")

	// 创建路由器
	router := api.NewRouter()

	// 设置路由
	router.Setup()

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.GetEngine().Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
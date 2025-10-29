package api

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Router 路由器
type Router struct {
	engine        *gin.Engine
	k8sController *K8sController
}

// NewRouter 创建路由器
func NewRouter() *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	return &Router{
		engine:        engine,
		k8sController: NewK8sController(),
	}
}

// Setup 设置路由和中间件
func (r *Router) Setup() {
	// 全局中间件
	r.setupGlobalMiddleware()

	// API路由组
	api := r.engine.Group("/api")
	r.setupRoutes(api)
}

// GetEngine 获取Gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// setupGlobalMiddleware 设置全局中间件
func (r *Router) setupGlobalMiddleware() {
	// 日志中间件
	r.engine.Use(gin.Logger())

	// 错误处理中间件
	r.engine.Use(gin.Recovery())

	// CORS中间件 - 完全开放
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Requested-With",
	}
	r.engine.Use(cors.New(corsConfig))
}

// setupRoutes 设置API路由
func (r *Router) setupRoutes(rg *gin.RouterGroup) {
	// 健康检查
	rg.GET("/health", r.healthCheck)

	// K8s 容器管理 API
	k8s := rg.Group("/k8s")
	{
		// 容器管理
		k8s.GET("/containers", r.k8sController.GetContainers)
		k8s.POST("/containers", r.k8sController.CreateContainer)
		k8s.POST("/containers/:namespace/:podName/start", r.k8sController.StartContainer)
		k8s.POST("/containers/:namespace/:podName/stop", r.k8sController.StopContainer)
		k8s.POST("/containers/:namespace/:podName/restart", r.k8sController.RestartContainer)
		k8s.DELETE("/containers/:namespace/:podName", r.k8sController.DeleteContainer)

		// 连接测试
		k8s.POST("/test-connection", r.k8sController.TestConnection)
	}
}

// healthCheck 健康检查
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"status":    "healthy",
			"timestamp": "2025-10-27T10:30:00Z",
			"version":   "1.0.0",
			"service":   "container-platform-backend",
		},
	})
}
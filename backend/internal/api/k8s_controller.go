package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"container-platform-backend/internal/model"
	"container-platform-backend/internal/services"
)

type K8sController struct {
	k8sService *services.K8sService
}

func NewK8sController() *K8sController {
	return &K8sController{
		k8sService: services.NewK8sService(),
	}
}

// GetContainers 获取容器列表
// @Summary 获取容器列表
// @Description 获取指定命名空间下的容器列表
// @Tags k8s
// @Accept json
// @Produce json
// @Param namespace query string false "命名空间" default(default)
// @Success 200 {object} APIResponse{data=[]services.ContainerInfo}
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/k8s/containers [get]
func (c *K8sController) GetContainers(ctx *gin.Context) {
	namespace := ctx.DefaultQuery("namespace", "default")

	// 这里需要从请求中获取连接信息，或者从用户配置中获取
	// 暂时使用一个示例连接
	connection := &model.K8sConnection{
		Name:       "default-cluster",
		Endpoint:   "https://kubernetes.default.svc:443",
		ConfigType: "kubeconfig",
		Namespace:  namespace,
	}

	// 连接到集群
	if err := c.k8sService.ConnectToCluster(connection); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to connect to Kubernetes cluster", err)
		return
	}

	// 获取容器列表
	containers, err := c.k8sService.ListContainers(namespace)
	if err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list containers", err)
		return
	}

	SuccessResponse(ctx, "Containers retrieved successfully", containers)
}

// CreateContainer 创建容器
// @Summary 创建容器
// @Description 创建一个新的容器
// @Tags k8s
// @Accept json
// @Produce json
// @Param container body services.CreateContainerRequest true "容器信息"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/k8s/containers [post]
func (c *K8sController) CreateContainer(ctx *gin.Context) {
	var req services.CreateContainerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 这里需要从请求中获取连接信息，或者从用户配置中获取
	connection := &model.K8sConnection{
		Name:       "default-cluster",
		Endpoint:   "https://kubernetes.default.svc:443",
		ConfigType: "kubeconfig",
		Namespace:  req.Namespace,
	}

	// 连接到集群
	if err := c.k8sService.ConnectToCluster(connection); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to connect to Kubernetes cluster", err)
		return
	}

	// 创建容器
	if err := c.k8sService.CreateContainer(&req); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create container", err)
		return
	}

	SuccessResponse(ctx, "Container created successfully", nil)
}

// StartContainer 启动容器
// @Summary 启动容器
// @Description 启动指定的容器
// @Tags k8s
// @Accept json
// @Produce json
// @Param namespace path string true "命名空间"
// @Param podName path string true "Pod 名称"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/k8s/containers/{namespace}/{podName}/start [post]
func (c *K8sController) StartContainer(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	podName := ctx.Param("podName")

	if namespace == "" || podName == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "Namespace and pod name are required", nil)
		return
	}

	// 连接到集群
	connection := &model.K8sConnection{
		Name:       "default-cluster",
		Endpoint:   "https://kubernetes.default.svc:443",
		ConfigType: "kubeconfig",
		Namespace:  namespace,
	}

	if err := c.k8sService.ConnectToCluster(connection); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to connect to Kubernetes cluster", err)
		return
	}

	// 启动容器
	if err := c.k8sService.StartContainer(namespace, podName); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to start container", err)
		return
	}

	SuccessResponse(ctx, "Container started successfully", nil)
}

// StopContainer 停止容器
// @Summary 停止容器
// @Description 停止指定的容器
// @Tags k8s
// @Accept json
// @Produce json
// @Param namespace path string true "命名空间"
// @Param podName path string true "Pod 名称"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/k8s/containers/{namespace}/{podName}/stop [post]
func (c *K8sController) StopContainer(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	podName := ctx.Param("podName")

	if namespace == "" || podName == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "Namespace and pod name are required", nil)
		return
	}

	// 连接到集群
	connection := &model.K8sConnection{
		Name:       "default-cluster",
		Endpoint:   "https://kubernetes.default.svc:443",
		ConfigType: "kubeconfig",
		Namespace:  namespace,
	}

	if err := c.k8sService.ConnectToCluster(connection); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to connect to Kubernetes cluster", err)
		return
	}

	// 停止容器
	if err := c.k8sService.StopContainer(namespace, podName); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to stop container", err)
		return
	}

	SuccessResponse(ctx, "Container stopped successfully", nil)
}

// RestartContainer 重启容器
// @Summary 重启容器
// @Description 重启指定的容器
// @Tags k8s
// @Accept json
// @Produce json
// @Param namespace path string true "命名空间"
// @Param podName path string true "Pod 名称"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/k8s/containers/{namespace}/{podName}/restart [post]
func (c *K8sController) RestartContainer(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	podName := ctx.Param("podName")

	if namespace == "" || podName == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "Namespace and pod name are required", nil)
		return
	}

	// 连接到集群
	connection := &model.K8sConnection{
		Name:       "default-cluster",
		Endpoint:   "https://kubernetes.default.svc:443",
		ConfigType: "kubeconfig",
		Namespace:  namespace,
	}

	if err := c.k8sService.ConnectToCluster(connection); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to connect to Kubernetes cluster", err)
		return
	}

	// 重启容器
	if err := c.k8sService.RestartContainer(namespace, podName); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to restart container", err)
		return
	}

	SuccessResponse(ctx, "Container restarted successfully", nil)
}

// DeleteContainer 删除容器
// @Summary 删除容器
// @Description 删除指定的容器
// @Tags k8s
// @Accept json
// @Produce json
// @Param namespace path string true "命名空间"
// @Param podName path string true "Pod 名称"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/k8s/containers/{namespace}/{podName} [delete]
func (c *K8sController) DeleteContainer(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	podName := ctx.Param("podName")

	if namespace == "" || podName == "" {
		ErrorResponse(ctx, http.StatusBadRequest, "Namespace and pod name are required", nil)
		return
	}

	// 连接到集群
	connection := &model.K8sConnection{
		Name:       "default-cluster",
		Endpoint:   "https://kubernetes.default.svc:443",
		ConfigType: "kubeconfig",
		Namespace:  namespace,
	}

	if err := c.k8sService.ConnectToCluster(connection); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to connect to Kubernetes cluster", err)
		return
	}

	// 删除容器
	if err := c.k8sService.DeleteContainer(namespace, podName); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete container", err)
		return
	}

	SuccessResponse(ctx, "Container deleted successfully", nil)
}

// TestConnection 测试 K8s 连接
// @Summary 测试 K8s 连接
// @Description 测试 Kubernetes 集群连接
// @Tags k8s
// @Accept json
// @Produce json
// @Param connection body model.K8sConnection true "连接配置"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/k8s/test-connection [post]
func (c *K8sController) TestConnection(ctx *gin.Context) {
	var connection model.K8sConnection
	if err := ctx.ShouldBindJSON(&connection); err != nil {
		ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// 测试连接
	if err := c.k8sService.TestConnection(&connection); err != nil {
		ErrorResponse(ctx, http.StatusInternalServerError, "Failed to connect to Kubernetes cluster", err)
		return
	}

	SuccessResponse(ctx, "Connection test successful", nil)
}
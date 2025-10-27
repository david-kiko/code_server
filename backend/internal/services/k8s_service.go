package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"container-platform-backend/internal/model"
)

type K8sService struct {
	clientSet *kubernetes.Clientset
	config    *rest.Config
}

type ContainerInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Image        string            `json:"image"`
	Status       string            `json:"status"`
	PodName      string            `json:"podName"`
	RestartCount int32             `json:"restartCount"`
	Age          string            `json:"age"`
	Node         string            `json:"node,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	ContainerID  string            `json:"containerId,omitempty"`
}

func NewK8sService() *K8sService {
	return &K8sService{}
}

// ConnectToCluster 连接到 Kubernetes 集群
func (s *K8sService) ConnectToCluster(connection *model.K8sConnection) error {
	var config *rest.Config
	var err error

	if connection.ConfigType == "kubeconfig" {
		// 使用 kubeconfig 文件连接
		clusterConfig := api.NewConfig()
		if err := json.Unmarshal([]byte(connection.Config), clusterConfig); err != nil {
			return fmt.Errorf("failed to parse kubeconfig: %w", err)
		}

		// 创建临时的 kubeconfig 文件
		tmpFile := "/tmp/kubeconfig-" + fmt.Sprintf("%d", connection.ID)
		if err := clientcmd.WriteToFile(*clusterConfig, tmpFile); err != nil {
			return fmt.Errorf("failed to write kubeconfig file: %w", err)
		}

		config, err = clientcmd.BuildConfigFromFlags("", tmpFile)
		if err != nil {
			return fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}
	} else {
		// 使用 token 连接
		config = &rest.Config{
			Host:        connection.Endpoint,
			BearerToken: connection.Token,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true, // 生产环境应该配置 proper TLS
			},
		}
	}

	// 设置默认命名空间
	if connection.Namespace != "" {
		config.Impersonate = rest.ImpersonationConfig{
			UserName: fmt.Sprintf("system:serviceaccount:%s:default", connection.Namespace),
		}
	}

	// 创建 clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	s.clientSet = clientSet
	s.config = config

	// 测试连接
	_, err = clientSet.ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	log.Printf("Successfully connected to Kubernetes cluster: %s", connection.Name)
	return nil
}

// TestConnection 测试 K8s 连接
func (s *K8sService) TestConnection(connection *model.K8sConnection) error {
	return s.ConnectToCluster(connection)
}

// ListContainers 获取容器列表
func (s *K8sService) ListContainers(namespace string) ([]ContainerInfo, error) {
	if s.clientSet == nil {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	pods, err := s.clientSet.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var containers []ContainerInfo
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			// 获取容器状态
			status := "Unknown"
			restartCount := int32(0)
			containerID := ""

			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Name == container.Name {
					if containerStatus.State.Running != nil {
						status = "Running"
					} else if containerStatus.State.Waiting != nil {
						status = "Pending"
					} else if containerStatus.State.Terminated != nil {
						if containerStatus.State.Terminated.ExitCode == 0 {
							status = "Succeeded"
						} else {
							status = "Failed"
						}
					}
					restartCount = containerStatus.RestartCount
					containerID = containerStatus.ContainerID
					break
				}
			}

			// 计算容器年龄
			age := calculateAge(pod.CreationTimestamp.Time)

			containerInfo := ContainerInfo{
				Name:         container.Name,
				Namespace:    pod.Namespace,
				Image:        container.Image,
				Status:       status,
				PodName:      pod.Name,
				RestartCount: restartCount,
				Age:          age,
				Node:         pod.Spec.NodeName,
				Labels:       pod.Labels,
				ContainerID:  containerID,
			}

			containers = append(containers, containerInfo)
		}
	}

	return containers, nil
}

// CreateContainer 创建容器 (通过创建 Pod)
func (s *K8sService) CreateContainer(req *CreateContainerRequest) error {
	if s.clientSet == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	// 解析环境变量
	var envVars []corev1.EnvVar
	if req.Env != "" {
		envMap := parseEnvVars(req.Env)
		for k, v := range envMap {
			envVars = append(envVars, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	// 解析端口映射
	var ports []corev1.ContainerPort
	if req.Ports != "" {
		portMap := parsePortMappings(req.Ports)
		for _, port := range portMap {
			ports = append(ports, corev1.ContainerPort{
				ContainerPort: port,
			})
		}
	}

	// 解析资源限制
	var resources corev1.ResourceRequirements
	if req.Resources != "" {
		resources = parseResources(req.Resources)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name + "-pod",
			Namespace: req.Namespace,
			Labels: map[string]string{
				"app":      req.Name,
				"managed": "container-platform",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:      req.Name,
					Image:     req.Image,
					Env:       envVars,
					Ports:     ports,
					Resources: resources,
				},
			},
			RestartPolicy: corev1.RestartPolicyAlways,
		},
	}

	_, err := s.clientSet.CoreV1().Pods(req.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	log.Printf("Successfully created pod: %s", pod.Name)
	return nil
}

// StartContainer 启动容器
func (s *K8sService) StartContainer(namespace, podName string) error {
	// 在 Kubernetes 中，Pod 的启动通常是通过创建或删除 Pod 来实现的
	// 这里可以通过删除失败的 Pod 来让它重新启动
	return s.RestartContainer(namespace, podName)
}

// StopContainer 停止容器
func (s *K8sService) StopContainer(namespace, podName string) error {
	if s.clientSet == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	// 删除 Pod
	deletePolicy := metav1.DeletePropagationForeground
	err := s.clientSet.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	log.Printf("Successfully stopped pod: %s", podName)
	return nil
}

// RestartContainer 重启容器
func (s *K8sService) RestartContainer(namespace, podName string) error {
	if s.clientSet == nil {
		return fmt.Errorf("kubernetes client not initialized")
	}

	// 删除现有的 Pod，让 Deployment 或 ReplicaSet 重新创建
	err := s.clientSet.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod for restart: %w", err)
	}

	log.Printf("Successfully restarted pod: %s", podName)
	return nil
}

// DeleteContainer 删除容器
func (s *K8sService) DeleteContainer(namespace, podName string) error {
	return s.StopContainer(namespace, podName)
}

type CreateContainerRequest struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`
	Command   string `json:"command"`
	Ports     string `json:"ports"`
	Env       string `json:"env"`
	Resources string `json:"resources"`
}

// 辅助函数
func calculateAge(creationTime time.Time) string {
	duration := time.Since(creationTime)

	if duration.Hours() < 1 {
		return fmt.Sprintf("%.0fm", duration.Minutes())
	} else if duration.Hours() < 24 {
		return fmt.Sprintf("%.0fh", duration.Hours())
	} else {
		return fmt.Sprintf("%.0fd", duration.Hours()/24)
	}
}

func parseEnvVars(envStr string) map[string]string {
	envMap := make(map[string]string)
	// 简单解析，实际应该更robust
	lines := strings.Split(envStr, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return envMap
}

func parsePortMappings(portsStr string) []int32 {
	var ports []int32
	// 简单解析端口映射
	parts := strings.Split(portsStr, ",")
	for _, part := range parts {
		portStr := strings.TrimSpace(part)
		if portStr != "" {
			if port, err := strconv.Atoi(portStr); err == nil {
				ports = append(ports, int32(port))
			}
		}
	}
	return ports
}

func parseResources(resourcesStr string) corev1.ResourceRequirements {
	requirements := corev1.ResourceRequirements{
		Requests: make(corev1.ResourceList),
		Limits:   make(corev1.ResourceList),
	}

	// 简单解析资源限制
	parts := strings.Split(resourcesStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "cpu:") {
			cpuValue := strings.TrimSpace(strings.TrimPrefix(part, "cpu:"))
			requirements.Requests[corev1.ResourceCPU] = resource.MustParse(cpuValue)
		} else if strings.Contains(part, "memory:") {
			memValue := strings.TrimSpace(strings.TrimPrefix(part, "memory:"))
			requirements.Requests[corev1.ResourceMemory] = resource.MustParse(memValue)
		}
	}

	return requirements
}
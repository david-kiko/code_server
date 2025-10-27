package k8s

import (
	"container-platform-backend/internal/model"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client Kubernetes客户端
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
	namespace string
}

// NewClient 创建Kubernetes客户端
func NewClient(configPath, namespace string) (*Client, error) {
	var config *rest.Config
	var err error

	// 尝试使用提供的配置路径
	if configPath != "" {
		config, err = clientcmd.BuildConfigFromFlags(configPath, "")
		if err != nil {
			log.Printf("Warning: 无法从配置路径 %s 加载配置: %v", configPath, err)
		}
	}

	// 如果没有配置或配置失败，尝试使用集群内配置
	if config == nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Printf("Warning: 无法使用集群内配置: %v", err)
		}
	}

	// 如果仍然没有配置，尝试使用默认的kubeconfig
	if config == nil {
		home := homedir.HomeDir()
		kubeconfig := filepath.Join(home, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("无法加载Kubernetes配置: %w", err)
		}
	}

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建Kubernetes客户端失败: %w", err)
	}

	// 设置默认命名空间
	if namespace == "" {
		namespace = "default"
	}

	log.Printf("Kubernetes客户端初始化成功，命名空间: %s", namespace)

	return &Client{
		clientset: clientset,
		config:    config,
		namespace: namespace,
	}, nil
}

// GetNamespace 获取当前命名空间
func (c *Client) GetNamespace() string {
	return c.namespace
}

// ListPods 列出Pod
func (c *Client) ListPods(ctx context.Context, labelSelector string) ([]corev1.Pod, error) {
	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("列出Pod失败: %w", err)
	}

	return pods.Items, nil
}

// GetPod 获取单个Pod
func (c *Client) GetPod(ctx context.Context, name string) (*corev1.Pod, error) {
	pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Pod %s 失败: %w", name, err)
	}

	return pod, nil
}

// CreatePod 创建Pod
func (c *Client) CreatePod(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, error) {
	createdPod, err := c.clientset.CoreV1().Pods(c.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建Pod失败: %w", err)
	}

	return createdPod, nil
}

// DeletePod 删除Pod
func (c *Client) DeletePod(ctx context.Context, name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := c.clientset.CoreV1().Pods(c.namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return fmt.Errorf("删除Pod %s 失败: %w", name, err)
	}

	return nil
}

// UpdatePod 更新Pod
func (c *Client) UpdatePod(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, error) {
	updatedPod, err := c.clientset.CoreV1().Pods(c.namespace).Update(ctx, pod, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新Pod失败: %w", err)
	}

	return updatedPod, nil
}

// GetPodLogs 获取Pod日志
func (c *Client) GetPodLogs(ctx context.Context, podName, containerName string, lines int64) (string, error) {
	logOptions := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &lines,
	}

	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(podName, logOptions)
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("获取Pod日志失败: %w", err)
	}
	defer logs.Close()

	// 读取日志
	buf := make([]byte, 4096)
	var result string
	for {
		n, err := logs.Read(buf)
		if err != nil {
			break
		}
		result += string(buf[:n])
	}

	return result, nil
}

// ListServices 列出Service
func (c *Client) ListServices(ctx context.Context, labelSelector string) ([]corev1.Service, error) {
	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	services, err := c.clientset.CoreV1().Services(c.namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("列出Service失败: %w", err)
	}

	return services.Items, nil
}

// GetService 获取单个Service
func (c *Client) GetService(ctx context.Context, name string) (*corev1.Service, error) {
	service, err := c.clientset.CoreV1().Services(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Service %s 失败: %w", name, err)
	}

	return service, nil
}

// CreateService 创建Service
func (c *Client) CreateService(ctx context.Context, service *corev1.Service) (*corev1.Service, error) {
	createdService, err := c.clientset.CoreV1().Services(c.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建Service失败: %w", err)
	}

	return createdService, nil
}

// DeleteService 删除Service
func (c *Client) DeleteService(ctx context.Context, name string) error {
	err := c.clientset.CoreV1().Services(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除Service %s 失败: %w", name, err)
	}

	return nil
}

// ListConfigMaps 列出ConfigMap
func (c *Client) ListConfigMaps(ctx context.Context, labelSelector string) ([]corev1.ConfigMap, error) {
	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	configMaps, err := c.clientset.CoreV1().ConfigMaps(c.namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("列出ConfigMap失败: %w", err)
	}

	return configMaps.Items, nil
}

// GetConfigMap 获取单个ConfigMap
func (c *Client) GetConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	configMap, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取ConfigMap %s 失败: %w", name, err)
	}

	return configMap, nil
}

// CreateConfigMap 创建ConfigMap
func (c *Client) CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	createdConfigMap, err := c.clientset.CoreV1().ConfigMaps(c.namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建ConfigMap失败: %w", err)
	}

	return createdConfigMap, nil
}

// ListSecrets 列出Secret
func (c *Client) ListSecrets(ctx context.Context, labelSelector string) ([]corev1.Secret, error) {
	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	secrets, err := c.clientset.CoreV1().Secrets(c.namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("列出Secret失败: %w", err)
	}

	return secrets.Items, nil
}

// GetSecret 获取单个Secret
func (c *Client) GetSecret(ctx context.Context, name string) (*corev1.Secret, error) {
	secret, err := c.clientset.CoreV1().Secrets(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Secret %s 失败: %w", name, err)
	}

	return secret, nil
}

// CreateSecret 创建Secret
func (c *Client) CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	createdSecret, err := c.clientset.CoreV1().Secrets(c.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建Secret失败: %w", err)
	}

	return createdSecret, nil
}

// ListPersistentVolumeClaims 列出PVC
func (c *Client) ListPersistentVolumeClaims(ctx context.Context, labelSelector string) ([]corev1.PersistentVolumeClaim, error) {
	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	pvcs, err := c.clientset.CoreV1().PersistentVolumeClaims(c.namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("列出PVC失败: %w", err)
	}

	return pvcs.Items, nil
}

// GetPersistentVolumeClaim 获取单个PVC
func (c *Client) GetPersistentVolumeClaim(ctx context.Context, name string) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := c.clientset.CoreV1().PersistentVolumeClaims(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取PVC %s 失败: %w", name, err)
	}

	return pvc, nil
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	// 检查API服务器连接
	_, err := c.clientset.ServerVersion()
	if err != nil {
		return fmt.Errorf("无法连接到Kubernetes API服务器: %w", err)
	}

	// 检查命名空间是否存在
	_, err = c.clientset.CoreV1().Namespaces().Get(ctx, c.namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("命名空间 %s 不存在: %w", c.namespace, err)
	}

	return nil
}

// GetClusterInfo 获取集群信息
func (c *Client) GetClusterInfo(ctx context.Context) (map[string]interface{}, error) {
	serverVersion, err := c.clientset.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("获取服务器版本失败: %w", err)
	}

	info := map[string]interface{}{
		"version": serverVersion.String(),
		"namespace": c.namespace,
		"platform": serverVersion.Platform,
	}

	return info, nil
}

// WaitForPodReady 等待Pod就绪
func (c *Client) WaitForPodReady(ctx context.Context, podName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("等待Pod %s 就绪超时", podName)
		default:
			pod, err := c.GetPod(ctx, podName)
			if err != nil {
				return fmt.Errorf("获取Pod状态失败: %w", err)
			}

			if pod.Status.Phase == corev1.PodRunning {
				// 检查所有容器是否就绪
				allReady := true
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if !containerStatus.Ready {
						allReady = false
						break
					}
				}

				if allReady {
					return nil
				}
			}

			time.Sleep(2 * time.Second)
		}
	}
}

// ConvertToContainerModel 将Kubernetes Pod转换为内部模型
func ConvertToContainerModel(pod corev1.Pod) *model.Container {
	container := &model.Container{
		Name:       pod.Name,
		K8sName:    pod.Name,
		PodName:    pod.Name,
		Status:     string(pod.Status.Phase),
		Phase:      string(pod.Status.Phase),
		PodIP:      pod.Status.PodIP,
		NodeName:   pod.Spec.NodeName,
		RestartCount: 0,
	}

	// 设置资源限制
	if pod.Spec.Containers != nil && len(pod.Spec.Containers) > 0 {
		c := pod.Spec.Containers[0]
		if c.Resources.Requests != nil {
			if cpu := c.Resources.Requests.Cpu(); !cpu.IsZero() {
				container.CPURequest = cpu.String()
			}
			if memory := c.Resources.Requests.Memory(); !memory.IsZero() {
				container.MemoryRequest = memory.String()
			}
		}
		if c.Resources.Limits != nil {
			if cpu := c.Resources.Limits.Cpu(); !cpu.IsZero() {
				container.CPULimit = cpu.String()
			}
			if memory := c.Resources.Limits.Memory(); !memory.IsZero() {
				container.MemoryLimit = memory.String()
			}
		}
	}

	// 设置时间
	if pod.Status.StartTime != nil {
		startedAt := pod.Status.StartTime.Time
		container.StartedAt = &startedAt
	}

	return container
}
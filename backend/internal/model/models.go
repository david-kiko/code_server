package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// User 用户模型
type User struct {
	BaseModel
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	FullName     string         `gorm:"size:100" json:"fullName"`
	AvatarURL    string         `gorm:"size:500" json:"avatarUrl"`
	Status       string         `gorm:"default:active;size:20" json:"status"`
	LastLoginAt  *time.Time     `json:"lastLoginAt"`
	Roles        []Role         `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	CreatedBy    *uint          `json:"createdBy"`
	UpdatedBy    *uint          `json:"updatedBy"`
}

// Role 角色模型
type Role struct {
	BaseModel
	Name        string   `gorm:"uniqueIndex;size:50;not null" json:"name"`
	DisplayName string   `gorm:"size:100;not null" json:"displayName"`
	Description string   `gorm:"type:text" json:"description"`
	IsSystem    bool     `gorm:"default:false" json:"isSystem"`
	Permissions JSONB    `gorm:"type:jsonb;default:'[]'" json:"permissions"`
	CreatedBy   *uint    `json:"createdBy"`
	UpdatedBy   *uint    `json:"updatedBy"`
}

// UserRole 用户角色关联
type UserRole struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null" json:"userId"`
	RoleID     uint      `gorm:"not null" json:"roleId"`
	AssignedAt time.Time `gorm:"not null" json:"assignedAt"`
	AssignedBy *uint     `json:"assignedBy"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	User       User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role       Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// Namespace 命名空间模型
type Namespace struct {
	BaseModel
	Name         string    `gorm:"uniqueIndex;size:63;not null" json:"name"`
	DisplayName  string    `gorm:"size:100" json:"displayName"`
	Description  string    `gorm:"type:text" json:"description"`
	K8sName      string    `gorm:"size:63;not null" json:"k8sName"`
	ClusterName  string    `gorm:"size:100;not null" json:"clusterName"`
	Status       string    `gorm:"default:active;size:20" json:"status"`
	ResourceQuota JSONB    `gorm:"type:jsonb" json:"resourceQuota"`
	CreatedBy    *uint     `json:"createdBy"`
	UpdatedBy    *uint     `json:"updatedBy"`
	Users        []User    `gorm:"many2many:namespace_permissions;" json:"users,omitempty"`
}

// NamespacePermission 命名空间权限
type NamespacePermission struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"not null" json:"userId"`
	NamespaceID    uint           `gorm:"not null" json:"namespaceId"`
	PermissionLevel string        `gorm:"size:20;not null" json:"permissionLevel"`
	GrantedBy      *uint          `json:"grantedBy"`
	GrantedAt      time.Time      `json:"grantedAt"`
	ExpiresAt      *time.Time     `json:"expiresAt"`
	User           User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Namespace      Namespace      `gorm:"foreignKey:NamespaceID" json:"namespace,omitempty"`
}

// ContainerImage 容器镜像模型
type ContainerImage struct {
	BaseModel
	Name         string    `gorm:"not null" json:"name"`
	Tag          string    `gorm:"size:128;not null" json:"tag"`
	Digest       string    `gorm:"size:128" json:"digest"`
	Repository   string    `gorm:"size:255" json:"repository"`
	SizeBytes    int64     `json:"sizeBytes"`
	Architecture string    `gorm:"size:20" json:"architecture"`
	OS           string    `gorm:"size:20" json:"os"`
	CreatedAt    time.Time `json:"createdAt"`
	PulledAt     time.Time `gorm:"not null" json:"pulledAt"`
	Status       string    `gorm:"default:available;size:20" json:"status"`
}

// Container 容器实例模型
type Container struct {
	BaseModel
	Name            string          `gorm:"not null" json:"name"`
	DisplayName     string          `gorm:"size:253" json:"displayName"`
	NamespaceID     uint            `gorm:"not null" json:"namespaceId"`
	ImageID         uint            `gorm:"not null" json:"imageId"`
	K8sName         string          `gorm:"size:253;not null" json:"k8sName"`
	PodName         string          `gorm:"size:253" json:"podName"`
	DeploymentName  string          `gorm:"size:253" json:"deploymentName"`
	ReplicaSetName  string          `gorm:"size:253" json:"replicaSetName"`
	Status          string          `gorm:"default:pending;size:20" json:"status"`
	Phase           string          `gorm:"size:20" json:"phase"`
	Reason          string          `gorm:"size:100" json:"reason"`
	Message         string          `gorm:"type:text" json:"message"`
	CPURequest      string          `gorm:"size:20" json:"cpuRequest"`
	CPULimit        string          `gorm:"size:20" json:"cpuLimit"`
	MemoryRequest   string          `gorm:"size:20" json:"memoryRequest"`
	MemoryLimit     string          `gorm:"size:20" json:"memoryLimit"`
	RestartCount    int             `gorm:"default:0" json:"restartCount"`
	PodIP           string          `gorm:"size:45" json:"podIp"`
	HostIP          string          `gorm:"size:45" json:"hostIp"`
	NodeName        string          `gorm:"size:63" json:"nodeName"`
	StartedAt       *time.Time      `json:"startedAt"`
	FinishedAt      *time.Time      `json:"finishedAt"`
	CreatedBy       *uint           `json:"createdBy"`
	UpdatedBy       *uint           `json:"updatedBy"`
	Namespace       Namespace      `gorm:"foreignKey:NamespaceID" json:"namespace,omitempty"`
	Image           ContainerImage `gorm:"foreignKey:ImageID" json:"image,omitempty"`
	Configs         []ContainerConfig `gorm:"foreignKey:ContainerID" json:"configs,omitempty"`
	Volumes         []ContainerVolume `gorm:"foreignKey:ContainerID" json:"volumes,omitempty"`
	PortMappings    []PortMapping    `gorm:"foreignKey:ContainerID" json:"portMappings,omitempty"`
	EnvVars         []EnvironmentVariable `gorm:"foreignKey:ContainerID" json:"envVars,omitempty"`
}

// ContainerConfig 容器配置模型
type ContainerConfig struct {
	BaseModel
	ContainerID uint            `gorm:"not null" json:"containerId"`
	ConfigType  string          `gorm:"size:50;not null" json:"configType"`
	ConfigName  string          `gorm:"size:63;not null" json:"configName"`
	ConfigValue JSONB           `gorm:"type:jsonb;not null" json:"configValue"`
	Container   Container      `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
}

// Volume 存储卷模型
type Volume struct {
	BaseModel
	Name         string    `gorm:"not null" json:"name"`
	NamespaceID  uint      `gorm:"not null" json:"namespaceId"`
	Type         string    `gorm:"size:50;not null" json:"type"`
	StorageClass string    `gorm:"size:100" json:"storageClass"`
	Size         string    `gorm:"size:20" json:"size"`
	AccessMode   string    `gorm:"size:20;default:ReadWriteOnce" json:"accessMode"`
	MountPath    string    `gorm:"size:500" json:"mountPath"`
	HostPath     string    `gorm:"size:500" json:"hostPath"`
	K8sName      string    `gorm:"size:253" json:"k8sName"`
	Status       string    `gorm:"default:available;size:20" json:"status"`
	CreatedBy    *uint     `json:"createdBy"`
	Namespace    Namespace `gorm:"foreignKey:NamespaceID" json:"namespace,omitempty"`
}

// ContainerVolume 容器存储卷关联
type ContainerVolume struct {
	BaseModel
	ContainerID uint   `gorm:"not null" json:"containerId"`
	VolumeID    uint   `gorm:"not null" json:"volumeId"`
	MountPath   string `gorm:"size:500;not null" json:"mountPath"`
	ReadOnly    bool   `gorm:"default:false" json:"readOnly"`
	SubPath     string `gorm:"size:500" json:"subPath"`
	Container   Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
	Volume      Volume    `gorm:"foreignKey:VolumeID" json:"volume,omitempty"`
}

// PortMapping 端口映射模型
type PortMapping struct {
	BaseModel
	ContainerID   uint   `gorm:"not null" json:"containerId"`
	Name          string `gorm:"size:100" json:"name"`
	ContainerPort int    `gorm:"not null" json:"containerPort"`
	HostPort      *int   `json:"hostPort"`
	Protocol      string `gorm:"size:10;default:TCP" json:"protocol"`
	ServiceName   string `gorm:"size:253" json:"serviceName"`
	ServicePort   *int   `json:"servicePort"`
	NodePort      *int   `json:"nodePort"`
	Container     Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
}

// EnvironmentVariable 环境变量模型
type EnvironmentVariable struct {
	BaseModel
	ContainerID uint   `gorm:"not null" json:"containerId"`
	Name        string `gorm:"size:253;not null" json:"name"`
	Value       string `gorm:"type:text" json:"value"`
	ValueFrom   string `gorm:"size:50" json:"valueFrom"`
	SourceName  string `gorm:"size:253" json:"sourceName"`
	SourceKey   string `gorm:"size:253" json:"sourceKey"`
	Container   Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
}

// ConfigMap 配置映射模型
type ConfigMap struct {
	BaseModel
	Name        string `gorm:"not null" json:"name"`
	NamespaceID uint   `gorm:"not null" json:"namespaceId"`
	K8sName     string `gorm:"size:253;not null" json:"k8sName"`
	Data        JSONB  `gorm:"type:jsonb;default:'{}'" json:"data"`
	BinaryData  JSONB  `gorm:"type:jsonb;default:'{}'" json:"binaryData"`
	CreatedBy   *uint  `json:"createdBy"`
	UpdatedBy   *uint  `json:"updatedBy"`
	Namespace   Namespace `gorm:"foreignKey:NamespaceID" json:"namespace,omitempty"`
}

// Secret 密钥模型
type Secret struct {
	BaseModel
	Name        string `gorm:"not null" json:"name"`
	NamespaceID uint   `gorm:"not null" json:"namespaceId"`
	K8sName     string `gorm:"size:253;not null" json:"k8sName"`
	Type        string `gorm:"size:100;default:Opaque" json:"type"`
	Data        JSONB  `gorm:"type:jsonb;default:'{}'" json:"data"`
	CreatedBy   *uint  `json:"createdBy"`
	UpdatedBy   *uint  `json:"updatedBy"`
	Namespace   Namespace `gorm:"foreignKey:NamespaceID" json:"namespace,omitempty"`
}

// Service 服务模型
type Service struct {
	BaseModel
	Name         string `gorm:"not null" json:"name"`
	NamespaceID  uint   `gorm:"not null" json:"namespaceId"`
	K8sName     string `gorm:"size:253;not null" json:"k8sName"`
	ServiceType  string `gorm:"size:50;default:ClusterIP" json:"serviceType"`
	ClusterIP    string `gorm:"size:45" json:"clusterIp"`
	ExternalIP   string `gorm:"size:45" json:"externalIp"`
	Ports        JSONB  `gorm:"type:jsonb;not null" json:"ports"`
	Selector     JSONB  `gorm:"type:jsonb;not null" json:"selector"`
	CreatedBy    *uint  `json:"createdBy"`
	UpdatedBy    *uint  `json:"updatedBy"`
	Namespace    Namespace `gorm:"foreignKey:NamespaceID" json:"namespace,omitempty"`
}

// OperationLog 操作日志模型
type OperationLog struct {
	BaseModel
	OperationID   string    `gorm:"uniqueIndex;not null" json:"operationId"`
	UserID        *uint     `json:"userId"`
	Username      string    `gorm:"size:50" json:"username"`
	Action        string    `gorm:"size:50;not null" json:"action"`
	ResourceType  string    `gorm:"size:50;not null" json:"resourceType"`
	ResourceID    *uint     `json:"resourceId"`
	ResourceName  string    `gorm:"size:253" json:"resourceName"`
	NamespaceID   *uint     `json:"namespaceId"`
	RequestMethod string    `gorm:"size:10" json:"requestMethod"`
	RequestPath   string    `gorm:"size:500" json:"requestPath"`
	RequestBody   JSONB     `gorm:"type:jsonb" json:"requestBody"`
	RequestHeaders JSONB   `gorm:"type:jsonb" json:"requestHeaders"`
	ClientIP      string    `gorm:"size:45" json:"clientIp"`
	UserAgent     string    `gorm:"type:text" json:"userAgent"`
	StatusCode   *int      `json:"statusCode"`
	ResponseBody  JSONB     `gorm:"type:jsonb" json:"responseBody"`
	DurationMs    *int      `json:"durationMs"`
	ErrorCode     string    `gorm:"size:50" json:"errorCode"`
	ErrorMessage  string    `gorm:"type:text" json:"errorMessage"`
	StartedAt     time.Time `gorm:"not null" json:"startedAt"`
	CompletedAt   *time.Time `json:"completedAt"`
	Metadata      JSONB     `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	User          User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Namespace     Namespace `gorm:"foreignKey:NamespaceID" json:"namespace,omitempty"`
}

// ContainerResourceUsage 容器资源使用记录
type ContainerResourceUsage struct {
	BaseModel
	ContainerID     uint    `gorm:"not null" json:"containerId"`
	Timestamp       time.Time `gorm:"not null" json:"timestamp"`
	CPUCoresUsed    float64 `json:"cpuCoresUsed"`
	CPUCoresRequest float64 `json:"cpuCoresRequest"`
	CPUCoresLimit   float64 `json:"cpuCoresLimit"`
	CPUUsagePercent float64 `json:"cpuUsagePercent"`
	MemoryBytesUsed int64   `json:"memoryBytesUsed"`
	MemoryBytesRequest int64 `json:"memoryBytesRequest"`
	MemoryBytesLimit int64   `json:"memoryBytesLimit"`
	MemoryUsagePercent float64 `json:"memoryUsagePercent"`
	NetworkBytesRx  int64   `json:"networkBytesRx"`
	NetworkBytesTx  int64   `json:"networkBytesTx"`
	NetworkPacketsRx int64  `json:"networkPacketsRx"`
	NetworkPacketsTx int64  `json:"networkPacketsTx"`
	DiskBytesUsed   int64   `json:"diskBytesUsed"`
	DiskBytesTotal  int64   `json:"diskBytesTotal"`
	DiskUsagePercent float64 `json:"diskUsagePercent"`
	FilesystemReads int64   `json:"filesystemReads"`
	FilesystemWrites int64  `json:"filesystemWrites"`
	Metadata        JSONB   `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	Container       Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
}

// K8sConnection Kubernetes 连接配置模型
type K8sConnection struct {
	BaseModel
	Name        string `gorm:"not null" json:"name"`
	Endpoint    string `gorm:"not null" json:"endpoint"`
	ConfigType  string `gorm:"not null;size:20" json:"configType"` // kubeconfig, token
	Config      string `gorm:"type:text" json:"config,omitempty"`   // kubeconfig 内容
	Token       string `gorm:"type:text" json:"token,omitempty"`    // service account token
	Namespace   string `gorm:"size:63;default:default" json:"namespace"`
	IsActive    bool   `gorm:"default:false" json:"isActive"`
	CreatedBy   *uint  `json:"createdBy"`
	UpdatedBy   *uint  `json:"updatedBy"`
}

// JSONB 自定义类型
type JSONB map[string]interface{}
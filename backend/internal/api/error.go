package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorCode 错误代码
type ErrorCode string

const (
	// 认证错误
	ErrUnauthorized          ErrorCode = "UNAUTHORIZED"
	ErrTokenExpired         ErrorCode = "TOKEN_EXPIRED"
	ErrInvalidToken         ErrorCode = "INVALID_TOKEN"
	ErrMissingToken         ErrorCode = "MISSING_TOKEN"
	ErrTokenInvalidFormat  ErrorCode = "INVALID_TOKEN_FORMAT"
	ErrInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"
	ErrRoleNotFound         ErrorCode = "ROLE_NOT_FOUND"
	ErrInvalidRole          ErrorCode = "INVALID_ROLE"

	// 验证错误
	ErrValidationFailed     ErrorCode = "VALIDATION_FAILED"
	ErrInvalidRequest       ErrorCode = "INVALID_REQUEST"
	ErrMissingField         ErrorCode = "MISSING_FIELD"
	ErrInvalidFieldType     ErrorCode = "INVALID_FIELD_TYPE"
	ErrFieldTooShort        ErrorCode = "FIELD_TOO_SHORT"
	ErrFieldTooLong         ErrorCode = "FIELD_TOO_LONG"
	ErrInvalidEmail         ErrorCode = "INVALID_EMAIL"
	ErrInvalidFormat        ErrorCode = "INVALID_FORMAT"

	// 业务错误
	ErrResourceNotFound     ErrorCode = "RESOURCE_NOT_FOUND"
	ErrResourceAlreadyExists ErrorCode = "RESOURCE_ALREADY_EXISTS"
	ErrOperationFailed     ErrorCode = "OPERATION_FAILED"
	ErrResourceLocked       ErrorCode = "RESOURCE_LOCKED"
	ErrDependencyFailed    ErrorCode = "DEPENDENCY_FAILED"
	ErrQuotaExceeded        ErrorCode = "QUOTA_EXCEEDED"
	ErrInvalidConfiguration ErrorCode = "INVALID_CONFIGURATION"

	// 系统错误
	ErrInternalServer       ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrDatabaseError        ErrorCode = "DATABASE_ERROR"
	ErrNetworkError         ErrorCode = "NETWORK_ERROR"
	ErrTimeout              ErrorCode = "TIMEOUT"
	ErrServiceUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
	ErrConfigurationError   ErrorCode = "CONFIGURATION_ERROR"

	// Kubernetes错误
	ErrKubernetesError      ErrorCode = "KUBERNETES_ERROR"
	ErrPodNotFound          ErrorCode = "POD_NOT_FOUND"
	ErrPodCreationFailed    ErrorCode = "POD_CREATION_FAILED"
	ErrServiceNotFound      ErrorCode = "SERVICE_NOT_FOUND"
	ErrNamespaceNotFound    ErrorCode = "NAMESPACE_NOT_FOUND"
	ErrConfigMapNotFound     ErrorCode = "CONFIGMAP_NOT_FOUND"
	ErrSecretNotFound        ErrorCode = "SECRET_NOT_FOUND"
	ErrImagePullFailed       ErrorCode = "IMAGE_PULL_FAILED"
	ErrResourceLimitExceeded ErrorCode = "RESOURCE_LIMIT_EXCEEDED"
)

// APIError API错误结构
type APIError struct {
	Code    ErrorCode              `json:"code"`
	Message string                  `json:"message"`
	Details interface{}             `json:"details,omitempty"`
}

// Error 返回错误响应
func Error(c *gin.Context, code ErrorCode, message string) {
	response := gin.H{
		"success": false,
		"error": APIError{
			Code:    code,
			Message: message,
		},
	}
	c.JSON(getHTTPStatus(code), response)
}

// ErrorWithDetails 返回带详细信息的错误响应
func ErrorWithDetails(c *gin.Context, code ErrorCode, message string, details interface{}) {
	response := gin.H{
		"success": false,
		"error": APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	c.JSON(getHTTPStatus(code), response)
}

// ValidationError 返回验证错误响应
func ValidationError(c *gin.Context, details interface{}) {
	ErrorWithDetails(c, ErrValidationFailed, "请求参数验证失败", details)
}

// NotFound 返回资源不存在错误
func NotFound(c *gin.Context, resource string) {
	Error(c, ErrResourceNotFound, fmt.Sprintf("%s不存在", resource))
}

// Unauthorized 返回未授权错误
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权访问"
	}
	Error(c, ErrUnauthorized, message)
}

// Forbidden 返回权限不足错误
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "权限不足"
	}
	Error(c, ErrInsufficientPermissions, message)
}

// InternalServerError 返回内部服务器错误
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "内部服务器错误"
	}
	Error(c, ErrInternalServer, message)
}

// BadRequest 返回请求参数错误
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "请求参数错误"
	}
	Error(c, ErrInvalidRequest, message)
}

// ServiceUnavailable 返回服务不可用错误
func ServiceUnavailable(c *gin.Context, message string) {
	if message == "" {
		message = "服务暂时不可用"
	}
	Error(c, ErrServiceUnavailable, message)
}

// Timeout 返回超时错误
func Timeout(c *gin.Context, message string) {
	if message == "" {
		message = "请求超时"
	}
	Error(c, ErrTimeout, message)
}

// getHTTPStatus 获取HTTP状态码
func getHTTPStatus(code ErrorCode) int {
	statusCodes := map[ErrorCode]int{
		// 认证错误
		ErrUnauthorized:          http.StatusUnauthorized,
		ErrTokenExpired:         http.StatusUnauthorized,
		ErrInvalidToken:         http.StatusUnauthorized,
		ErrMissingToken:         http.StatusUnauthorized,
		ErrTokenInvalidFormat:  http.StatusUnauthorized,
		ErrInsufficientPermissions: http.StatusForbidden,
		ErrRoleNotFound:         http.StatusForbidden,
		ErrInvalidRole:          http.StatusForbidden,

		// 验证错误
		ErrValidationFailed:     http.StatusBadRequest,
		ErrInvalidRequest:       http.StatusBadRequest,
		ErrMissingField:         http.StatusBadRequest,
		ErrInvalidFieldType:     http.StatusBadRequest,
		ErrFieldTooShort:        http.StatusBadRequest,
		ErrFieldTooLong:         http.StatusBadRequest,
		ErrInvalidEmail:         http.StatusBadRequest,
		ErrInvalidFormat:        http.StatusBadRequest,

		// 业务错误
		ErrResourceNotFound:     http.StatusNotFound,
		ErrResourceAlreadyExists: http.StatusConflict,
		ErrOperationFailed:     http.StatusInternalServerError,
		ErrResourceLocked:       http.StatusLocked,
		ErrDependencyFailed:    http.StatusServiceUnavailable,
		ErrQuotaExceeded:        http.StatusForbidden,
		ErrInvalidConfiguration: http.StatusBadRequest,

		// 系统错误
		ErrInternalServer:       http.StatusInternalServerError,
		ErrDatabaseError:        http.StatusInternalServerError,
		ErrNetworkError:         http.StatusServiceUnavailable,
		ErrTimeout:              http.StatusRequestTimeout,
		ErrServiceUnavailable:   http.StatusServiceUnavailable,
		ErrConfigurationError:   http.StatusInternalServerError,

		// Kubernetes错误
		ErrKubernetesError:      http.StatusBadGateway,
		ErrPodNotFound:          http.StatusNotFound,
		ErrPodCreationFailed:    http.StatusInternalServerError,
		ErrServiceNotFound:      http.StatusNotFound,
		ErrNamespaceNotFound:    http.StatusNotFound,
		ErrConfigMapNotFound:     http.StatusNotFound,
		ErrSecretNotFound:        http.StatusNotFound,
		ErrImagePullFailed:       http.StatusInternalServerError,
		ErrResourceLimitExceeded: http.StatusForbidden,
	}

	if status, exists := statusCodes[code]; exists {
		return status
	}

	return http.StatusInternalServerError
}

// ErrorResponse 返回错误响应
func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	response := gin.H{
		"success": false,
		"message": message,
	}
	if err != nil {
		response["error"] = err.Error()
	}
	c.JSON(statusCode, response)
}

// SuccessResponse 返回成功响应
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	response := gin.H{
		"success": true,
		"message": message,
		"data":    data,
	}
	c.JSON(http.StatusOK, response)
}

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否有错误被设置
		if len(c.Errors) > 0 {
			lastError := c.Errors.Last()

			// 根据错误类型返回适当的响应
			switch e := lastError.Err.(type) {
			case *gin.Error:
				// Gin框架错误
				Error(c, ErrInternalServer, e.Error())
			default:
				// 其他错误
				Error(c, ErrInternalServer, lastError.Error())
			}
		}
		c.Next()
	}
}
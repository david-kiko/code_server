package middleware

import (
	"container-platform-backend/internal/model"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey     string
	ExpiresIn     time.Duration
	RefreshExpire time.Duration
}

// DefaultJWTConfig 默认JWT配置
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		SecretKey:     "your-secret-key-change-in-production",
		ExpiresIn:     24 * time.Hour,
		RefreshExpire: 7 * 24 * time.Hour,
	}
}

// JWTAuth JWT认证中间件
type JWTAuth struct {
	config *JWTConfig
}

// NewJWTAuth 创建JWT认证中间件
func NewJWTAuth(config *JWTConfig) *JWTAuth {
	if config == nil {
		config = DefaultJWTConfig()
	}
	return &JWTAuth{config: config}
}

// GenerateToken 生成JWT令牌
func (j *JWTAuth) GenerateToken(user *model.User) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     getUserRole(user),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "container-platform",
			Subject:   "user-auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTAuth) GenerateRefreshToken(user *model.User) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     getUserRole(user),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.RefreshExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "container-platform",
			Subject:   "refresh-token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// ValidateToken 验证JWT令牌
func (j *JWTAuth) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// AuthMiddleware 认证中间件
func (j *JWTAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "MISSING_TOKEN",
					"message": "认证令牌缺失",
				},
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "认证令牌格式无效",
				},
			})
			c.Abort()
			return
		}

		// 提取令牌
		token := authHeader[len(bearerPrefix):]
		claims, err := j.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "认证令牌无效",
					"details": err.Error(),
				},
			})
			c.Abort()
			return
		}

		// 检查令牌是否过期
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TOKEN_EXPIRED",
					"message": "认证令牌已过期",
				},
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole 角色验证中间件
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ROLE_NOT_FOUND",
					"message": "用户角色信息缺失",
				},
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_ROLE",
					"message": "用户角色信息无效",
				},
			})
			c.Abort()
			return
		}

		// 管理员可以访问所有资源
		if role == "admin" {
			c.Next()
			return
		}

		// 检查角色权限
		if !hasRolePermission(role, requiredRole) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INSUFFICIENT_PERMISSIONS",
					"message": "权限不足",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
func (j *JWTAuth) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有认证令牌，继续处理
			c.Next()
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.Next()
			return
		}

		token := authHeader[len(bearerPrefix):]
		claims, err := j.ValidateToken(token)
		if err != nil {
			// 令牌无效，继续处理但不设置用户信息
			c.Next()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RefreshTokenMiddleware 刷新令牌中间件
func (j *JWTAuth) RefreshTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			RefreshToken string `json:"refreshToken" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_REQUEST",
					"message": "请求参数无效",
					"details": err.Error(),
				},
			})
			return
		}

		claims, err := j.ValidateToken(request.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_REFRESH_TOKEN",
					"message": "刷新令牌无效",
					"details": err.Error(),
				},
			})
			return
		}

		// 这里应该从数据库重新获取用户信息
		// 为了简化，我们直接使用令牌中的信息
		user := &model.User{
			ID:       claims.UserID,
			Username: claims.Username,
		}

		// 生成新的访问令牌
		newToken, err := j.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TOKEN_GENERATION_FAILED",
					"message": "令牌生成失败",
					"details": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"token": newToken,
				"expiresIn": j.config.ExpiresIn.Seconds(),
			},
		})
	}
}

// Helper functions

func getUserRole(user *model.User) string {
	// 简化版本，实际应该查询用户的角色
	if len(user.Roles) > 0 {
		return user.Roles[0].Name
	}
	return "operator" // 默认角色
}

func hasRolePermission(userRole, requiredRole string) bool {
	permissions := map[string][]string{
		"admin":    {"admin", "operator", "viewer"},
		"operator": {"operator", "viewer"},
		"viewer":   {"viewer"},
	}

	if allowedRoles, exists := permissions[userRole]; exists {
		for _, role := range allowedRoles {
			if role == requiredRole {
				return true
			}
		}
	}
	return false
}
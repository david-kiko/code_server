package service

import (
	"container-platform-backend/internal/middleware"
	"container-platform-backend/internal/model"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db       *gorm.DB
	jwtAuth  *middleware.JWTAuth
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB, jwtAuth *middleware.JWTAuth) *UserService {
	return &UserService{
		db:      db,
		jwtAuth: jwtAuth,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"fullName" binding:"max=100"`
	Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int64     `json:"expiresIn"`
	User         *UserResp `json:"user"`
}

// UserResp 用户响应
type UserResp struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *CreateUserRequest) (*UserResp, error) {
	// 检查用户名是否已存在
	var existingUser model.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("邮箱已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Status:       "active",
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 保存用户
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 创建或获取角色
	role, err := s.getOrCreateRole(tx, req.Role)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("处理角色失败: %w", err)
	}

	// 分配角色给用户
	userRole := &model.UserRole{
		UserID:     user.ID,
		RoleID:     role.ID,
		AssignedBy: &user.ID,
		AssignedAt: time.Now(),
	}
	if err := tx.Create(userRole).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("分配角色失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return s.toUserResponse(user), nil
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest) (*AuthResponse, error) {
	// 查找用户
	var user model.User
	if err := s.db.Preload("Roles").Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户账户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	if err := s.db.Model(&user).Update("last_login_at", &now).Error; err != nil {
		// 不影响登录流程，只记录日志
		fmt.Printf("更新最后登录时间失败: %v\n", err)
	}

	// 生成JWT令牌
	token, err := s.jwtAuth.GenerateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err := s.jwtAuth.GenerateRefreshToken(&user)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtAuth.ExpiresIn.Seconds()),
		User:         s.toUserResponse(&user),
	}, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(userID uint) (*UserResp, error) {
	var user model.User
	if err := s.db.Preload("Roles").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return s.toUserResponse(&user), nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(userID uint, updates map[string]interface{}) (*UserResp, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 更新用户信息
	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	// 重新查询用户信息
	if err := s.db.Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("查询更新后的用户信息失败: %w", err)
	}

	return s.toUserResponse(&user), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(userID uint) error {
	// 检查用户是否存在
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 软删除用户
	if err := s.db.Delete(&user).Error; err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize int, search string) ([]*UserResp, int64, error) {
	var users []*model.User
	var total int64

	query := s.db.Model(&model.User{}).Preload("Roles")

	// 添加搜索条件
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR full_name LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户列表失败: %w", err)
	}

	// 转换为响应格式
	userResponses := make([]*UserResp, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	return userResponses, total, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// RefreshToken 刷新令牌
func (s *UserService) RefreshToken(refreshToken string) (*AuthResponse, error) {
	// 验证刷新令牌
	claims, err := s.jwtAuth.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌无效: %w", err)
	}

	// 查询用户信息
	var user model.User
	if err := s.db.Preload("Roles").First(&user, claims.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("用户账户已被禁用")
	}

	// 生成新的访问令牌
	token, err := s.jwtAuth.GenerateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("生成新令牌失败: %w", err)
	}

	return &AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtAuth.ExpiresIn.Seconds()),
		User:         s.toUserResponse(&user),
	}, nil
}

// Helper methods

func (s *UserService) getOrCreateRole(tx *gorm.DB, roleName string) (*model.Role, error) {
	var role model.Role
	if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新角色
			role = model.Role{
				Name:        roleName,
				DisplayName: getRoleDisplayName(roleName),
				IsSystem:    false,
				Permissions: getRolePermissions(roleName),
			}
			if err := tx.Create(&role).Error; err != nil {
				return nil, fmt.Errorf("创建角色失败: %w", err)
			}
			return &role, nil
		}
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	return &role, nil
}

func (s *UserService) toUserResponse(user *model.User) *UserResp {
	role := "viewer"
	if len(user.Roles) > 0 {
		role = user.Roles[0].Name
	}

	return &UserResp{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     role,
		Status:   user.Status,
	}
}

func getRoleDisplayName(roleName string) string {
	displayNames := map[string]string{
		"admin":    "系统管理员",
		"operator": "操作员",
		"viewer":   "查看者",
	}
	if displayName, exists := displayNames[roleName]; exists {
		return displayName
	}
	return roleName
}

func getRolePermissions(roleName string) interface{} {
	permissions := map[string]interface{}{
		"admin":    []string{"*"},
		"operator": []string{"containers:*", "namespaces:read", "services:*"},
		"viewer":   []string{"containers:read", "namespaces:read", "services:read"},
	}
	if perms, exists := permissions[roleName]; exists {
		return perms
	}
	return []string{}
}
import { apiClient, ApiResponse } from './api'

// 认证相关接口
export interface LoginRequest {
  username: string
  password: string
  remember?: boolean
}

export interface LoginResponse {
  user: {
    id: number
    username: string
    email: string
    fullName: string
    role: string
    status: string
  }
  token: string
  refreshToken: string
  expiresIn: number
}

export interface RefreshTokenRequest {
  refreshToken: string
}

export interface RefreshTokenResponse {
  token: string
  refreshToken: string
  expiresIn: number
}

export interface ChangePasswordRequest {
  currentPassword: string
  newPassword: string
}

export interface User {
  id: number
  username: string
  email: string
  fullName: string
  role: string
  status: string
  createdAt: string
  updatedAt: string
}

export interface UserListResponse {
  items: User[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export interface CreateUserRequest {
  username: string
  email: string
  fullName: string
  password: string
  role: string
}

export interface UpdateUserRequest {
  email?: string
  fullName?: string
  role?: string
  status?: string
}

// 认证服务类
class AuthService {
  // 登录
  async login(credentials: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    const response = await apiClient.post<LoginResponse>('/auth/login', credentials)

    // 存储 token
    if (response.success && response.data) {
      this.storeTokens(response.data.token, response.data.refreshToken, credentials.remember || false)
    }

    return response
  }

  // 登出
  async logout(): Promise<ApiResponse<void>> {
    try {
      const response = await apiClient.post<void>('/auth/logout')
      this.clearTokens()
      return response
    } catch (error) {
      // 即使请求失败也要清除本地 token
      this.clearTokens()
      throw error
    }
  }

  // 刷新 token
  async refreshToken(): Promise<ApiResponse<RefreshTokenResponse>> {
    const refreshToken = this.getRefreshToken()
    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    const response = await apiClient.post<RefreshTokenResponse>('/auth/refresh', {
      refreshToken,
    })

    // 更新存储的 token
    if (response.success && response.data) {
      this.storeTokens(response.data.token, response.data.refreshToken, false)
    }

    return response
  }

  // 获取当前用户信息
  async getCurrentUser(): Promise<ApiResponse<User>> {
    return await apiClient.get<User>('/auth/me')
  }

  // 修改密码
  async changePassword(data: ChangePasswordRequest): Promise<ApiResponse<void>> {
    return await apiClient.post<void>('/auth/change-password', data)
  }

  // 用户管理 - 获取用户列表
  async getUsers(page = 1, pageSize = 20, filters?: {
    search?: string
    role?: string
    status?: string
  }): Promise<ApiResponse<UserListResponse>> {
    const params = new URLSearchParams({
      page: page.toString(),
      pageSize: pageSize.toString(),
    })

    if (filters?.search) {
      params.append('search', filters.search)
    }
    if (filters?.role) {
      params.append('role', filters.role)
    }
    if (filters?.status) {
      params.append('status', filters.status)
    }

    return await apiClient.get<UserListResponse>(`/users?${params}`)
  }

  // 获取单个用户
  async getUser(id: number): Promise<ApiResponse<User>> {
    return await apiClient.get<User>(`/users/${id}`)
  }

  // 创建用户
  async createUser(userData: CreateUserRequest): Promise<ApiResponse<User>> {
    return await apiClient.post<User>('/users', userData)
  }

  // 更新用户
  async updateUser(id: number, userData: UpdateUserRequest): Promise<ApiResponse<User>> {
    return await apiClient.put<User>(`/users/${id}`, userData)
  }

  // 删除用户
  async deleteUser(id: number): Promise<ApiResponse<void>> {
    return await apiClient.delete<void>(`/users/${id}`)
  }

  // 重置用户密码
  async resetUserPassword(id: number, newPassword: string): Promise<ApiResponse<void>> {
    return await apiClient.post<void>(`/users/${id}/reset-password`, {
      password: newPassword,
    })
  }

  // Token 存储和获取
  private storeTokens(token: string, refreshToken: string, remember: boolean): void {
    if (typeof window !== 'undefined') {
      if (remember) {
        localStorage.setItem('auth_token', token)
        localStorage.setItem('refresh_token', refreshToken)
      } else {
        sessionStorage.setItem('auth_token', token)
        sessionStorage.setItem('refresh_token', refreshToken)
      }
    }
  }

  private clearTokens(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('refresh_token')
      sessionStorage.removeItem('auth_token')
      sessionStorage.removeItem('refresh_token')
    }
  }

  public getAccessToken(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token')
    }
    return null
  }

  public getRefreshToken(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('refresh_token') || sessionStorage.getItem('refresh_token')
    }
    return null
  }

  // 检查是否已认证
  public isAuthenticated(): boolean {
    const token = this.getAccessToken()
    if (!token) return false

    try {
      // 检查 token 是否过期
      const payload = JSON.parse(atob(token.split('.')[1]))
      const currentTime = Date.now() / 1000
      return payload.exp > currentTime
    } catch {
      return false
    }
  }

  // 检查用户权限
  public hasRole(requiredRole: string): boolean {
    const token = this.getAccessToken()
    if (!token) return false

    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      return payload.role === requiredRole || payload.role === 'admin'
    } catch {
      return false
    }
  }

  // 检查是否为管理员
  public isAdmin(): boolean {
    return this.hasRole('admin')
  }
}

// 创建并导出认证服务实例
export const authService = new AuthService()

export default authService
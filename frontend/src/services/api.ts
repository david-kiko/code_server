import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'

// API 响应接口
export interface ApiResponse<T = any> {
  success: boolean
  data: T
  message?: string
  code?: number
}

// 分页响应接口
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

// 错误响应接口
export interface ApiError {
  code: number
  message: string
  details?: any
  timestamp: string
}

// API 客户端类
class ApiClient {
  private client: AxiosInstance
  private static instance: ApiClient

  constructor() {
    this.client = axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || '/api',
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.setupInterceptors()
  }

  public static getInstance(): ApiClient {
    if (!ApiClient.instance) {
      ApiClient.instance = new ApiClient()
    }
    return ApiClient.instance
  }

  private setupInterceptors(): void {
    // 请求拦截器
    this.client.interceptors.request.use(
      (config) => {
        // 添加认证 token
        const token = this.getAuthToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }

        // 添加请求 ID
        config.headers['X-Request-ID'] = this.generateRequestId()

        // 添加时间戳
        config.headers['X-Timestamp'] = Date.now().toString()

        console.log(`[API Request] ${config.method?.toUpperCase()} ${config.url}`)
        return config
      },
      (error) => {
        console.error('[API Request Error]', error)
        return Promise.reject(error)
      }
    )

    // 响应拦截器
    this.client.interceptors.response.use(
      (response: AxiosResponse) => {
        console.log(`[API Response] ${response.status} ${response.config.url}`)
        return response
      },
      (error) => {
        console.error('[API Response Error]', error)

        // 处理认证错误
        if (error.response?.status === 401) {
          this.handleAuthError()
        }

        // 处理网络错误
        if (!error.response) {
          this.handleNetworkError(error)
        }

        return Promise.reject(this.formatError(error))
      }
    )
  }

  private getAuthToken(): string | null {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token')
    }
    return null
  }

  private generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private handleAuthError(): void {
    if (typeof window !== 'undefined') {
      // 清除认证信息
      localStorage.removeItem('auth_token')
      sessionStorage.removeItem('auth_token')

      // 重定向到登录页面
      window.location.href = '/login'
    }
  }

  private handleNetworkError(error: any): void {
    console.error('Network error:', error)
    // 可以在这里添加网络错误通知逻辑
  }

  private formatError(error: any): ApiError {
    if (error.response) {
      // 服务器响应错误
      return {
        code: error.response.status,
        message: error.response.data?.message || error.message,
        details: error.response.data?.details,
        timestamp: new Date().toISOString(),
      }
    } else if (error.request) {
      // 网络错误
      return {
        code: 0,
        message: 'Network error. Please check your connection.',
        timestamp: new Date().toISOString(),
      }
    } else {
      // 其他错误
      return {
        code: -1,
        message: error.message || 'Unknown error occurred',
        timestamp: new Date().toISOString(),
      }
    }
  }

  // HTTP 方法
  public async get<T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.client.get<ApiResponse<T>>(url, config)
    return response.data
  }

  public async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.client.post<ApiResponse<T>>(url, data, config)
    return response.data
  }

  public async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.client.put<ApiResponse<T>>(url, data, config)
    return response.data
  }

  public async patch<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.client.patch<ApiResponse<T>>(url, data, config)
    return response.data
  }

  public async delete<T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const response = await this.client.delete<ApiResponse<T>>(url, config)
    return response.data
  }

  // 文件上传
  public async upload<T>(url: string, file: File, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await this.client.post<ApiResponse<T>>(url, formData, {
      ...config,
      headers: {
        'Content-Type': 'multipart/form-data',
        ...config?.headers,
      },
    })

    return response.data
  }

  // 文件下载
  public async download(url: string, filename?: string, config?: AxiosRequestConfig): Promise<void> {
    const response = await this.client.get(url, {
      ...config,
      responseType: 'blob',
    })

    const blob = new Blob([response.data])
    const downloadUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = filename || 'download'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(downloadUrl)
  }

  // 取消请求
  public createCancelToken() {
    return axios.CancelToken.source()
  }

  // 检查是否为取消错误
  public isCancel(error: any): boolean {
    return axios.isCancel(error)
  }
}

// 创建并导出 API 客户端实例
export const apiClient = ApiClient.getInstance()

// 导出默认实例
export default apiClient

// 导出常用类型
export type { AxiosRequestConfig, AxiosResponse }
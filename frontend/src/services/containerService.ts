import { apiClient, ApiResponse, PaginatedResponse } from './api'
import { Container, ContainerConfig } from '@/store/slices/containerSlice'

// 容器相关接口
export interface ContainerListParams {
  page?: number
  pageSize?: number
  namespace?: string
  status?: string
  search?: string
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

export interface ContainerCreateRequest {
  config: ContainerConfig
  namespace: string
}

export interface ContainerActionRequest {
  action: 'start' | 'stop' | 'restart' | 'pause' | 'resume' | 'destroy'
}

export interface ContainerLogParams {
  containerId: string
  namespace?: string
  follow?: boolean
  tail?: number
  since?: string
  timestamps?: boolean
}

export interface ContainerLogResponse {
  logs: string[]
  hasMore: boolean
  totalLines: number
}

export interface ContainerStats {
  cpu: {
    usage: string
    usagePercent: number
  }
  memory: {
    usage: string
    usagePercent: number
    limit: string
  }
  network: {
    rx: string
    tx: string
  }
  fs: {
    reads: string
    writes: string
  }
  timestamp: string
}

export interface ExecRequest {
  command: string[]
  container: string
  namespace?: string
  tty?: boolean
  stdin?: boolean
}

export interface ExecResponse {
  sessionId: string
  websocketUrl: string
}

// 容器服务类
class ContainerService {
  // 获取容器列表
  async getContainers(params: ContainerListParams = {}): Promise<ApiResponse<PaginatedResponse<Container>>> {
    const searchParams = new URLSearchParams()

    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        searchParams.append(key, value.toString())
      }
    })

    const queryString = searchParams.toString()
    const url = queryString ? `/containers?${queryString}` : '/containers'

    return await apiClient.get<PaginatedResponse<Container>>(url)
  }

  // 获取单个容器
  async getContainer(id: string, namespace?: string): Promise<ApiResponse<Container>> {
    const params = namespace ? `?namespace=${namespace}` : ''
    return await apiClient.get<Container>(`/containers/${id}${params}`)
  }

  // 创建容器
  async createContainer(request: ContainerCreateRequest): Promise<ApiResponse<Container>> {
    return await apiClient.post<Container>('/containers', request)
  }

  // 更新容器配置
  async updateContainer(id: string, config: Partial<ContainerConfig>, namespace?: string): Promise<ApiResponse<Container>> {
    const body = { config, namespace }
    return await apiClient.put<Container>(`/containers/${id}`, body)
  }

  // 容器操作
  async performAction(id: string, action: ContainerActionRequest, namespace?: string): Promise<ApiResponse<void>> {
    const body = { ...action, namespace }
    return await apiClient.post<void>(`/containers/${id}/action`, body)
  }

  // 启动容器
  async startContainer(id: string, namespace?: string): Promise<ApiResponse<void>> {
    return this.performAction(id, { action: 'start' }, namespace)
  }

  // 停止容器
  async stopContainer(id: string, namespace?: string): Promise<ApiResponse<void>> {
    return this.performAction(id, { action: 'stop' }, namespace)
  }

  // 重启容器
  async restartContainer(id: string, namespace?: string): Promise<ApiResponse<void>> {
    return this.performAction(id, { action: 'restart' }, namespace)
  }

  // 暂停容器
  async pauseContainer(id: string, namespace?: string): Promise<ApiResponse<void>> {
    return this.performAction(id, { action: 'pause' }, namespace)
  }

  // 恢复容器
  async resumeContainer(id: string, namespace?: string): Promise<ApiResponse<void>> {
    return this.performAction(id, { action: 'resume' }, namespace)
  }

  // 销毁容器
  async destroyContainer(id: string, namespace?: string): Promise<ApiResponse<void>> {
    return this.performAction(id, { action: 'destroy' }, namespace)
  }

  // 获取容器日志
  async getContainerLogs(params: ContainerLogParams): Promise<ApiResponse<ContainerLogResponse>> {
    const searchParams = new URLSearchParams()

    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== '') {
        searchParams.append(key, value.toString())
      }
    })

    return await apiClient.get<ContainerLogResponse>(`/containers/${params.containerId}/logs?${searchParams}`)
  }

  // 获取容器统计信息
  async getContainerStats(id: string, namespace?: string): Promise<ApiResponse<ContainerStats>> {
    const params = namespace ? `?namespace=${namespace}` : ''
    return await apiClient.get<ContainerStats>(`/containers/${id}/stats${params}`)
  }

  // 获取容器事件
  async getContainerEvents(id: string, namespace?: string, limit = 50): Promise<ApiResponse<any[]>> {
    const params = new URLSearchParams({ limit: limit.toString() })
    if (namespace) {
      params.append('namespace', namespace)
    }
    return await apiClient.get<any[]>(`/containers/${id}/events?${params}`)
  }

  // 在容器中执行命令
  async execCommand(request: ExecRequest): Promise<ApiResponse<ExecResponse>> {
    return await apiClient.post<ExecResponse>('/containers/exec', request)
  }

  // 批量操作
  async batchOperation(ids: string[], action: ContainerActionRequest['action'], namespace?: string): Promise<ApiResponse<void>> {
    return await apiClient.post<void>('/containers/batch', {
      ids,
      action,
      namespace,
    })
  }

  // 获取容器镜像信息
  async getContainerImages(): Promise<ApiResponse<any[]>> {
    return await apiClient.get<any[]>('/containers/images')
  }

  // 获取可用镜像列表
  async getAvailableImages(): Promise<ApiResponse<string[]>> {
    return await apiClient.get<string[]>('/containers/images/available')
  }

  // 拉取镜像
  async pullImage(imageName: string): Promise<ApiResponse<void>> {
    return await apiClient.post<void>('/containers/images/pull', { imageName })
  }

  // 删除镜像
  async deleteImage(imageName: string): Promise<ApiResponse<void>> {
    return await apiClient.delete<void>(`/containers/images/${encodeURIComponent(imageName)}`)
  }

  // 获取命名空间列表
  async getNamespaces(): Promise<ApiResponse<string[]>> {
    return await apiClient.get<string[]>('/namespaces')
  }

  // 验证容器配置
  async validateConfig(config: ContainerConfig): Promise<ApiResponse<{ valid: boolean; errors?: string[] }>> {
    return await apiClient.post<{ valid: boolean; errors?: string[] }>('/containers/validate', { config })
  }

  // 获取容器配置模板
  async getConfigTemplates(): Promise<ApiResponse<ContainerConfig[]>> {
    return await apiClient.get<ContainerConfig[]>('/containers/templates')
  }

  // 导出容器配置
  async exportConfig(id: string, namespace?: string): Promise<ApiResponse<any>> {
    const params = namespace ? `?namespace=${namespace}` : ''
    return await apiClient.get<any>(`/containers/${id}/export${params}`)
  }

  // 导入容器配置
  async importConfig(configData: any, namespace?: string): Promise<ApiResponse<Container>> {
    return await apiClient.post<Container>('/containers/import', {
      config: configData,
      namespace,
    })
  }

  // 获取容器资源使用趋势
  async getResourceUsage(id: string, timeRange: '1h' | '6h' | '24h' | '7d' = '24h', namespace?: string): Promise<ApiResponse<any[]>> {
    const params = new URLSearchParams({ timeRange })
    if (namespace) {
      params.append('namespace', namespace)
    }
    return await apiClient.get<any[]>(`/containers/${id}/usage?${params}`)
  }
}

// 创建并导出容器服务实例
export const containerService = new ContainerService()

export default containerService
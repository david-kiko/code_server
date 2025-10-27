import { apiClient } from './api'

// K8s 连接相关类型
export interface K8sConnection {
  id: string
  name: string
  endpoint: string
  configType: 'kubeconfig' | 'token'
  config?: string
  token?: string
  namespace?: string
  isActive: boolean
  createdAt: string
}

// 容器相关类型
export interface Container {
  name: string
  namespace: string
  image: string
  status: 'Running' | 'Pending' | 'Failed' | 'Succeeded' | 'Unknown'
  podName: string
  restartCount: number
  age: string
  node?: string
  labels?: Record<string, string>
  containerId?: string
}

export interface CreateContainerRequest {
  name: string
  namespace: string
  image: string
  command?: string
  ports?: string
  env?: string
  resources?: string
}

// K8s API 服务
export const k8sApi = {
  // 获取容器列表
  getContainers: async (namespace = 'default'): Promise<Container[]> => {
    try {
      const response = await apiClient.get<Container[]>(`/k8s/containers?namespace=${namespace}`)
      return response.data
    } catch (error) {
      console.error('Failed to fetch containers:', error)
      throw error
    }
  },

  // 创建容器
  createContainer: async (data: CreateContainerRequest): Promise<void> => {
    try {
      await apiClient.post('/k8s/containers', data)
    } catch (error) {
      console.error('Failed to create container:', error)
      throw error
    }
  },

  // 启动容器
  startContainer: async (namespace: string, podName: string): Promise<void> => {
    try {
      await apiClient.post(`/k8s/containers/${namespace}/${podName}/start`)
    } catch (error) {
      console.error('Failed to start container:', error)
      throw error
    }
  },

  // 停止容器
  stopContainer: async (namespace: string, podName: string): Promise<void> => {
    try {
      await apiClient.post(`/k8s/containers/${namespace}/${podName}/stop`)
    } catch (error) {
      console.error('Failed to stop container:', error)
      throw error
    }
  },

  // 重启容器
  restartContainer: async (namespace: string, podName: string): Promise<void> => {
    try {
      await apiClient.post(`/k8s/containers/${namespace}/${podName}/restart`)
    } catch (error) {
      console.error('Failed to restart container:', error)
      throw error
    }
  },

  // 删除容器
  deleteContainer: async (namespace: string, podName: string): Promise<void> => {
    try {
      await apiClient.delete(`/k8s/containers/${namespace}/${podName}`)
    } catch (error) {
      console.error('Failed to delete container:', error)
      throw error
    }
  },

  // 测试 K8s 连接
  testConnection: async (connection: K8sConnection): Promise<void> => {
    try {
      await apiClient.post('/k8s/test-connection', connection)
    } catch (error) {
      console.error('Failed to test K8s connection:', error)
      throw error
    }
  },
}

export default k8sApi
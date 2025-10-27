import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export interface ClusterInfo {
  name: string
  version: string
  nodeCount: number
  namespaceCount: number
  podCount: number
  serviceCount: number
  resourceQuota: {
    cpu: {
      used: string
      total: string
    }
    memory: {
      used: string
      total: string
    }
    storage: {
      used: string
      total: string
    }
  }
}

export interface NodeInfo {
  name: string
  status: 'Ready' | 'NotReady' | 'Unknown' | 'MemoryPressure' | 'DiskPressure'
  roles: string[]
  version: string
  internalIP: string
  externalIP?: string
  os: string
  kernelVersion: string
  containerRuntime: string
  resources: {
    cpu: {
      allocatable: string
      capacity: string
      allocated: string
      usage: string
    }
    memory: {
      allocatable: string
      capacity: string
      allocated: string
      usage: string
    }
    pods: {
      capacity: number
      allocated: number
      running: number
    }
  }
  conditions: Array<{
    type: string
    status: 'True' | 'False' | 'Unknown'
    lastTransitionTime: string
    reason: string
    message: string
  }>
  createdAt: string
}

export interface RecentActivity {
  id: string
  type: 'container' | 'service' | 'configmap' | 'secret' | 'deployment'
  action: 'create' | 'update' | 'delete' | 'start' | 'stop' | 'restart'
  resource: string
  namespace: string
  user: string
  status: 'success' | 'failed' | 'pending'
  timestamp: string
  details?: string
}

export interface ResourceUsage {
  timestamp: string
  cpu: number // 0-100 percentage
  memory: number // 0-100 percentage
  disk: number // 0-100 percentage
  network: {
    inbound: number // bytes/second
    outbound: number // bytes/second
  }
}

export interface Alert {
  id: string
  severity: 'critical' | 'warning' | 'info'
  title: string
  message: string
  resource?: string
  namespace?: string
  timestamp: string
  acknowledged: boolean
  acknowledgedBy?: string
  acknowledgedAt?: string
}

interface DashboardState {
  // 集群信息
  clusterInfo: ClusterInfo | null

  // 节点信息
  nodes: NodeInfo[]
  selectedNode: NodeInfo | null

  // 最近活动
  recentActivities: RecentActivity[]
  activitiesLoading: boolean
  activitiesError: string | null

  // 资源使用历史
  resourceUsage: ResourceUsage[]
  resourceUsageLoading: boolean
  resourceUsageError: string | null

  // 告警信息
  alerts: Alert[]
  alertsLoading: boolean
  alertsError: string | null
  alertFilters: {
    severity: string
    acknowledged: boolean
    namespace: string
  }

  // 统计数据
  statistics: {
    totalContainers: number
    runningContainers: number
    stoppedContainers: number
    errorContainers: number
    totalServices: number
    totalVolumes: number
    totalNamespaces: number
  }

  // 刷新设置
  autoRefresh: boolean
  refreshInterval: number // seconds
  lastRefresh: string | null

  // 仪表板布局
  widgets: Array<{
    id: string
    type: string
    title: string
    visible: boolean
    position: {
      x: number
      y: number
      w: number
      h: number
    }
    config: any
  }>

  // 加载状态
  loading: boolean
  error: string | null
}

const initialState: DashboardState = {
  clusterInfo: null,
  nodes: [],
  selectedNode: null,
  recentActivities: [],
  activitiesLoading: false,
  activitiesError: null,
  resourceUsage: [],
  resourceUsageLoading: false,
  resourceUsageError: null,
  alerts: [],
  alertsLoading: false,
  alertsError: null,
  alertFilters: {
    severity: 'all',
    acknowledged: false,
    namespace: 'all',
  },
  statistics: {
    totalContainers: 0,
    runningContainers: 0,
    stoppedContainers: 0,
    errorContainers: 0,
    totalServices: 0,
    totalVolumes: 0,
    totalNamespaces: 0,
  },
  autoRefresh: true,
  refreshInterval: 30,
  lastRefresh: null,
  widgets: [
    {
      id: 'cluster-overview',
      type: 'cluster-overview',
      title: '集群概览',
      visible: true,
      position: { x: 0, y: 0, w: 6, h: 4 },
      config: {},
    },
    {
      id: 'resource-usage',
      type: 'resource-usage',
      title: '资源使用',
      visible: true,
      position: { x: 6, y: 0, w: 6, h: 4 },
      config: {},
    },
    {
      id: 'recent-activities',
      type: 'recent-activities',
      title: '最近活动',
      visible: true,
      position: { x: 0, y: 4, w: 8, h: 6 },
      config: { limit: 10 },
    },
    {
      id: 'alerts',
      type: 'alerts',
      title: '告警信息',
      visible: true,
      position: { x: 8, y: 4, w: 4, h: 6 },
      config: { limit: 5 },
    },
    {
      id: 'nodes-status',
      type: 'nodes-status',
      title: '节点状态',
      visible: true,
      position: { x: 0, y: 10, w: 12, h: 4 },
      config: {},
    },
  ],
  loading: false,
  error: null,
}

const dashboardSlice = createSlice({
  name: 'dashboard',
  initialState,
  reducers: {
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload
    },

    // 集群信息
    setClusterInfo: (state, action: PayloadAction<ClusterInfo>) => {
      state.clusterInfo = action.payload
    },

    // 节点信息
    setNodes: (state, action: PayloadAction<NodeInfo[]>) => {
      state.nodes = action.payload
    },
    setSelectedNode: (state, action: PayloadAction<NodeInfo | null>) => {
      state.selectedNode = action.payload
    },
    updateNode: (state, action: PayloadAction<{ nodeName: string; updates: Partial<NodeInfo> }>) => {
      const { nodeName, updates } = action.payload
      const index = state.nodes.findIndex(n => n.name === nodeName)
      if (index !== -1) {
        state.nodes[index] = { ...state.nodes[index], ...updates }
      }
    },

    // 最近活动
    setActivitiesLoading: (state, action: PayloadAction<boolean>) => {
      state.activitiesLoading = action.payload
    },
    setActivitiesError: (state, action: PayloadAction<string | null>) => {
      state.activitiesError = action.payload
    },
    setRecentActivities: (state, action: PayloadAction<RecentActivity[]>) => {
      state.recentActivities = action.payload
    },
    addRecentActivity: (state, action: PayloadAction<RecentActivity>) => {
      state.recentActivities.unshift(action.payload)
      // 保持最多 100 条记录
      if (state.recentActivities.length > 100) {
        state.recentActivities = state.recentActivities.slice(0, 100)
      }
    },

    // 资源使用
    setResourceUsageLoading: (state, action: PayloadAction<boolean>) => {
      state.resourceUsageLoading = action.payload
    },
    setResourceUsageError: (state, action: PayloadAction<string | null>) => {
      state.resourceUsageError = action.payload
    },
    setResourceUsage: (state, action: PayloadAction<ResourceUsage[]>) => {
      state.resourceUsage = action.payload
    },
    addResourceUsageData: (state, action: PayloadAction<ResourceUsage>) => {
      state.resourceUsage.push(action.payload)
      // 保持最多 1 小时的数据 (每分钟一个数据点)
      if (state.resourceUsage.length > 60) {
        state.resourceUsage = state.resourceUsage.slice(-60)
      }
    },

    // 告警信息
    setAlertsLoading: (state, action: PayloadAction<boolean>) => {
      state.alertsLoading = action.payload
    },
    setAlertsError: (state, action: PayloadAction<string | null>) => {
      state.alertsError = action.payload
    },
    setAlerts: (state, action: PayloadAction<Alert[]>) => {
      state.alerts = action.payload
    },
    addAlert: (state, action: PayloadAction<Alert>) => {
      state.alerts.unshift(action.payload)
    },
    acknowledgeAlert: (state, action: PayloadAction<{ alertId: string; acknowledgedBy: string }>) => {
      const { alertId, acknowledgedBy } = action.payload
      const alert = state.alerts.find(a => a.id === alertId)
      if (alert) {
        alert.acknowledged = true
        alert.acknowledgedBy = acknowledgedBy
        alert.acknowledgedAt = new Date().toISOString()
      }
    },
    setAlertFilters: (state, action: PayloadAction<Partial<DashboardState['alertFilters']>>) => {
      state.alertFilters = { ...state.alertFilters, ...action.payload }
    },

    // 统计数据
    setStatistics: (state, action: PayloadAction<Partial<DashboardState['statistics']>>) => {
      state.statistics = { ...state.statistics, ...action.payload }
    },

    // 刷新设置
    setAutoRefresh: (state, action: PayloadAction<boolean>) => {
      state.autoRefresh = action.payload
    },
    setRefreshInterval: (state, action: PayloadAction<number>) => {
      state.refreshInterval = action.payload
    },
    setLastRefresh: (state, action: PayloadAction<string>) => {
      state.lastRefresh = action.payload
    },

    // 仪表板布局
    updateWidgets: (state, action: PayloadAction<DashboardState['widgets']>) => {
      state.widgets = action.payload
    },
    updateWidget: (state, action: PayloadAction<{ widgetId: string; updates: Partial<DashboardState['widgets'][0]> }>) => {
      const { widgetId, updates } = action.payload
      const widget = state.widgets.find(w => w.id === widgetId)
      if (widget) {
        Object.assign(widget, updates)
      }
    },

    // 重置状态
    resetState: (state) => {
      Object.assign(state, initialState)
    },
  },
})

export const {
  setLoading,
  setError,
  setClusterInfo,
  setNodes,
  setSelectedNode,
  updateNode,
  setActivitiesLoading,
  setActivitiesError,
  setRecentActivities,
  addRecentActivity,
  setResourceUsageLoading,
  setResourceUsageError,
  setResourceUsage,
  addResourceUsageData,
  setAlertsLoading,
  setAlertsError,
  setAlerts,
  addAlert,
  acknowledgeAlert,
  setAlertFilters,
  setStatistics,
  setAutoRefresh,
  setRefreshInterval,
  setLastRefresh,
  updateWidgets,
  updateWidget,
  resetState,
} = dashboardSlice.actions

export default dashboardSlice.reducer
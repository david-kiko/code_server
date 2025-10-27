import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export interface NotificationMessage {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message: string
  duration?: number
  timestamp: number
}

export interface ModalState {
  [modalId: string]: {
    visible: boolean
    title?: string
    data?: any
    loading?: boolean
  }
}

export interface LoadingState {
  [key: string]: boolean
}

interface UIState {
  // 主题设置
  theme: 'light' | 'dark' | 'auto'
  primaryColor: string

  // 侧边栏状态
  sidebarCollapsed: boolean
  sidebarWidth: number

  // 通知系统
  notifications: NotificationMessage[]

  // 模态框管理
  modals: ModalState

  // 全局加载状态
  loading: LoadingState

  // 页面标题
  pageTitle: string

  // 布局设置
  layout: {
    headerHeight: number
    footerHeight: number
    contentPadding: number
  }

  // 窗口大小
  viewport: {
    width: number
    height: number
    isMobile: boolean
    isTablet: boolean
    isDesktop: boolean
  }

  // 面包屑导航
  breadcrumbs: Array<{
    title: string
    path?: string
  }>

  // 快捷键状态
  shortcuts: {
    enabled: boolean
    helpVisible: boolean
  }
}

const initialState: UIState = {
  theme: 'auto',
  primaryColor: '#1890ff',
  sidebarCollapsed: false,
  sidebarWidth: 240,
  notifications: [],
  modals: {},
  loading: {},
  pageTitle: '容器编排管理平台',
  layout: {
    headerHeight: 64,
    footerHeight: 48,
    contentPadding: 24,
  },
  viewport: {
    width: 1920,
    height: 1080,
    isMobile: false,
    isTablet: false,
    isDesktop: true,
  },
  breadcrumbs: [],
  shortcuts: {
    enabled: true,
    helpVisible: false,
  },
}

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    // 主题相关
    setTheme: (state, action: PayloadAction<'light' | 'dark' | 'auto'>) => {
      state.theme = action.payload
    },
    setPrimaryColor: (state, action: PayloadAction<string>) => {
      state.primaryColor = action.payload
    },

    // 侧边栏相关
    toggleSidebar: (state) => {
      state.sidebarCollapsed = !state.sidebarCollapsed
    },
    setSidebarCollapsed: (state, action: PayloadAction<boolean>) => {
      state.sidebarCollapsed = action.payload
    },
    setSidebarWidth: (state, action: PayloadAction<number>) => {
      state.sidebarWidth = action.payload
    },

    // 通知系统
    addNotification: (state, action: PayloadAction<Omit<NotificationMessage, 'id' | 'timestamp'>>) => {
      const notification: NotificationMessage = {
        ...action.payload,
        id: Date.now().toString(),
        timestamp: Date.now(),
      }
      state.notifications.push(notification)
    },
    removeNotification: (state, action: PayloadAction<string>) => {
      state.notifications = state.notifications.filter(n => n.id !== action.payload)
    },
    clearNotifications: (state) => {
      state.notifications = []
    },

    // 模态框管理
    openModal: (state, action: PayloadAction<{ modalId: string; title?: string; data?: any }>) => {
      const { modalId, title, data } = action.payload
      state.modals[modalId] = {
        visible: true,
        title,
        data,
        loading: false,
      }
    },
    closeModal: (state, action: PayloadAction<string>) => {
      const modalId = action.payload
      if (state.modals[modalId]) {
        state.modals[modalId].visible = false
      }
    },
    setModalLoading: (state, action: PayloadAction<{ modalId: string; loading: boolean }>) => {
      const { modalId, loading } = action.payload
      if (state.modals[modalId]) {
        state.modals[modalId].loading = loading
      }
    },
    updateModalData: (state, action: PayloadAction<{ modalId: string; data: any }>) => {
      const { modalId, data } = action.payload
      if (state.modals[modalId]) {
        state.modals[modalId].data = data
      }
    },

    // 全局加载状态
    setLoading: (state, action: PayloadAction<{ key: string; loading: boolean }>) => {
      const { key, loading } = action.payload
      state.loading[key] = loading
    },
    clearLoading: (state, action: PayloadAction<string>) => {
      delete state.loading[action.payload]
    },

    // 页面标题
    setPageTitle: (state, action: PayloadAction<string>) => {
      state.pageTitle = action.payload
    },

    // 布局设置
    updateLayout: (state, action: PayloadAction<Partial<UIState['layout']>>) => {
      state.layout = { ...state.layout, ...action.payload }
    },

    // 窗口大小
    updateViewport: (state, action: PayloadAction<{ width: number; height: number }>) => {
      const { width, height } = action.payload
      state.viewport.width = width
      state.viewport.height = height
      state.viewport.isMobile = width < 768
      state.viewport.isTablet = width >= 768 && width < 1024
      state.viewport.isDesktop = width >= 1024
    },

    // 面包屑导航
    setBreadcrumbs: (state, action: PayloadAction<Array<{ title: string; path?: string }>>) => {
      state.breadcrumbs = action.payload
    },
    addBreadcrumb: (state, action: PayloadAction<{ title: string; path?: string }>) => {
      state.breadcrumbs.push(action.payload)
    },

    // 快捷键状态
    toggleShortcuts: (state) => {
      state.shortcuts.enabled = !state.shortcuts.enabled
    },
    setShortcutsEnabled: (state, action: PayloadAction<boolean>) => {
      state.shortcuts.enabled = action.payload
    },
    setShortcutsHelpVisible: (state, action: PayloadAction<boolean>) => {
      state.shortcuts.helpVisible = action.payload
    },
  },
})

export const {
  setTheme,
  setPrimaryColor,
  toggleSidebar,
  setSidebarCollapsed,
  setSidebarWidth,
  addNotification,
  removeNotification,
  clearNotifications,
  openModal,
  closeModal,
  setModalLoading,
  updateModalData,
  setLoading,
  clearLoading,
  setPageTitle,
  updateLayout,
  updateViewport,
  setBreadcrumbs,
  addBreadcrumb,
  toggleShortcuts,
  setShortcutsEnabled,
  setShortcutsHelpVisible,
} = uiSlice.actions

export default uiSlice.reducer
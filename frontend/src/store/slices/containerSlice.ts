import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export interface Container {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'paused' | 'error' | 'pending'
  namespace: string
  podName?: string
  ports: ContainerPort[]
  volumes: ContainerVolume[]
  resources: ContainerResources
  labels: Record<string, string>
  annotations: Record<string, string>
  createdAt: string
  updatedAt: string
}

export interface ContainerPort {
  name?: string
  containerPort: number
  protocol: 'TCP' | 'UDP'
  hostPort?: number
  servicePort?: number
}

export interface ContainerVolume {
  name: string
  mountPath: string
  volumeType: 'configMap' | 'secret' | 'persistentVolume' | 'hostPath'
  source: string
  readOnly?: boolean
}

export interface ContainerResources {
  cpu: {
    request: string
    limit: string
  }
  memory: {
    request: string
    limit: string
  }
  storage?: {
    request: string
    limit: string
  }
}

export interface ContainerConfig {
  name: string
  image: string
  command?: string[]
  args?: string[]
  env: Record<string, string>
  workingDir?: string
  restartPolicy: 'Always' | 'OnFailure' | 'Never'
  ports: Omit<ContainerPort, 'servicePort'>[]
  volumes: Omit<ContainerVolume, 'source'>[]
  resources: ContainerResources
  labels: Record<string, string>
  annotations: Record<string, string>
}

interface ContainerState {
  containers: Container[]
  selectedContainer: Container | null
  configForms: Record<string, ContainerConfig>
  loading: boolean
  error: string | null
  filter: {
    namespace: string
    status: string
    search: string
  }
  pagination: {
    page: number
    pageSize: number
    total: number
  }
}

const initialState: ContainerState = {
  containers: [],
  selectedContainer: null,
  configForms: {},
  loading: false,
  error: null,
  filter: {
    namespace: 'default',
    status: '',
    search: '',
  },
  pagination: {
    page: 1,
    pageSize: 20,
    total: 0,
  },
}

const containerSlice = createSlice({
  name: 'containers',
  initialState,
  reducers: {
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload
    },
    setContainers: (state, action: PayloadAction<Container[]>) => {
      state.containers = action.payload
    },
    addContainer: (state, action: PayloadAction<Container>) => {
      state.containers.push(action.payload)
    },
    updateContainer: (state, action: PayloadAction<{ id: string; updates: Partial<Container> }>) => {
      const { id, updates } = action.payload
      const index = state.containers.findIndex(c => c.id === id)
      if (index !== -1) {
        state.containers[index] = { ...state.containers[index], ...updates }
      }
    },
    removeContainer: (state, action: PayloadAction<string>) => {
      state.containers = state.containers.filter(c => c.id !== action.payload)
    },
    setSelectedContainer: (state, action: PayloadAction<Container | null>) => {
      state.selectedContainer = action.payload
    },
    setConfigForm: (state, action: PayloadAction<{ formId: string; config: ContainerConfig }>) => {
      const { formId, config } = action.payload
      state.configForms[formId] = config
    },
    updateConfigForm: (state, action: PayloadAction<{ formId: string; updates: Partial<ContainerConfig> }>) => {
      const { formId, updates } = action.payload
      if (state.configForms[formId]) {
        state.configForms[formId] = { ...state.configForms[formId], ...updates }
      }
    },
    removeConfigForm: (state, action: PayloadAction<string>) => {
      delete state.configForms[action.payload]
    },
    setFilter: (state, action: PayloadAction<Partial<ContainerState['filter']>>) => {
      state.filter = { ...state.filter, ...action.payload }
    },
    setPagination: (state, action: PayloadAction<Partial<ContainerState['pagination']>>) => {
      state.pagination = { ...state.pagination, ...action.payload }
    },
    resetState: (state) => {
      Object.assign(state, initialState)
    },
  },
})

export const {
  setLoading,
  setError,
  setContainers,
  addContainer,
  updateContainer,
  removeContainer,
  setSelectedContainer,
  setConfigForm,
  updateConfigForm,
  removeConfigForm,
  setFilter,
  setPagination,
  resetState,
} = containerSlice.actions

export default containerSlice.reducer
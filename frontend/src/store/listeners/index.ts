import { setupListeners as setupReduxListeners } from '@reduxjs/toolkit/query'
import { store } from '../index'

// 设置 Redux 监听器
export const setupListeners = () => {
  return setupReduxListeners(store.dispatch)
}

// 导出配置
export default setupListeners
import { combineReducers } from '@reduxjs/toolkit'
import authSlice from '../slices/authSlice'
import containerSlice from '../slices/containerSlice'
import uiSlice from '../slices/uiSlice'
import dashboardSlice from '../slices/dashboardSlice'

export const rootReducer = combineReducers({
  auth: authSlice,
  containers: containerSlice,
  ui: uiSlice,
  dashboard: dashboardSlice,
})

export type RootState = ReturnType<typeof rootReducer>
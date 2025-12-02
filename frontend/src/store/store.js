import { configureStore } from '@reduxjs/toolkit'
import authReducer from './slices/authSlice'
import transactionReducer from './slices/transactionSlice'
import accountReducer from './slices/accountSlice'
import categoryReducer from './slices/categorySlice'
import analyticsReducer from './slices/analyticsSlice'
import uiReducer from './slices/uiSlice'

export const store = configureStore({
	reducer: {
		auth: authReducer,
		transactions: transactionReducer,
		accounts: accountReducer,
		categories: categoryReducer,
		analytics: analyticsReducer,
		ui: uiReducer,
	},
	middleware: getDefaultMiddleware =>
		getDefaultMiddleware({
			serializableCheck: {
				ignoredActions: ['auth/setUser'],
				ignoredPaths: ['auth.user'],
			},
		}),
})

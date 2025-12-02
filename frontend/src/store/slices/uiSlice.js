import { createSlice } from '@reduxjs/toolkit'

const initialState = {
	sidebarOpen: true,
	theme: localStorage.getItem('theme') || 'light',
	notifications: [],
}

const uiSlice = createSlice({
	name: 'ui',
	initialState,
	reducers: {
		toggleSidebar: state => {
			state.sidebarOpen = !state.sidebarOpen
		},
		setSidebarOpen: (state, action) => {
			state.sidebarOpen = action.payload
		},
		setTheme: (state, action) => {
			state.theme = action.payload
		},
		addNotification: (state, action) => {
			state.notifications.push({
				id: Date.now(),
				...action.payload,
			})
		},
		removeNotification: (state, action) => {
			state.notifications = state.notifications.filter(
				n => n.id !== action.payload
			)
		},
		clearNotifications: state => {
			state.notifications = []
		},
	},
})

export const {
	toggleSidebar,
	setSidebarOpen,
	setTheme,
	addNotification,
	removeNotification,
	clearNotifications,
} = uiSlice.actions

export default uiSlice.reducer

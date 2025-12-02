import axios from 'axios'

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8082'
const AUTH_URL = process.env.REACT_APP_AUTH_URL || 'http://localhost:8081'
const ANALYTICS_URL =
	process.env.REACT_APP_ANALYTICS_URL || 'http://localhost:8083'

// Create axios instances
export const authAPI = axios.create({
	baseURL: AUTH_URL,
	headers: {
		'Content-Type': 'application/json',
	},
})

export const apiClient = axios.create({
	baseURL: API_URL,
	headers: {
		'Content-Type': 'application/json',
	},
})

export const analyticsAPI = axios.create({
	baseURL: ANALYTICS_URL,
	headers: {
		'Content-Type': 'application/json',
	},
})

// Request interceptor to add token
const requestInterceptor = config => {
	const token = localStorage.getItem('token')
	if (token) {
		config.headers.Authorization = `Bearer ${token}`
	}
	return config
}

// Response interceptor for error handling
const responseInterceptor = response => response

const errorInterceptor = error => {
	if (error.response?.status === 401) {
		localStorage.removeItem('token')
		window.location.href = '/login'
	}
	return Promise.reject(error)
}

// Apply interceptors
authAPI.interceptors.request.use(requestInterceptor)
authAPI.interceptors.response.use(responseInterceptor, errorInterceptor)

apiClient.interceptors.request.use(requestInterceptor)
apiClient.interceptors.response.use(responseInterceptor, errorInterceptor)

analyticsAPI.interceptors.request.use(requestInterceptor)
analyticsAPI.interceptors.response.use(responseInterceptor, errorInterceptor)

export default apiClient

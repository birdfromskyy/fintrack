import { authAPI } from './api'

const authService = {
	register: data => {
		return authAPI.post('/api/v1/auth/register', data)
	},

	login: data => {
		return authAPI.post('/api/v1/auth/login', data)
	},

	logout: () => {
		return authAPI.post('/api/v1/auth/logout')
	},

	verifyEmail: data => {
		return authAPI.post('/api/v1/auth/verify-email', data)
	},

	resendVerificationCode: email => {
		return authAPI.post('/api/v1/auth/resend-code', { email })
	},

	getCurrentUser: () => {
		return authAPI.get('/api/v1/auth/me')
	},
	changePassword: data => {
		return authAPI.post('/api/v1/auth/change-password', data)
	},
}

export default authService

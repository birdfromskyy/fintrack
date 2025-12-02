import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import authService from '../../services/authService'

const initialState = {
	user: null,
	token: localStorage.getItem('token'),
	isAuthenticated: false,
	isLoading: true,
	error: null,
	verificationSent: false,
}

export const register = createAsyncThunk(
	'auth/register',
	async (userData, { rejectWithValue }) => {
		try {
			const response = await authService.register(userData)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Registration failed'
			)
		}
	}
)

export const login = createAsyncThunk(
	'auth/login',
	async (credentials, { rejectWithValue }) => {
		try {
			const response = await authService.login(credentials)
			return response.data
		} catch (error) {
			return rejectWithValue(error.response?.data?.error || 'Login failed')
		}
	}
)

export const verifyEmail = createAsyncThunk(
	'auth/verifyEmail',
	async (data, { rejectWithValue }) => {
		try {
			const response = await authService.verifyEmail(data)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Verification failed'
			)
		}
	}
)

export const resendVerificationCode = createAsyncThunk(
	'auth/resendCode',
	async (email, { rejectWithValue }) => {
		try {
			const response = await authService.resendVerificationCode(email)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to resend code'
			)
		}
	}
)

export const logout = createAsyncThunk(
	'auth/logout',
	async (_, { rejectWithValue }) => {
		try {
			await authService.logout()
			return null
		} catch (error) {
			return rejectWithValue(error.response?.data?.error || 'Logout failed')
		}
	}
)

export const checkAuth = createAsyncThunk(
	'auth/checkAuth',
	async (_, { rejectWithValue }) => {
		try {
			const token = localStorage.getItem('token')
			if (!token) {
				return null
			}
			const response = await authService.getCurrentUser()
			return response.data
		} catch (error) {
			localStorage.removeItem('token')
			return rejectWithValue(
				error.response?.data?.error || 'Authentication check failed'
			)
		}
	}
)

const authSlice = createSlice({
	name: 'auth',
	initialState,
	reducers: {
		clearError: state => {
			state.error = null
		},
		setVerificationSent: (state, action) => {
			state.verificationSent = action.payload
		},
	},
	extraReducers: builder => {
		// Register
		builder.addCase(register.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(register.fulfilled, (state, action) => {
			state.isLoading = false
			state.verificationSent = true
		})
		builder.addCase(register.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Login
		builder.addCase(login.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(login.fulfilled, (state, action) => {
			state.isLoading = false
			state.isAuthenticated = true
			state.user = action.payload.user
			state.token = action.payload.token
			localStorage.setItem('token', action.payload.token)
		})
		builder.addCase(login.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Verify Email
		builder.addCase(verifyEmail.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(verifyEmail.fulfilled, state => {
			state.isLoading = false
			state.verificationSent = false
		})
		builder.addCase(verifyEmail.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Resend Code
		builder.addCase(resendVerificationCode.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(resendVerificationCode.fulfilled, state => {
			state.isLoading = false
			state.verificationSent = true
		})
		builder.addCase(resendVerificationCode.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Logout
		builder.addCase(logout.fulfilled, state => {
			state.user = null
			state.token = null
			state.isAuthenticated = false
			localStorage.removeItem('token')
		})

		// Check Auth
		builder.addCase(checkAuth.pending, state => {
			state.isLoading = true
		})
		builder.addCase(checkAuth.fulfilled, (state, action) => {
			state.isLoading = false
			if (action.payload && action.payload.user) {
				state.isAuthenticated = true
				state.user = action.payload.user
			} else {
				state.isAuthenticated = false
				state.user = null
				state.token = null
			}
		})
		builder.addCase(checkAuth.rejected, state => {
			state.isLoading = false
			state.isAuthenticated = false
			state.user = null
			state.token = null
		})
	},
})

export const { clearError, setVerificationSent } = authSlice.actions
export default authSlice.reducer

import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import analyticsService from '../../services/analyticsService'

const initialState = {
	overview: null,
	trends: [],
	forecast: null,
	insights: [],
	cashflow: [],
	isLoading: false,
	error: null,
}

export const fetchOverview = createAsyncThunk(
	'analytics/fetchOverview',
	async (period = 'month', { rejectWithValue }) => {
		try {
			const response = await analyticsService.getOverview(period)
			console.log('Overview response:', response.data)
			return response.data
		} catch (error) {
			console.error('Overview error:', error)
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch overview'
			)
		}
	}
)

export const fetchTrends = createAsyncThunk(
	'analytics/fetchTrends',
	async (days = 30, { rejectWithValue }) => {
		try {
			const response = await analyticsService.getTrends(days)
			console.log('Trends response:', response.data)
			return response.data
		} catch (error) {
			console.error('Trends error:', error)
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch trends'
			)
		}
	}
)

export const fetchForecast = createAsyncThunk(
	'analytics/fetchForecast',
	async (months = 3, { rejectWithValue }) => {
		try {
			const response = await analyticsService.getForecast(months)
			console.log('Forecast response:', response.data)
			return response.data
		} catch (error) {
			console.error('Forecast error:', error)
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch forecast'
			)
		}
	}
)

export const fetchInsights = createAsyncThunk(
	'analytics/fetchInsights',
	async (_, { rejectWithValue }) => {
		try {
			const response = await analyticsService.getInsights()
			console.log('Insights response:', response.data)
			return response.data
		} catch (error) {
			console.error('Insights error:', error)
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch insights'
			)
		}
	}
)

export const fetchCashflow = createAsyncThunk(
	'analytics/fetchCashflow',
	async ({ startDate, endDate }, { rejectWithValue }) => {
		try {
			const response = await analyticsService.getCashflow(startDate, endDate)
			console.log('Cashflow response:', response.data)
			return response.data
		} catch (error) {
			console.error('Cashflow error:', error)
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch cashflow'
			)
		}
	}
)

export const exportTransactions = createAsyncThunk(
	'analytics/exportTransactions',
	async (params, { rejectWithValue }) => {
		try {
			const response = await analyticsService.exportTransactions(params)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to export transactions'
			)
		}
	}
)

const analyticsSlice = createSlice({
	name: 'analytics',
	initialState,
	reducers: {
		clearError: state => {
			state.error = null
		},
	},
	extraReducers: builder => {
		// Overview
		builder.addCase(fetchOverview.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchOverview.fulfilled, (state, action) => {
			state.isLoading = false
			console.log('Overview stored in Redux:', action.payload)
			state.overview = action.payload
		})
		builder.addCase(fetchOverview.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
			console.error('Overview rejected:', action.payload)
		})

		// Trends
		builder.addCase(fetchTrends.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchTrends.fulfilled, (state, action) => {
			state.isLoading = false
			console.log('Trends stored in Redux:', action.payload)
			state.trends = action.payload || []
		})
		builder.addCase(fetchTrends.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
			console.error('Trends rejected:', action.payload)
		})

		// Forecast
		builder.addCase(fetchForecast.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchForecast.fulfilled, (state, action) => {
			state.isLoading = false
			console.log('Forecast stored in Redux:', action.payload)
			state.forecast = action.payload
		})
		builder.addCase(fetchForecast.rejected, (state, action) => {
			state.isLoading = false
			// Don't set error for forecast, just log it
			console.log('Forecast error (ignored):', action.payload)
			state.forecast = null
		})

		// Insights
		builder.addCase(fetchInsights.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchInsights.fulfilled, (state, action) => {
			state.isLoading = false
			console.log('Insights stored in Redux:', action.payload)
			state.insights = action.payload || []
		})
		builder.addCase(fetchInsights.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
			console.error('Insights rejected:', action.payload)
		})

		// Cashflow
		builder.addCase(fetchCashflow.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchCashflow.fulfilled, (state, action) => {
			state.isLoading = false
			console.log('Cashflow stored in Redux:', action.payload)
			state.cashflow = action.payload || []
		})
		builder.addCase(fetchCashflow.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
			console.error('Cashflow rejected:', action.payload)
		})

		// Export
		builder.addCase(exportTransactions.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(exportTransactions.fulfilled, state => {
			state.isLoading = false
		})
		builder.addCase(exportTransactions.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})
	},
})

export const { clearError } = analyticsSlice.actions
export default analyticsSlice.reducer

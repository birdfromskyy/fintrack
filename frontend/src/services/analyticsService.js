import { analyticsAPI } from './api'

const analyticsService = {
	getOverview: (period = 'month') => {
		return analyticsAPI.get('/api/v1/analytics/overview', {
			params: { period },
		})
	},

	getTrends: (days = 30) => {
		return analyticsAPI.get('/api/v1/analytics/trends', { params: { days } })
	},

	getForecast: (months = 3) => {
		return analyticsAPI.get('/api/v1/analytics/forecast', {
			params: { months },
		})
	},

	getInsights: () => {
		return analyticsAPI.get('/api/v1/analytics/insights')
	},

	getCashflow: (startDate, endDate) => {
		return analyticsAPI.get('/api/v1/analytics/cashflow', {
			params: {
				start_date: startDate,
				end_date: endDate,
			},
		})
	},

	exportTransactions: params => {
		return analyticsAPI.get('/api/v1/export/transactions', {
			params,
			responseType: 'blob',
		})
	},

	exportSummary: (startDate, endDate) => {
		return analyticsAPI.get('/api/v1/export/summary', {
			params: {
				start_date: startDate,
				end_date: endDate,
			},
			responseType: 'blob',
		})
	},

	generateReport: (period = 'month') => {
		return analyticsAPI.get('/api/v1/export/report', { params: { period } })
	},
}

export default analyticsService

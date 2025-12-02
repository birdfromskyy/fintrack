import apiClient from './api'

const transactionService = {
	getTransactions: (params = {}) => {
		return apiClient.get('/api/v1/transactions', { params })
	},

	getTransaction: id => {
		return apiClient.get(`/api/v1/transactions/${id}`)
	},

	createTransaction: data => {
		return apiClient.post('/api/v1/transactions', data)
	},

	updateTransaction: (id, data) => {
		return apiClient.put(`/api/v1/transactions/${id}`, data)
	},

	deleteTransaction: id => {
		return apiClient.delete(`/api/v1/transactions/${id}`)
	},
}

export default transactionService

import apiClient from './api'

const accountService = {
	getAccounts: () => {
		return apiClient.get('/api/v1/accounts')
	},

	getAccount: id => {
		return apiClient.get(`/api/v1/accounts/${id}`)
	},

	createAccount: data => {
		return apiClient.post('/api/v1/accounts', data)
	},

	updateAccount: (id, data) => {
		return apiClient.put(`/api/v1/accounts/${id}`, data)
	},

	deleteAccount: id => {
		return apiClient.delete(`/api/v1/accounts/${id}`)
	},

	setDefaultAccount: id => {
		return apiClient.post(`/api/v1/accounts/${id}/set-default`)
	},
}

export default accountService

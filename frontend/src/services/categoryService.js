import apiClient from './api'

const categoryService = {
	getCategories: (type = null) => {
		const params = type ? { type } : {}
		return apiClient.get('/api/v1/categories', { params })
	},

	getCategory: id => {
		return apiClient.get(`/api/v1/categories/${id}`)
	},

	createCategory: data => {
		return apiClient.post('/api/v1/categories', data)
	},

	updateCategory: (id, data) => {
		return apiClient.put(`/api/v1/categories/${id}`, data)
	},

	deleteCategory: id => {
		return apiClient.delete(`/api/v1/categories/${id}`)
	},
}

export default categoryService

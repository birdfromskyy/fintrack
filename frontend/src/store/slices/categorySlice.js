import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import categoryService from '../../services/categoryService'

const initialState = {
	categories: [],
	incomeCategories: [],
	expenseCategories: [],
	currentCategory: null,
	isLoading: false,
	error: null,
}

export const fetchCategories = createAsyncThunk(
	'categories/fetchAll',
	async (_, { rejectWithValue }) => {
		try {
			const response = await categoryService.getCategories()
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch categories'
			)
		}
	}
)

export const fetchCategory = createAsyncThunk(
	'categories/fetchOne',
	async (id, { rejectWithValue }) => {
		try {
			const response = await categoryService.getCategory(id)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch category'
			)
		}
	}
)

export const createCategory = createAsyncThunk(
	'categories/create',
	async (data, { rejectWithValue }) => {
		try {
			const response = await categoryService.createCategory(data)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to create category'
			)
		}
	}
)

export const updateCategory = createAsyncThunk(
	'categories/update',
	async ({ id, data }, { rejectWithValue }) => {
		try {
			const response = await categoryService.updateCategory(id, data)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to update category'
			)
		}
	}
)

export const deleteCategory = createAsyncThunk(
	'categories/delete',
	async (id, { rejectWithValue }) => {
		try {
			await categoryService.deleteCategory(id)
			return id
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to delete category'
			)
		}
	}
)

const categorySlice = createSlice({
	name: 'categories',
	initialState,
	reducers: {
		clearError: state => {
			state.error = null
		},
	},
	extraReducers: builder => {
		// Fetch all
		builder.addCase(fetchCategories.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchCategories.fulfilled, (state, action) => {
			state.isLoading = false
			state.categories = action.payload.categories || []
			state.incomeCategories = state.categories.filter(c => c.type === 'income')
			state.expenseCategories = state.categories.filter(
				c => c.type === 'expense'
			)
		})
		builder.addCase(fetchCategories.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Fetch one
		builder.addCase(fetchCategory.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchCategory.fulfilled, (state, action) => {
			state.isLoading = false
			state.currentCategory = action.payload.category
		})
		builder.addCase(fetchCategory.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Create
		builder.addCase(createCategory.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(createCategory.fulfilled, (state, action) => {
			state.isLoading = false
			const category = action.payload.category
			state.categories.push(category)
			if (category.type === 'income') {
				state.incomeCategories.push(category)
			} else {
				state.expenseCategories.push(category)
			}
		})
		builder.addCase(createCategory.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Update
		builder.addCase(updateCategory.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(updateCategory.fulfilled, (state, action) => {
			state.isLoading = false
			const updated = action.payload.category
			const index = state.categories.findIndex(c => c.id === updated.id)
			if (index !== -1) {
				state.categories[index] = updated
				state.incomeCategories = state.categories.filter(
					c => c.type === 'income'
				)
				state.expenseCategories = state.categories.filter(
					c => c.type === 'expense'
				)
			}
		})
		builder.addCase(updateCategory.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Delete
		builder.addCase(deleteCategory.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(deleteCategory.fulfilled, (state, action) => {
			state.isLoading = false
			state.categories = state.categories.filter(c => c.id !== action.payload)
			state.incomeCategories = state.categories.filter(c => c.type === 'income')
			state.expenseCategories = state.categories.filter(
				c => c.type === 'expense'
			)
		})
		builder.addCase(deleteCategory.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})
	},
})

export const { clearError } = categorySlice.actions
export default categorySlice.reducer

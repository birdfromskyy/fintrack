import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import transactionService from '../../services/transactionService'

const initialState = {
	transactions: [],
	currentTransaction: null,
	isLoading: false,
	error: null,
	filters: {
		accountId: null,
		categoryId: null,
		type: null,
		dateFrom: null,
		dateTo: null,
	},
	pagination: {
		page: 1,
		limit: 50,
		total: 0,
	},
}

export const fetchTransactions = createAsyncThunk(
	'transactions/fetchAll',
	async (params, { rejectWithValue }) => {
		try {
			const response = await transactionService.getTransactions(params)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch transactions'
			)
		}
	}
)

export const fetchTransaction = createAsyncThunk(
	'transactions/fetchOne',
	async (id, { rejectWithValue }) => {
		try {
			const response = await transactionService.getTransaction(id)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch transaction'
			)
		}
	}
)

export const createTransaction = createAsyncThunk(
	'transactions/create',
	async (data, { rejectWithValue }) => {
		try {
			const response = await transactionService.createTransaction(data)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to create transaction'
			)
		}
	}
)

export const updateTransaction = createAsyncThunk(
	'transactions/update',
	async ({ id, data }, { rejectWithValue }) => {
		try {
			const response = await transactionService.updateTransaction(id, data)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to update transaction'
			)
		}
	}
)

export const deleteTransaction = createAsyncThunk(
	'transactions/delete',
	async (id, { rejectWithValue }) => {
		try {
			await transactionService.deleteTransaction(id)
			return id
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to delete transaction'
			)
		}
	}
)

const transactionSlice = createSlice({
	name: 'transactions',
	initialState,
	reducers: {
		setFilters: (state, action) => {
			state.filters = { ...state.filters, ...action.payload }
		},
		clearFilters: state => {
			state.filters = initialState.filters
		},
		setPagination: (state, action) => {
			state.pagination = { ...state.pagination, ...action.payload }
		},
		clearError: state => {
			state.error = null
		},
	},
	extraReducers: builder => {
		// Fetch all
		builder.addCase(fetchTransactions.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchTransactions.fulfilled, (state, action) => {
			state.isLoading = false
			state.transactions = action.payload.transactions || []
			state.pagination.total = action.payload.count || 0
		})
		builder.addCase(fetchTransactions.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Fetch one
		builder.addCase(fetchTransaction.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchTransaction.fulfilled, (state, action) => {
			state.isLoading = false
			state.currentTransaction = action.payload.transaction
		})
		builder.addCase(fetchTransaction.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Create
		builder.addCase(createTransaction.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(createTransaction.fulfilled, (state, action) => {
			state.isLoading = false
			state.transactions.unshift(action.payload.transaction)
		})
		builder.addCase(createTransaction.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Update
		builder.addCase(updateTransaction.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(updateTransaction.fulfilled, (state, action) => {
			state.isLoading = false
			const index = state.transactions.findIndex(
				t => t.id === action.payload.transaction.id
			)
			if (index !== -1) {
				state.transactions[index] = action.payload.transaction
			}
		})
		builder.addCase(updateTransaction.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Delete
		builder.addCase(deleteTransaction.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(deleteTransaction.fulfilled, (state, action) => {
			state.isLoading = false
			state.transactions = state.transactions.filter(
				t => t.id !== action.payload
			)
		})
		builder.addCase(deleteTransaction.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})
	},
})

export const { setFilters, clearFilters, setPagination, clearError } =
	transactionSlice.actions
export default transactionSlice.reducer

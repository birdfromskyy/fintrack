import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import accountService from '../../services/accountService'

const initialState = {
	accounts: [],
	currentAccount: null,
	defaultAccountId: null,
	isLoading: false,
	error: null,
}

export const fetchAccounts = createAsyncThunk(
	'accounts/fetchAll',
	async (_, { rejectWithValue }) => {
		try {
			const response = await accountService.getAccounts()
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch accounts'
			)
		}
	}
)

export const fetchAccount = createAsyncThunk(
	'accounts/fetchOne',
	async (id, { rejectWithValue }) => {
		try {
			const response = await accountService.getAccount(id)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to fetch account'
			)
		}
	}
)

export const createAccount = createAsyncThunk(
	'accounts/create',
	async (data, { rejectWithValue }) => {
		try {
			const response = await accountService.createAccount(data)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to create account'
			)
		}
	}
)

export const updateAccount = createAsyncThunk(
	'accounts/update',
	async ({ id, data }, { rejectWithValue }) => {
		try {
			const response = await accountService.updateAccount(id, data)
			return response.data
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to update account'
			)
		}
	}
)

export const deleteAccount = createAsyncThunk(
	'accounts/delete',
	async (id, { rejectWithValue }) => {
		try {
			await accountService.deleteAccount(id)
			return id
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to delete account'
			)
		}
	}
)

export const setDefaultAccount = createAsyncThunk(
	'accounts/setDefault',
	async (id, { rejectWithValue }) => {
		try {
			await accountService.setDefaultAccount(id)
			return id
		} catch (error) {
			return rejectWithValue(
				error.response?.data?.error || 'Failed to set default account'
			)
		}
	}
)

const accountSlice = createSlice({
	name: 'accounts',
	initialState,
	reducers: {
		clearError: state => {
			state.error = null
		},
		setCurrentAccount: (state, action) => {
			state.currentAccount = action.payload
		},
	},
	extraReducers: builder => {
		// Fetch all
		builder.addCase(fetchAccounts.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchAccounts.fulfilled, (state, action) => {
			state.isLoading = false
			state.accounts = action.payload.accounts || []
			const defaultAccount = state.accounts.find(a => a.is_default)
			state.defaultAccountId = defaultAccount?.id || null
		})
		builder.addCase(fetchAccounts.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Fetch one
		builder.addCase(fetchAccount.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(fetchAccount.fulfilled, (state, action) => {
			state.isLoading = false
			state.currentAccount = action.payload.account
		})
		builder.addCase(fetchAccount.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Create
		builder.addCase(createAccount.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(createAccount.fulfilled, (state, action) => {
			state.isLoading = false
			state.accounts.push(action.payload.account)
			if (action.payload.account.is_default) {
				state.defaultAccountId = action.payload.account.id
				state.accounts = state.accounts.map(a => ({
					...a,
					is_default: a.id === action.payload.account.id,
				}))
			}
		})
		builder.addCase(createAccount.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Update
		builder.addCase(updateAccount.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(updateAccount.fulfilled, (state, action) => {
			state.isLoading = false
			const index = state.accounts.findIndex(
				a => a.id === action.payload.account.id
			)
			if (index !== -1) {
				state.accounts[index] = action.payload.account
			}
		})
		builder.addCase(updateAccount.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Delete
		builder.addCase(deleteAccount.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(deleteAccount.fulfilled, (state, action) => {
			state.isLoading = false
			state.accounts = state.accounts.filter(a => a.id !== action.payload)
		})
		builder.addCase(deleteAccount.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})

		// Set default
		builder.addCase(setDefaultAccount.pending, state => {
			state.isLoading = true
			state.error = null
		})
		builder.addCase(setDefaultAccount.fulfilled, (state, action) => {
			state.isLoading = false
			state.defaultAccountId = action.payload
			state.accounts = state.accounts.map(a => ({
				...a,
				is_default: a.id === action.payload,
			}))
		})
		builder.addCase(setDefaultAccount.rejected, (state, action) => {
			state.isLoading = false
			state.error = action.payload
		})
	},
})

export const { clearError, setCurrentAccount } = accountSlice.actions
export default accountSlice.reducer

import React, { useState, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	Button,
	TextField,
	FormControl,
	InputLabel,
	Select,
	MenuItem,
	Box,
	Typography,
	ToggleButton,
	ToggleButtonGroup,
	InputAdornment,
	Alert,
} from '@mui/material'
import { ArrowUpward, ArrowDownward } from '@mui/icons-material'
import { DatePicker } from '@mui/x-date-pickers/DatePicker'

import {
	createTransaction,
	updateTransaction,
} from '../../store/slices/transactionSlice'
import { fetchOverview } from '../../store/slices/analyticsSlice'
import { validateAmount } from '../../utils/validators'

const TransactionDialog = ({ open, onClose, transaction, onSave }) => {
	const dispatch = useDispatch()
	const { accounts } = useSelector(state => state.accounts)
	const { categories } = useSelector(state => state.categories)
	const { isLoading } = useSelector(state => state.transactions)

	const [formData, setFormData] = useState({
		type: 'expense',
		account_id: '',
		category_id: '',
		amount: '',
		description: '',
		date: new Date(),
	})
	const [error, setError] = useState('')

	useEffect(() => {
		if (transaction) {
			setFormData({
				type: transaction.type,
				account_id: transaction.account_id,
				category_id: transaction.category_id,
				amount: transaction.amount.toString(),
				description: transaction.description || '',
				date: new Date(transaction.date),
			})
		} else {
			// Set default account
			const defaultAccount = accounts.find(a => a.is_default)
			setFormData(prev => ({
				...prev,
				account_id: defaultAccount?.id || accounts[0]?.id || '',
			}))
		}
	}, [transaction, accounts])

	const handleTypeChange = (event, newType) => {
		if (newType !== null) {
			setFormData({
				...formData,
				type: newType,
				category_id: '', // Reset category when type changes
			})
		}
	}

	const handleChange = field => event => {
		setFormData({
			...formData,
			[field]: event.target.value,
		})
		setError('')
	}

	const handleDateChange = date => {
		setFormData({
			...formData,
			date: date,
		})
	}

	const handleSubmit = async () => {
		// Validation
		if (!formData.account_id) {
			setError('Выберите счёт')
			return
		}
		if (!formData.category_id) {
			setError('Выберите категорию')
			return
		}
		if (!validateAmount(formData.amount)) {
			setError('Введите корректную сумму')
			return
		}
		if (!formData.date) {
			setError('Выберите дату')
			return
		}

		const data = {
			account_id: formData.account_id,
			category_id: formData.category_id,
			amount: parseFloat(formData.amount),
			description: formData.description,
			date:
				formData.date instanceof Date
					? formData.date.toISOString().split('T')[0]
					: formData.date, // Already a string in YYYY-MM-DD format
		}

		let result
		if (transaction) {
			result = await dispatch(updateTransaction({ id: transaction.id, data }))
		} else {
			result = await dispatch(createTransaction(data))
		}

		if (
			createTransaction.fulfilled.match(result) ||
			updateTransaction.fulfilled.match(result)
		) {
			onSave()
			handleClose()
			// Trigger analytics refresh
			dispatch(fetchOverview('month'))
		}
	}

	const handleClose = () => {
		setFormData({
			type: 'expense',
			account_id: '',
			category_id: '',
			amount: '',
			description: '',
			date: new Date(),
		})
		setError('')
		onClose()
	}

	const filteredCategories = categories.filter(c => c.type === formData.type)

	return (
		<Dialog open={open} onClose={handleClose} maxWidth='sm' fullWidth>
			<DialogTitle>
				{transaction ? 'Редактировать транзакцию' : 'Новая транзакция'}
			</DialogTitle>

			<DialogContent>
				{error && (
					<Alert severity='error' sx={{ mb: 2 }}>
						{error}
					</Alert>
				)}

				<Box sx={{ mb: 3, mt: 2 }}>
					<ToggleButtonGroup
						value={formData.type}
						exclusive
						onChange={handleTypeChange}
						fullWidth
					>
						<ToggleButton value='income' color='success'>
							<ArrowUpward sx={{ mr: 1 }} />
							Доход
						</ToggleButton>
						<ToggleButton value='expense' color='error'>
							<ArrowDownward sx={{ mr: 1 }} />
							Расход
						</ToggleButton>
					</ToggleButtonGroup>
				</Box>

				<FormControl fullWidth sx={{ mb: 2 }}>
					<InputLabel>Счёт</InputLabel>
					<Select
						value={formData.account_id}
						onChange={handleChange('account_id')}
						label='Счёт'
					>
						{accounts.map(account => (
							<MenuItem key={account.id} value={account.id}>
								<Box display='flex' justifyContent='space-between' width='100%'>
									<span>{account.name}</span>
									<Typography variant='body2' color='text.secondary'>
										{account.balance.toFixed(2)} ₽
									</Typography>
								</Box>
							</MenuItem>
						))}
					</Select>
				</FormControl>

				<FormControl fullWidth sx={{ mb: 2 }}>
					<InputLabel>Категория</InputLabel>
					<Select
						value={formData.category_id}
						onChange={handleChange('category_id')}
						label='Категория'
					>
						{filteredCategories.map(category => (
							<MenuItem key={category.id} value={category.id}>
								<Box display='flex' alignItems='center' gap={1}>
									<span>{category.icon}</span>
									<span>{category.name}</span>
								</Box>
							</MenuItem>
						))}
					</Select>
				</FormControl>

				<TextField
					fullWidth
					label='Сумма'
					type='number'
					value={formData.amount}
					onChange={handleChange('amount')}
					sx={{ mb: 2 }}
					InputProps={{
						startAdornment: <InputAdornment position='start'>₽</InputAdornment>,
					}}
					inputProps={{
						step: 0.01,
						min: 0,
					}}
				/>

				<DatePicker
					label='Дата'
					value={formData.date}
					onChange={handleDateChange}
					slotProps={{ textField: { fullWidth: true, sx: { mb: 2 } } }}
				/>

				<TextField
					fullWidth
					label='Описание (необязательно)'
					value={formData.description}
					onChange={handleChange('description')}
					multiline
					rows={2}
				/>
			</DialogContent>

			<DialogActions>
				<Button onClick={handleClose} disabled={isLoading}>
					Отмена
				</Button>
				<Button onClick={handleSubmit} variant='contained' disabled={isLoading}>
					{transaction ? 'Сохранить' : 'Добавить'}
				</Button>
			</DialogActions>
		</Dialog>
	)
}

export default TransactionDialog

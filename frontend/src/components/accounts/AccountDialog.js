import React, { useState, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	Button,
	TextField,
	FormControlLabel,
	Switch,
	Alert,
	InputAdornment,
} from '@mui/material'

import { createAccount, updateAccount } from '../../store/slices/accountSlice'

const AccountDialog = ({ open, onClose, account, onSave }) => {
	const dispatch = useDispatch()
	const { isLoading } = useSelector(state => state.accounts)

	const [formData, setFormData] = useState({
		name: '',
		balance: '0',
		is_default: false,
	})
	const [error, setError] = useState('')

	useEffect(() => {
		if (account) {
			setFormData({
				name: account.name,
				balance: account.balance.toString(),
				is_default: account.is_default,
			})
		} else {
			setFormData({
				name: '',
				balance: '0',
				is_default: false,
			})
		}
	}, [account])

	const handleChange = field => event => {
		const value =
			field === 'is_default' ? event.target.checked : event.target.value
		setFormData({
			...formData,
			[field]: value,
		})
		setError('')
	}

	const handleSubmit = async () => {
		// Validation
		if (!formData.name.trim()) {
			setError('Введите название счёта')
			return
		}

		const balance = parseFloat(formData.balance)
		if (isNaN(balance)) {
			setError('Введите корректный баланс')
			return
		}

		const data = {
			name: formData.name.trim(),
			balance: balance,
			...(account ? {} : { is_default: formData.is_default }),
		}

		let result
		if (account) {
			result = await dispatch(updateAccount({ id: account.id, data }))
		} else {
			result = await dispatch(createAccount(data))
		}

		if (
			createAccount.fulfilled.match(result) ||
			updateAccount.fulfilled.match(result)
		) {
			onSave()
			handleClose()
		}
	}

	const handleClose = () => {
		setFormData({
			name: '',
			balance: '0',
			is_default: false,
		})
		setError('')
		onClose()
	}

	return (
		<Dialog open={open} onClose={handleClose} maxWidth='sm' fullWidth>
			<DialogTitle>{account ? 'Редактировать счёт' : 'Новый счёт'}</DialogTitle>

			<DialogContent>
				{error && (
					<Alert severity='error' sx={{ mb: 2 }}>
						{error}
					</Alert>
				)}

				<TextField
					autoFocus
					fullWidth
					label='Название счёта'
					value={formData.name}
					onChange={handleChange('name')}
					sx={{ mb: 2 }}
					placeholder='Например: Основной счёт, Накопления...'
				/>

				<TextField
					fullWidth
					label='Начальный баланс'
					type='number'
					value={formData.balance}
					onChange={handleChange('balance')}
					sx={{ mb: 2 }}
					InputProps={{
						startAdornment: <InputAdornment position='start'>₽</InputAdornment>,
					}}
					inputProps={{
						step: 0.01,
					}}
					helperText={
						account ? 'Изменение баланса не влияет на историю транзакций' : ''
					}
				/>

				{!account && (
					<FormControlLabel
						control={
							<Switch
								checked={formData.is_default}
								onChange={handleChange('is_default')}
							/>
						}
						label='Сделать основным счётом'
					/>
				)}
			</DialogContent>

			<DialogActions>
				<Button onClick={handleClose} disabled={isLoading}>
					Отмена
				</Button>
				<Button onClick={handleSubmit} variant='contained' disabled={isLoading}>
					{account ? 'Сохранить' : 'Создать'}
				</Button>
			</DialogActions>
		</Dialog>
	)
}

export default AccountDialog

import React, { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import {
	Paper,
	TextField,
	Button,
	Typography,
	Box,
	Alert,
	CircularProgress,
} from '@mui/material'
import { register, clearError } from '../../store/slices/authSlice'

const Register = () => {
	const navigate = useNavigate()
	const dispatch = useDispatch()
	const { isLoading, error, verificationSent } = useSelector(
		state => state.auth
	)

	const [formData, setFormData] = useState({
		email: '',
		password: '',
		confirmPassword: '',
	})
	const [validationError, setValidationError] = useState('')

	useEffect(() => {
		dispatch(clearError())
	}, [dispatch])

	useEffect(() => {
		if (verificationSent) {
			navigate('/verify-email', { state: { email: formData.email } })
		}
	}, [verificationSent, navigate, formData.email])

	const handleChange = e => {
		setFormData({
			...formData,
			[e.target.name]: e.target.value,
		})
		setValidationError('')
	}

	const handleSubmit = async e => {
		e.preventDefault()

		if (formData.password !== formData.confirmPassword) {
			setValidationError('Пароли не совпадают')
			return
		}

		if (formData.password.length < 6) {
			setValidationError('Пароль должен содержать минимум 6 символов')
			return
		}

		const result = await dispatch(
			register({
				email: formData.email,
				password: formData.password,
			})
		)

		if (register.fulfilled.match(result)) {
			// Navigation handled by useEffect
		}
	}

	return (
		<Paper elevation={3} sx={{ p: 4 }}>
			<Box sx={{ mb: 3, textAlign: 'center' }}>
				<Typography variant='h4' component='h1' gutterBottom>
					Регистрация
				</Typography>
				<Typography variant='body2' color='text.secondary'>
					Создайте аккаунт для начала работы
				</Typography>
			</Box>

			{(error || validationError) && (
				<Alert severity='error' sx={{ mb: 2 }}>
					{error || validationError}
				</Alert>
			)}

			<form onSubmit={handleSubmit}>
				<TextField
					fullWidth
					label='Email'
					name='email'
					type='email'
					value={formData.email}
					onChange={handleChange}
					margin='normal'
					required
					autoComplete='email'
					autoFocus
				/>

				<TextField
					fullWidth
					label='Пароль'
					name='password'
					type='password'
					value={formData.password}
					onChange={handleChange}
					margin='normal'
					required
					autoComplete='new-password'
					helperText='Минимум 6 символов'
				/>

				<TextField
					fullWidth
					label='Подтвердите пароль'
					name='confirmPassword'
					type='password'
					value={formData.confirmPassword}
					onChange={handleChange}
					margin='normal'
					required
					autoComplete='new-password'
				/>

				<Button
					type='submit'
					fullWidth
					variant='contained'
					size='large'
					sx={{ mt: 3, mb: 2 }}
					disabled={isLoading}
				>
					{isLoading ? <CircularProgress size={24} /> : 'Зарегистрироваться'}
				</Button>

				<Box sx={{ textAlign: 'center' }}>
					<Typography variant='body2'>
						Уже есть аккаунт?{' '}
						<Link to='/login' style={{ textDecoration: 'none' }}>
							Войти
						</Link>
					</Typography>
				</Box>
			</form>
		</Paper>
	)
}

export default Register

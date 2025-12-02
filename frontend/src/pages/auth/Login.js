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
import { login, clearError } from '../../store/slices/authSlice'

const Login = () => {
	const navigate = useNavigate()
	const dispatch = useDispatch()
	const { isLoading, error } = useSelector(state => state.auth)

	const [formData, setFormData] = useState({
		email: '',
		password: '',
	})

	useEffect(() => {
		dispatch(clearError())
	}, [dispatch])

	const handleChange = e => {
		setFormData({
			...formData,
			[e.target.name]: e.target.value,
		})
	}

	const handleSubmit = async e => {
		e.preventDefault()
		const result = await dispatch(login(formData))
		if (login.fulfilled.match(result)) {
			navigate('/dashboard')
		}
	}

	return (
		<Paper elevation={3} sx={{ p: 4 }}>
			<Box sx={{ mb: 3, textAlign: 'center' }}>
				<Typography variant='h4' component='h1' gutterBottom>
					Вход в Fintrack
				</Typography>
				<Typography variant='body2' color='text.secondary'>
					Введите ваши данные для входа
				</Typography>
			</Box>

			{error && (
				<Alert severity='error' sx={{ mb: 2 }}>
					{error}
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
					autoComplete='current-password'
				/>

				<Button
					type='submit'
					fullWidth
					variant='contained'
					size='large'
					sx={{ mt: 3, mb: 2 }}
					disabled={isLoading}
				>
					{isLoading ? <CircularProgress size={24} /> : 'Войти'}
				</Button>

				<Box sx={{ textAlign: 'center' }}>
					<Typography variant='body2'>
						Нет аккаунта?{' '}
						<Link to='/register' style={{ textDecoration: 'none' }}>
							Зарегистрироваться
						</Link>
					</Typography>
				</Box>
			</form>
		</Paper>
	)
}

export default Login

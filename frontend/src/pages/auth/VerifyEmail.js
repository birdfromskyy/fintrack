import React, { useState, useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
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
import {
	verifyEmail,
	resendVerificationCode,
	clearError,
} from '../../store/slices/authSlice'

const VerifyEmail = () => {
	const navigate = useNavigate()
	const location = useLocation()
	const dispatch = useDispatch()
	const { isLoading, error } = useSelector(state => state.auth)

	const [code, setCode] = useState('')
	const [email, setEmail] = useState(location.state?.email || '')
	const [resendDisabled, setResendDisabled] = useState(false)
	const [resendTimer, setResendTimer] = useState(0)

	useEffect(() => {
		dispatch(clearError())
	}, [dispatch])

	useEffect(() => {
		if (resendTimer > 0) {
			const timer = setTimeout(() => setResendTimer(resendTimer - 1), 1000)
			return () => clearTimeout(timer)
		} else {
			setResendDisabled(false)
		}
	}, [resendTimer])

	const handleSubmit = async e => {
		e.preventDefault()

		if (code.length !== 6) {
			return
		}

		const result = await dispatch(verifyEmail({ email, code }))
		if (verifyEmail.fulfilled.match(result)) {
			navigate('/login')
		}
	}

	const handleResend = async () => {
		const result = await dispatch(resendVerificationCode(email))
		if (resendVerificationCode.fulfilled.match(result)) {
			setResendDisabled(true)
			setResendTimer(60)
		}
	}

	const handleCodeChange = e => {
		const value = e.target.value.replace(/\D/g, '')
		if (value.length <= 6) {
			setCode(value)
		}
	}

	return (
		<Paper elevation={3} sx={{ p: 4 }}>
			<Box sx={{ mb: 3, textAlign: 'center' }}>
				<Typography variant='h4' component='h1' gutterBottom>
					Подтверждение Email
				</Typography>
				<Typography variant='body2' color='text.secondary'>
					Введите код, отправленный на {email}
				</Typography>
			</Box>

			{error && (
				<Alert severity='error' sx={{ mb: 2 }}>
					{error}
				</Alert>
			)}

			<form onSubmit={handleSubmit}>
				{!location.state?.email && (
					<TextField
						fullWidth
						label='Email'
						type='email'
						value={email}
						onChange={e => setEmail(e.target.value)}
						margin='normal'
						required
						autoComplete='email'
					/>
				)}

				<TextField
					fullWidth
					label='Код подтверждения'
					value={code}
					onChange={handleCodeChange}
					margin='normal'
					required
					autoFocus
					placeholder='000000'
					inputProps={{
						maxLength: 6,
						style: {
							textAlign: 'center',
							fontSize: '1.5rem',
							letterSpacing: '0.5rem',
						},
					}}
				/>

				<Button
					type='submit'
					fullWidth
					variant='contained'
					size='large'
					sx={{ mt: 3, mb: 2 }}
					disabled={isLoading || code.length !== 6}
				>
					{isLoading ? <CircularProgress size={24} /> : 'Подтвердить'}
				</Button>

				<Box sx={{ textAlign: 'center' }}>
					<Button
						onClick={handleResend}
						disabled={resendDisabled || isLoading}
						variant='text'
					>
						{resendTimer > 0
							? `Отправить повторно через ${resendTimer}с`
							: 'Отправить код повторно'}
					</Button>
				</Box>
			</form>
		</Paper>
	)
}

export default VerifyEmail

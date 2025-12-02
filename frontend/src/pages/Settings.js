import React, { useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
	Box,
	Paper,
	Typography,
	TextField,
	Button,
	Switch,
	FormControlLabel,
	Divider,
	Alert,
	Grid,
	List,
	ListItem,
	ListItemText,
	ListItemSecondaryAction,
	Select,
	MenuItem,
	FormControl,
	InputLabel,
	Card,
	CardContent,
	IconButton,
} from '@mui/material'
import {
	Save as SaveIcon,
	Download as DownloadIcon,
	Upload as UploadIcon,
	Delete as DeleteIcon,
	Security as SecurityIcon,
	Notifications as NotificationsIcon,
	Palette as PaletteIcon,
	Language as LanguageIcon,
} from '@mui/icons-material'

import { logout } from '../store/slices/authSlice'
import { setTheme } from '../store/slices/uiSlice'
import { exportTransactions } from '../store/slices/analyticsSlice'
import ConfirmDialog from '../components/common/ConfirmDialog'
import authService from '../services/authService'

const Settings = () => {
	const dispatch = useDispatch()
	const { user } = useSelector(state => state.auth)
	const { theme } = useSelector(state => state.ui)

	const [profileData, setProfileData] = useState({
		email: user?.email || '',
		currentPassword: '',
		newPassword: '',
		confirmPassword: '',
	})

	const [preferences, setPreferences] = useState({
		currency: 'RUB',
		language: 'ru',
		dateFormat: 'DD.MM.YYYY',
		startOfWeek: 'monday',
		theme: theme,
	})

	const [notifications, setNotifications] = useState({
		emailNotifications: true,
		monthlyReport: true,
		budgetAlerts: true,
		securityAlerts: true,
	})

	const [deleteAccountOpen, setDeleteAccountOpen] = useState(false)
	const [exportDataOpen, setExportDataOpen] = useState(false)
	const [successMessage, setSuccessMessage] = useState('')
	const [errorMessage, setErrorMessage] = useState('')

	const handleProfileChange = field => event => {
		setProfileData({
			...profileData,
			[field]: event.target.value,
		})
	}

	const handlePreferenceChange = field => event => {
		const value = event.target.value
		setPreferences({
			...preferences,
			[field]: value,
		})

		if (field === 'theme') {
			dispatch(setTheme(value))
		}
	}

	const handleNotificationChange = field => event => {
		setNotifications({
			...notifications,
			[field]: event.target.checked,
		})
	}

	const handleSaveProfile = async () => {
		// Validation
		if (!profileData.currentPassword && profileData.newPassword) {
			setErrorMessage('Введите текущий пароль')
			return
		}

		if (
			profileData.newPassword &&
			profileData.newPassword !== profileData.confirmPassword
		) {
			setErrorMessage('Пароли не совпадают')
			return
		}

		if (profileData.newPassword && profileData.newPassword.length < 6) {
			setErrorMessage('Пароль должен содержать минимум 6 символов')
			return
		}

		// Change password API call
		if (profileData.currentPassword && profileData.newPassword) {
			try {
				await authService.changePassword({
					current_password: profileData.currentPassword,
					new_password: profileData.newPassword,
				})
				setSuccessMessage('Пароль успешно изменен')
				setProfileData(prev => ({
					...prev,
					currentPassword: '',
					newPassword: '',
					confirmPassword: '',
				}))
			} catch (error) {
				setErrorMessage(
					error.response?.data?.error || 'Ошибка при смене пароля'
				)
			}
		}
	}

	const handleSavePreferences = () => {
		// Here would be API call to save preferences
		localStorage.setItem('preferences', JSON.stringify(preferences))
		setSuccessMessage('Настройки сохранены')
		setTimeout(() => setSuccessMessage(''), 3000)
	}

	const handleSaveNotifications = () => {
		// Here would be API call to save notification settings
		localStorage.setItem('notifications', JSON.stringify(notifications))
		setSuccessMessage('Настройки уведомлений сохранены')
		setTimeout(() => setSuccessMessage(''), 3000)
	}

	const handleExportData = async () => {
		const result = await dispatch(exportTransactions({ format: 'csv' }))
		if (exportTransactions.fulfilled.match(result)) {
			const blob = new Blob([result.payload], { type: 'text/csv' })
			const url = window.URL.createObjectURL(blob)
			const link = document.createElement('a')
			link.href = url
			link.download = `fintrack_export_${
				new Date().toISOString().split('T')[0]
			}.csv`
			link.click()
			window.URL.revokeObjectURL(url)
			setExportDataOpen(false)
			setSuccessMessage('Данные успешно экспортированы')
			setTimeout(() => setSuccessMessage(''), 3000)
		}
	}

	const handleDeleteAccount = () => {
		// Here would be API call to delete account
		dispatch(logout())
	}

	return (
		<Box>
			<Typography variant='h4' gutterBottom>
				Настройки
			</Typography>

			{successMessage && (
				<Alert
					severity='success'
					sx={{ mb: 2 }}
					onClose={() => setSuccessMessage('')}
				>
					{successMessage}
				</Alert>
			)}

			{errorMessage && (
				<Alert
					severity='error'
					sx={{ mb: 2 }}
					onClose={() => setErrorMessage('')}
				>
					{errorMessage}
				</Alert>
			)}

			<Grid container spacing={3}>
				{/* Profile Settings */}
				<Grid item xs={12} md={6}>
					<Paper sx={{ p: 3 }}>
						<Box display='flex' alignItems='center' mb={2}>
							<SecurityIcon sx={{ mr: 1 }} />
							<Typography variant='h6'>Профиль и безопасность</Typography>
						</Box>
						<Divider sx={{ mb: 2 }} />

						<TextField
							fullWidth
							label='Email'
							value={profileData.email}
							disabled
							sx={{ mb: 2 }}
						/>

						<TextField
							fullWidth
							label='Текущий пароль'
							type='password'
							value={profileData.currentPassword}
							onChange={handleProfileChange('currentPassword')}
							sx={{ mb: 2 }}
						/>

						<TextField
							fullWidth
							label='Новый пароль'
							type='password'
							value={profileData.newPassword}
							onChange={handleProfileChange('newPassword')}
							sx={{ mb: 2 }}
						/>

						<TextField
							fullWidth
							label='Подтвердите новый пароль'
							type='password'
							value={profileData.confirmPassword}
							onChange={handleProfileChange('confirmPassword')}
							sx={{ mb: 3 }}
						/>

						<Button
							variant='contained'
							startIcon={<SaveIcon />}
							onClick={handleSaveProfile}
							fullWidth
						>
							Сохранить изменения
						</Button>
					</Paper>
				</Grid>

				{/* Preferences */}
				<Grid item xs={12} md={6}>
					<Paper sx={{ p: 3 }}>
						<Box display='flex' alignItems='center' mb={2}>
							<PaletteIcon sx={{ mr: 1 }} />
							<Typography variant='h6'>Предпочтения</Typography>
						</Box>
						<Divider sx={{ mb: 2 }} />

						<FormControl fullWidth sx={{ mb: 2 }}>
							<InputLabel>Валюта</InputLabel>
							<Select
								value={preferences.currency}
								onChange={handlePreferenceChange('currency')}
								label='Валюта'
								disabled
							>
								<MenuItem value='RUB'>₽ Российский рубль</MenuItem>
							</Select>
						</FormControl>

						<FormControl fullWidth sx={{ mb: 2 }}>
							<InputLabel>Язык</InputLabel>
							<Select
								value={preferences.language}
								onChange={handlePreferenceChange('language')}
								label='Язык'
								disabled
							>
								<MenuItem value='ru'>Русский</MenuItem>
							</Select>
						</FormControl>

						<FormControl fullWidth sx={{ mb: 2 }}>
							<InputLabel>Формат даты</InputLabel>
							<Select
								value={preferences.dateFormat}
								onChange={handlePreferenceChange('dateFormat')}
								label='Формат даты'
							>
								<MenuItem value='DD.MM.YYYY'>ДД.ММ.ГГГГ</MenuItem>
								<MenuItem value='MM/DD/YYYY'>ММ/ДД/ГГГГ</MenuItem>
								<MenuItem value='YYYY-MM-DD'>ГГГГ-ММ-ДД</MenuItem>
							</Select>
						</FormControl>

						<FormControl fullWidth sx={{ mb: 2 }}>
							<InputLabel>Начало недели</InputLabel>
							<Select
								value={preferences.startOfWeek}
								onChange={handlePreferenceChange('startOfWeek')}
								label='Начало недели'
							>
								<MenuItem value='monday'>Понедельник</MenuItem>
								<MenuItem value='sunday'>Воскресенье</MenuItem>
							</Select>
						</FormControl>

						<FormControl fullWidth sx={{ mb: 3 }}>
							<InputLabel>Тема</InputLabel>
							<Select
								value={preferences.theme}
								onChange={handlePreferenceChange('theme')}
								label='Тема'
							>
								<MenuItem value='light'>Светлая</MenuItem>
								<MenuItem value='dark'>Тёмная</MenuItem>
							</Select>
						</FormControl>

						<Button
							variant='contained'
							startIcon={<SaveIcon />}
							onClick={handleSavePreferences}
							fullWidth
						>
							Сохранить настройки
						</Button>
					</Paper>
				</Grid>

				{/* Notifications */}
				<Grid item xs={12} md={6}>
					<Paper sx={{ p: 3 }}>
						<Box display='flex' alignItems='center' mb={2}>
							<NotificationsIcon sx={{ mr: 1 }} />
							<Typography variant='h6'>Уведомления</Typography>
						</Box>
						<Divider sx={{ mb: 2 }} />

						<List>
							<ListItem>
								<ListItemText
									primary='Email уведомления'
									secondary='Получать важные уведомления на email'
								/>
								<ListItemSecondaryAction>
									<Switch
										checked={notifications.emailNotifications}
										onChange={handleNotificationChange('emailNotifications')}
									/>
								</ListItemSecondaryAction>
							</ListItem>

							<ListItem>
								<ListItemText
									primary='Ежемесячный отчет'
									secondary='Получать отчет о финансах каждый месяц'
								/>
								<ListItemSecondaryAction>
									<Switch
										checked={notifications.monthlyReport}
										onChange={handleNotificationChange('monthlyReport')}
									/>
								</ListItemSecondaryAction>
							</ListItem>

							<ListItem>
								<ListItemText
									primary='Уведомления о бюджете'
									secondary='Предупреждения о превышении бюджета'
								/>
								<ListItemSecondaryAction>
									<Switch
										checked={notifications.budgetAlerts}
										onChange={handleNotificationChange('budgetAlerts')}
									/>
								</ListItemSecondaryAction>
							</ListItem>

							<ListItem>
								<ListItemText
									primary='Уведомления безопасности'
									secondary='Уведомления о входе с нового устройства'
								/>
								<ListItemSecondaryAction>
									<Switch
										checked={notifications.securityAlerts}
										onChange={handleNotificationChange('securityAlerts')}
									/>
								</ListItemSecondaryAction>
							</ListItem>
						</List>

						<Button
							variant='contained'
							startIcon={<SaveIcon />}
							onClick={handleSaveNotifications}
							fullWidth
						>
							Сохранить настройки
						</Button>
					</Paper>
				</Grid>

				{/* Data Management */}
				<Grid item xs={12} md={6}>
					<Paper sx={{ p: 3 }}>
						<Typography variant='h6' gutterBottom>
							Управление данными
						</Typography>
						<Divider sx={{ mb: 2 }} />

						<Card variant='outlined' sx={{ mb: 2 }}>
							<CardContent>
								<Box
									display='flex'
									justifyContent='space-between'
									alignItems='center'
								>
									<Box>
										<Typography variant='subtitle1'>Экспорт данных</Typography>
										<Typography variant='body2' color='text.secondary'>
											Скачать все ваши данные в CSV формате
										</Typography>
									</Box>
									<IconButton
										color='primary'
										onClick={() => setExportDataOpen(true)}
									>
										<DownloadIcon />
									</IconButton>
								</Box>
							</CardContent>
						</Card>

						<Card variant='outlined' sx={{ mb: 2 }}>
							<CardContent>
								<Box
									display='flex'
									justifyContent='space-between'
									alignItems='center'
								>
									<Box>
										<Typography variant='subtitle1'>Импорт данных</Typography>
										<Typography variant='body2' color='text.secondary'>
											Загрузить данные из файла (скоро)
										</Typography>
									</Box>
									<IconButton color='primary' disabled>
										<UploadIcon />
									</IconButton>
								</Box>
							</CardContent>
						</Card>

						<Card variant='outlined' sx={{ borderColor: 'error.main' }}>
							<CardContent>
								<Box
									display='flex'
									justifyContent='space-between'
									alignItems='center'
								>
									<Box>
										<Typography variant='subtitle1' color='error'>
											Удалить аккаунт
										</Typography>
										<Typography variant='body2' color='text.secondary'>
											Это действие необратимо
										</Typography>
									</Box>
									<IconButton
										color='error'
										onClick={() => setDeleteAccountOpen(true)}
									>
										<DeleteIcon />
									</IconButton>
								</Box>
							</CardContent>
						</Card>
					</Paper>
				</Grid>
			</Grid>

			{/* Export Data Dialog */}
			<ConfirmDialog
				open={exportDataOpen}
				title='Экспорт данных'
				message='Вы хотите скачать все ваши данные в CSV формате?'
				confirmText='Экспортировать'
				cancelText='Отмена'
				onConfirm={handleExportData}
				onCancel={() => setExportDataOpen(false)}
			/>

			{/* Delete Account Dialog */}
			<ConfirmDialog
				open={deleteAccountOpen}
				title='Удалить аккаунт?'
				message='Это действие необратимо. Все ваши данные будут удалены навсегда.'
				confirmText='Удалить аккаунт'
				cancelText='Отмена'
				confirmColor='error'
				onConfirm={handleDeleteAccount}
				onCancel={() => setDeleteAccountOpen(false)}
			/>
		</Box>
	)
}

export default Settings

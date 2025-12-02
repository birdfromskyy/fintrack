import React, { useState, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	Button,
	TextField,
	Box,
	Typography,
	Grid,
	Alert,
} from '@mui/material'
import { CirclePicker } from 'react-color'
import EmojiPicker from 'emoji-picker-react'

import {
	createCategory,
	updateCategory,
} from '../../store/slices/categorySlice'
import { CATEGORY_COLORS } from '../../utils/constants'

const CategoryDialog = ({ open, onClose, category, type, onSave }) => {
	const dispatch = useDispatch()
	const { isLoading } = useSelector(state => state.categories)

	const [formData, setFormData] = useState({
		name: '',
		icon: '',
		color: '#4CAF50',
	})
	const [error, setError] = useState('')
	const [showEmojiPicker, setShowEmojiPicker] = useState(false)

	useEffect(() => {
		if (category) {
			setFormData({
				name: category.name,
				icon: category.icon || '',
				color: category.color || '#4CAF50',
			})
		} else {
			setFormData({
				name: '',
				icon: '',
				color: '#4CAF50',
			})
		}
	}, [category])

	const handleChange = field => event => {
		setFormData({
			...formData,
			[field]: event.target.value,
		})
		setError('')
	}

	const handleColorChange = color => {
		setFormData({
			...formData,
			color: color.hex,
		})
	}

	const handleEmojiSelect = emoji => {
		setFormData({
			...formData,
			icon: emoji.emoji,
		})
		setShowEmojiPicker(false)
	}

	const handleSubmit = async () => {
		// Validation
		if (!formData.name.trim()) {
			setError('Введите название категории')
			return
		}

		const data = {
			name: formData.name.trim(),
			icon: formData.icon,
			color: formData.color,
			type: type,
		}

		let result
		if (category) {
			result = await dispatch(updateCategory({ id: category.id, data }))
		} else {
			result = await dispatch(createCategory(data))
		}

		if (
			createCategory.fulfilled.match(result) ||
			updateCategory.fulfilled.match(result)
		) {
			onSave()
			handleClose()
		}
	}

	const handleClose = () => {
		setFormData({
			name: '',
			icon: '',
			color: '#4CAF50',
		})
		setError('')
		setShowEmojiPicker(false)
		onClose()
	}

	return (
		<Dialog open={open} onClose={handleClose} maxWidth='sm' fullWidth>
			<DialogTitle>
				{category
					? 'Редактировать категорию'
					: `Новая категория ${type === 'income' ? 'доходов' : 'расходов'}`}
			</DialogTitle>

			<DialogContent>
				{error && (
					<Alert severity='error' sx={{ mb: 2 }}>
						{error}
					</Alert>
				)}

				<TextField
					autoFocus
					fullWidth
					label='Название категории'
					value={formData.name}
					onChange={handleChange('name')}
					sx={{ mb: 3 }}
					placeholder='Например: Продукты, Зарплата...'
				/>

				<Box mb={3}>
					<Typography variant='subtitle2' gutterBottom>
						Иконка
					</Typography>
					<Box display='flex' alignItems='center' gap={2}>
						<Box
							sx={{
								width: 60,
								height: 60,
								borderRadius: 2,
								border: '2px solid',
								borderColor: 'divider',
								display: 'flex',
								alignItems: 'center',
								justifyContent: 'center',
								fontSize: '2rem',
								bgcolor: formData.color + '20',
							}}
						>
							{formData.icon || '📂'}
						</Box>
						<Button
							variant='outlined'
							onClick={() => setShowEmojiPicker(!showEmojiPicker)}
						>
							{showEmojiPicker ? 'Закрыть' : 'Выбрать эмодзи'}
						</Button>
					</Box>
					{showEmojiPicker && (
						<Box mt={2}>
							<EmojiPicker
								onEmojiClick={handleEmojiSelect}
								width='100%'
								height={300}
							/>
						</Box>
					)}
				</Box>

				<Box mb={2}>
					<Typography variant='subtitle2' gutterBottom>
						Цвет
					</Typography>
					<CirclePicker
						color={formData.color}
						onChangeComplete={handleColorChange}
						colors={CATEGORY_COLORS}
						width='100%'
					/>
				</Box>

				<Box
					sx={{
						p: 2,
						borderRadius: 1,
						bgcolor: 'grey.100',
						mt: 3,
					}}
				>
					<Typography variant='subtitle2' gutterBottom>
						Предпросмотр
					</Typography>
					<Box display='flex' alignItems='center' gap={2}>
						<Box
							sx={{
								width: 40,
								height: 40,
								borderRadius: '50%',
								bgcolor: formData.color,
								display: 'flex',
								alignItems: 'center',
								justifyContent: 'center',
								fontSize: '1.5rem',
							}}
						>
							{formData.icon || '📂'}
						</Box>
						<Typography variant='body1'>
							{formData.name || 'Название категории'}
						</Typography>
					</Box>
				</Box>
			</DialogContent>

			<DialogActions>
				<Button onClick={handleClose} disabled={isLoading}>
					Отмена
				</Button>
				<Button onClick={handleSubmit} variant='contained' disabled={isLoading}>
					{category ? 'Сохранить' : 'Создать'}
				</Button>
			</DialogActions>
		</Dialog>
	)
}

export default CategoryDialog

import React, { useState, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
	Box,
	Paper,
	Typography,
	Button,
	Tabs,
	Tab,
	Grid,
	Card,
	CardContent,
	IconButton,
	Chip,
	Menu,
	MenuItem,
	Avatar,
	List,
	ListItem,
	ListItemAvatar,
	ListItemText,
	ListItemSecondaryAction,
} from '@mui/material'
import {
	Add as AddIcon,
	Edit as EditIcon,
	Delete as DeleteIcon,
	MoreVert as MoreVertIcon,
	TrendingUp as IncomeIcon,
	TrendingDown as ExpenseIcon,
	Lock as LockIcon,
} from '@mui/icons-material'

import { fetchCategories, deleteCategory } from '../store/slices/categorySlice'
import { formatCurrency } from '../utils/formatters'
import CategoryDialog from '../components/categories/CategoryDialog'
import ConfirmDialog from '../components/common/ConfirmDialog'
import LoadingSpinner from '../components/common/LoadingSpinner'
import ErrorAlert from '../components/common/ErrorAlert'

const Categories = () => {
	const dispatch = useDispatch()
	const { categories, incomeCategories, expenseCategories, isLoading, error } =
		useSelector(state => state.categories)

	const [tabValue, setTabValue] = useState(0)
	const [dialogOpen, setDialogOpen] = useState(false)
	const [editingCategory, setEditingCategory] = useState(null)
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
	const [deletingId, setDeletingId] = useState(null)
	const [anchorEl, setAnchorEl] = useState(null)
	const [selectedCategory, setSelectedCategory] = useState(null)
	const [categoryType, setCategoryType] = useState('expense')

	useEffect(() => {
		dispatch(fetchCategories())
	}, [dispatch])

	const handleTabChange = (event, newValue) => {
		setTabValue(newValue)
	}

	const handleAddCategory = type => {
		setCategoryType(type)
		setEditingCategory(null)
		setDialogOpen(true)
	}

	const handleEditCategory = category => {
		setEditingCategory(category)
		setCategoryType(category.type)
		setDialogOpen(true)
		handleCloseMenu()
	}

	const handleDeleteClick = category => {
		setDeletingId(category.id)
		setDeleteDialogOpen(true)
		handleCloseMenu()
	}

	const handleDeleteConfirm = async () => {
		if (deletingId) {
			await dispatch(deleteCategory(deletingId))
			setDeleteDialogOpen(false)
			setDeletingId(null)
		}
	}

	const handleMenuClick = (event, category) => {
		setAnchorEl(event.currentTarget)
		setSelectedCategory(category)
	}

	const handleCloseMenu = () => {
		setAnchorEl(null)
		setSelectedCategory(null)
	}

	const renderCategoryList = (categoryList, type) => {
		const systemCategories = categoryList.filter(c => c.is_system)
		const userCategories = categoryList.filter(c => !c.is_system)

		return (
			<Box>
				{userCategories.length > 0 && (
					<Box mb={3}>
						<Typography variant='subtitle1' gutterBottom sx={{ mb: 2 }}>
							Пользовательские категории
						</Typography>
						<Grid container spacing={2}>
							{userCategories.map(category => (
								<Grid item xs={12} sm={6} md={4} key={category.id}>
									<Card>
										<CardContent>
											<Box
												display='flex'
												alignItems='center'
												justifyContent='space-between'
											>
												<Box display='flex' alignItems='center'>
													<Avatar
														sx={{
															bgcolor: category.color || 'grey.400',
															width: 40,
															height: 40,
															mr: 2,
														}}
													>
														{category.icon || category.name[0]}
													</Avatar>
													<Box>
														<Typography variant='subtitle1'>
															{category.name}
														</Typography>
														{category.stats && (
															<Typography
																variant='caption'
																color='text.secondary'
															>
																{category.stats.count} транзакций
															</Typography>
														)}
													</Box>
												</Box>
												<IconButton
													size='small'
													onClick={e => handleMenuClick(e, category)}
												>
													<MoreVertIcon />
												</IconButton>
											</Box>
											{category.stats && (
												<Box mt={2}>
													<Typography variant='body2' color='text.secondary'>
														Всего: {formatCurrency(category.stats.total)}
													</Typography>
												</Box>
											)}
										</CardContent>
									</Card>
								</Grid>
							))}
						</Grid>
					</Box>
				)}

				<Box>
					<Typography variant='subtitle1' gutterBottom sx={{ mb: 2 }}>
						Системные категории
					</Typography>
					<Paper>
						<List>
							{systemCategories.map((category, index) => (
								<ListItem
									key={category.id}
									divider={index < systemCategories.length - 1}
								>
									<ListItemAvatar>
										<Avatar
											sx={{
												bgcolor: category.color || 'grey.400',
												width: 36,
												height: 36,
											}}
										>
											{category.icon || category.name[0]}
										</Avatar>
									</ListItemAvatar>
									<ListItemText
										primary={category.name}
										secondary={
											category.stats
												? `${
														category.stats.count
												  } транзакций • ${formatCurrency(
														category.stats.total
												  )}`
												: 'Системная категория'
										}
									/>
									<ListItemSecondaryAction>
										<Chip
											icon={<LockIcon />}
											label='Системная'
											size='small'
											variant='outlined'
										/>
									</ListItemSecondaryAction>
								</ListItem>
							))}
						</List>
					</Paper>
				</Box>
			</Box>
		)
	}

	if (isLoading && categories.length === 0) {
		return <LoadingSpinner />
	}

	return (
		<Box>
			<Box
				display='flex'
				justifyContent='space-between'
				alignItems='center'
				mb={3}
			>
				<Typography variant='h4'>Категории</Typography>
			</Box>

			{error && <ErrorAlert error={error} />}

			<Paper sx={{ mb: 3 }}>
				<Tabs value={tabValue} onChange={handleTabChange} variant='fullWidth'>
					<Tab
						icon={<ExpenseIcon />}
						label={`Расходы (${expenseCategories.length})`}
						iconPosition='start'
					/>
					<Tab
						icon={<IncomeIcon />}
						label={`Доходы (${incomeCategories.length})`}
						iconPosition='start'
					/>
				</Tabs>
			</Paper>

			<TabPanel value={tabValue} index={0}>
				<Box display='flex' justifyContent='flex-end' mb={3}>
					<Button
						variant='contained'
						startIcon={<AddIcon />}
						onClick={() => handleAddCategory('expense')}
					>
						Добавить категорию расходов
					</Button>
				</Box>
				{renderCategoryList(expenseCategories, 'expense')}
			</TabPanel>

			<TabPanel value={tabValue} index={1}>
				<Box display='flex' justifyContent='flex-end' mb={3}>
					<Button
						variant='contained'
						startIcon={<AddIcon />}
						onClick={() => handleAddCategory('income')}
					>
						Добавить категорию доходов
					</Button>
				</Box>
				{renderCategoryList(incomeCategories, 'income')}
			</TabPanel>

			{/* Action Menu */}
			<Menu
				anchorEl={anchorEl}
				open={Boolean(anchorEl)}
				onClose={handleCloseMenu}
			>
				<MenuItem
					onClick={() => handleEditCategory(selectedCategory)}
					disabled={selectedCategory?.is_system}
				>
					<EditIcon fontSize='small' sx={{ mr: 1 }} />
					Редактировать
				</MenuItem>
				<MenuItem
					onClick={() => handleDeleteClick(selectedCategory)}
					disabled={selectedCategory?.is_system}
					sx={{ color: 'error.main' }}
				>
					<DeleteIcon fontSize='small' sx={{ mr: 1 }} />
					Удалить
				</MenuItem>
			</Menu>

			{/* Category Dialog */}
			<CategoryDialog
				open={dialogOpen}
				onClose={() => setDialogOpen(false)}
				category={editingCategory}
				type={categoryType}
				onSave={() => {
					setDialogOpen(false)
					dispatch(fetchCategories())
				}}
			/>

			{/* Delete Confirmation */}
			<ConfirmDialog
				open={deleteDialogOpen}
				title='Удалить категорию?'
				message='Внимание! Категорию можно удалить только если в ней нет транзакций.'
				confirmText='Удалить'
				cancelText='Отмена'
				confirmColor='error'
				onConfirm={handleDeleteConfirm}
				onCancel={() => setDeleteDialogOpen(false)}
			/>
		</Box>
	)
}

// TabPanel component
function TabPanel({ children, value, index }) {
	return (
		<div role='tabpanel' hidden={value !== index}>
			{value === index && <Box>{children}</Box>}
		</div>
	)
}

export default Categories

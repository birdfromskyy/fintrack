import React, { useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
	Box,
	Paper,
	Typography,
	Button,
	IconButton,
	Table,
	TableBody,
	TableCell,
	TableContainer,
	TableHead,
	TableRow,
	TablePagination,
	Chip,
	Menu,
	MenuItem,
	TextField,
	InputAdornment,
	Select,
	FormControl,
	InputLabel,
	Grid,
} from '@mui/material'
import {
	Add as AddIcon,
	Edit as EditIcon,
	Delete as DeleteIcon,
	MoreVert as MoreVertIcon,
	Search as SearchIcon,
	FilterList as FilterIcon,
	Download as DownloadIcon,
	Clear as ClearIcon,
} from '@mui/icons-material'
import { DatePicker } from '@mui/x-date-pickers/DatePicker'

import {
	fetchTransactions,
	deleteTransaction,
	setFilters,
	clearFilters,
} from '../store/slices/transactionSlice'
import { fetchAccounts } from '../store/slices/accountSlice'
import { fetchCategories } from '../store/slices/categorySlice'
import { exportTransactions } from '../store/slices/analyticsSlice'
import { formatCurrency, formatShortDate } from '../utils/formatters'
import TransactionDialog from '../components/transactions/TransactionDialog'
import ConfirmDialog from '../components/common/ConfirmDialog'
import LoadingSpinner from '../components/common/LoadingSpinner'
import EmptyState from '../components/common/EmptyState'
import { Receipt } from '@mui/icons-material'

const Transactions = () => {
	const dispatch = useDispatch()

	const { transactions, filters, pagination, isLoading } = useSelector(
		state => state.transactions
	)
	const { accounts } = useSelector(state => state.accounts)
	const { categories } = useSelector(state => state.categories)

	const [dialogOpen, setDialogOpen] = useState(false)
	const [editingTransaction, setEditingTransaction] = useState(null)
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
	const [deletingId, setDeletingId] = useState(null)
	const [anchorEl, setAnchorEl] = useState(null)
	const [selectedTransaction, setSelectedTransaction] = useState(null)
	const [searchQuery, setSearchQuery] = useState('')
	const [showFilters, setShowFilters] = useState(false)
	const [page, setPage] = useState(0)
	const [rowsPerPage, setRowsPerPage] = useState(25)

	// Filter states
	const [filterType, setFilterType] = useState('')
	const [filterAccount, setFilterAccount] = useState('')
	const [filterCategory, setFilterCategory] = useState('')
	const [filterDateFrom, setFilterDateFrom] = useState(null)
	const [filterDateTo, setFilterDateTo] = useState(null)

	useEffect(() => {
		loadData()
	}, [page, rowsPerPage, filters])

	const loadData = () => {
		dispatch(
			fetchTransactions({
				...filters,
				limit: rowsPerPage,
				offset: page * rowsPerPage,
			})
		)

		if (accounts.length === 0) {
			dispatch(fetchAccounts())
		}
		if (categories.length === 0) {
			dispatch(fetchCategories())
		}
	}

	const handleAddTransaction = () => {
		setEditingTransaction(null)
		setDialogOpen(true)
	}

	const handleEditTransaction = transaction => {
		setEditingTransaction(transaction)
		setDialogOpen(true)
		handleCloseMenu()
	}

	const handleDeleteClick = transaction => {
		setDeletingId(transaction.id)
		setDeleteDialogOpen(true)
		handleCloseMenu()
	}

	const filteredTransactions = transactions.filter(transaction => {
		if (!searchQuery) return true

		const query = searchQuery.toLowerCase()
		return (
			transaction.description?.toLowerCase().includes(query) ||
			transaction.category_name?.toLowerCase().includes(query) ||
			transaction.account_name?.toLowerCase().includes(query) ||
			transaction.amount.toString().includes(query)
		)
	})

	const handleDeleteConfirm = async () => {
		if (deletingId) {
			await dispatch(deleteTransaction(deletingId))
			setDeleteDialogOpen(false)
			setDeletingId(null)
			loadData()
		}
	}

	const handleMenuClick = (event, transaction) => {
		setAnchorEl(event.currentTarget)
		setSelectedTransaction(transaction)
	}

	const handleCloseMenu = () => {
		setAnchorEl(null)
		setSelectedTransaction(null)
	}

	const handleApplyFilters = () => {
		dispatch(
			setFilters({
				type: filterType || null,
				account_id: filterAccount || null,
				category_id: filterCategory || null,
				date_from: filterDateFrom
					? filterDateFrom.toISOString().split('T')[0]
					: null,
				date_to: filterDateTo ? filterDateTo.toISOString().split('T')[0] : null,
			})
		)
		setPage(0)
	}

	const handleClearFilters = () => {
		setFilterType('')
		setFilterAccount('')
		setFilterCategory('')
		setFilterDateFrom(null)
		setFilterDateTo(null)
		dispatch(clearFilters())
		setPage(0)
	}

	const handleExport = async () => {
		const result = await dispatch(exportTransactions(filters))
		if (exportTransactions.fulfilled.match(result)) {
			// Create download link
			const blob = new Blob([result.payload], { type: 'text/csv' })
			const url = window.URL.createObjectURL(blob)
			const link = document.createElement('a')
			link.href = url
			link.download = `transactions_${
				new Date().toISOString().split('T')[0]
			}.csv`
			link.click()
			window.URL.revokeObjectURL(url)
		}
	}

	const handleChangePage = (event, newPage) => {
		setPage(newPage)
	}

	const handleChangeRowsPerPage = event => {
		setRowsPerPage(parseInt(event.target.value, 10))
		setPage(0)
	}

	if (isLoading && transactions.length === 0) {
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
				<Typography variant='h4'>Транзакции</Typography>
				<Box>
					<Button
						variant='outlined'
						startIcon={<DownloadIcon />}
						onClick={handleExport}
						sx={{ mr: 1 }}
					>
						Экспорт
					</Button>
					<Button
						variant='contained'
						startIcon={<AddIcon />}
						onClick={handleAddTransaction}
					>
						Добавить транзакцию
					</Button>
				</Box>
			</Box>

			{/* Search and Filters */}
			<Paper sx={{ p: 2, mb: 3 }}>
				<Grid container spacing={2} alignItems='center'>
					<Grid item xs={12} md={6}>
						<TextField
							fullWidth
							placeholder='Поиск транзакций...'
							value={searchQuery}
							onChange={e => setSearchQuery(e.target.value)}
							InputProps={{
								startAdornment: (
									<InputAdornment position='start'>
										<SearchIcon />
									</InputAdornment>
								),
							}}
						/>
					</Grid>
					<Grid item xs={12} md={6}>
						<Box display='flex' justifyContent='flex-end' gap={1}>
							<Button
								startIcon={<FilterIcon />}
								onClick={() => setShowFilters(!showFilters)}
							>
								Фильтры
							</Button>
							{Object.values(filters).some(v => v) && (
								<Button
									startIcon={<ClearIcon />}
									onClick={handleClearFilters}
									color='secondary'
								>
									Сбросить
								</Button>
							)}
						</Box>
					</Grid>

					{/* Filter Fields */}
					{showFilters && (
						<>
							<Grid item xs={12} md={3}>
								<FormControl fullWidth>
									<InputLabel>Тип</InputLabel>
									<Select
										value={filterType}
										onChange={e => setFilterType(e.target.value)}
										label='Тип'
									>
										<MenuItem value=''>Все</MenuItem>
										<MenuItem value='income'>Доход</MenuItem>
										<MenuItem value='expense'>Расход</MenuItem>
									</Select>
								</FormControl>
							</Grid>

							<Grid item xs={12} md={3}>
								<FormControl fullWidth>
									<InputLabel>Счёт</InputLabel>
									<Select
										value={filterAccount}
										onChange={e => setFilterAccount(e.target.value)}
										label='Счёт'
									>
										<MenuItem value=''>Все</MenuItem>
										{accounts.map(account => (
											<MenuItem key={account.id} value={account.id}>
												{account.name}
											</MenuItem>
										))}
									</Select>
								</FormControl>
							</Grid>

							<Grid item xs={12} md={3}>
								<FormControl fullWidth>
									<InputLabel>Категория</InputLabel>
									<Select
										value={filterCategory}
										onChange={e => setFilterCategory(e.target.value)}
										label='Категория'
									>
										<MenuItem value=''>Все</MenuItem>
										{categories
											.filter(c => !filterType || c.type === filterType)
											.map(category => (
												<MenuItem key={category.id} value={category.id}>
													{category.icon} {category.name}
												</MenuItem>
											))}
									</Select>
								</FormControl>
							</Grid>

							<Grid item xs={12} md={3}>
								<DatePicker
									label='Дата от'
									value={filterDateFrom}
									onChange={setFilterDateFrom}
									slotProps={{ textField: { fullWidth: true } }}
								/>
							</Grid>

							<Grid item xs={12} md={3}>
								<DatePicker
									label='Дата до'
									value={filterDateTo}
									onChange={setFilterDateTo}
									slotProps={{ textField: { fullWidth: true } }}
								/>
							</Grid>

							<Grid item xs={12} md={3}>
								<Button
									fullWidth
									variant='contained'
									onClick={handleApplyFilters}
								>
									Применить
								</Button>
							</Grid>
						</>
					)}
				</Grid>
			</Paper>

			{/* Transactions Table */}
			{transactions.length > 0 ? (
				<TableContainer component={Paper}>
					<Table>
						<TableHead>
							<TableRow>
								<TableCell>Дата</TableCell>
								<TableCell>Категория</TableCell>
								<TableCell>Описание</TableCell>
								<TableCell>Счёт</TableCell>
								<TableCell>Тип</TableCell>
								<TableCell align='right'>Сумма</TableCell>
								<TableCell align='center'>Действия</TableCell>
							</TableRow>
						</TableHead>
						<TableBody>
							{filteredTransactions.map(
								(
									transaction // ✅ Используется filteredTransactions
								) => (
									<TableRow key={transaction.id} hover>
										<TableCell>{formatShortDate(transaction.date)}</TableCell>
										<TableCell>
											<Box display='flex' alignItems='center' gap={1}>
												<span>{transaction.category_icon}</span>
												{transaction.category_name}
											</Box>
										</TableCell>
										<TableCell>{transaction.description || '—'}</TableCell>
										<TableCell>{transaction.account_name}</TableCell>
										<TableCell>
											<Chip
												label={
													transaction.type === 'income' ? 'Доход' : 'Расход'
												}
												color={
													transaction.type === 'income' ? 'success' : 'error'
												}
												size='small'
											/>
										</TableCell>
										<TableCell align='right'>
											<Typography
												color={
													transaction.type === 'income'
														? 'success.main'
														: 'error.main'
												}
												fontWeight='medium'
											>
												{transaction.type === 'income' ? '+' : '-'}
												{formatCurrency(transaction.amount)}
											</Typography>
										</TableCell>
										<TableCell align='center'>
											<IconButton
												size='small'
												onClick={e => handleMenuClick(e, transaction)}
											>
												<MoreVertIcon />
											</IconButton>
										</TableCell>
									</TableRow>
								)
							)}
						</TableBody>
					</Table>

					<TablePagination
						component='div'
						count={pagination.total}
						page={page}
						onPageChange={handleChangePage}
						rowsPerPage={rowsPerPage}
						onRowsPerPageChange={handleChangeRowsPerPage}
						rowsPerPageOptions={[10, 25, 50, 100]}
						labelRowsPerPage='Строк на странице:'
					/>
				</TableContainer>
			) : (
				<Paper sx={{ p: 4 }}>
					<EmptyState
						icon={Receipt}
						title='Нет транзакций'
						message='Начните отслеживать свои финансы, добавив первую транзакцию'
						actionLabel='Добавить транзакцию'
						onAction={handleAddTransaction}
					/>
				</Paper>
			)}

			{/* Action Menu */}
			<Menu
				anchorEl={anchorEl}
				open={Boolean(anchorEl)}
				onClose={handleCloseMenu}
			>
				<MenuItem onClick={() => handleEditTransaction(selectedTransaction)}>
					<EditIcon fontSize='small' sx={{ mr: 1 }} />
					Редактировать
				</MenuItem>
				<MenuItem
					onClick={() => handleDeleteClick(selectedTransaction)}
					sx={{ color: 'error.main' }}
				>
					<DeleteIcon fontSize='small' sx={{ mr: 1 }} />
					Удалить
				</MenuItem>
			</Menu>

			{/* Transaction Dialog */}
			<TransactionDialog
				open={dialogOpen}
				onClose={() => setDialogOpen(false)}
				transaction={editingTransaction}
				onSave={() => {
					setDialogOpen(false)
					loadData()
				}}
			/>

			{/* Delete Confirmation */}
			<ConfirmDialog
				open={deleteDialogOpen}
				title='Удалить транзакцию?'
				message='Это действие нельзя отменить. Транзакция будет удалена навсегда.'
				confirmText='Удалить'
				cancelText='Отмена'
				confirmColor='error'
				onConfirm={handleDeleteConfirm}
				onCancel={() => setDeleteDialogOpen(false)}
			/>
		</Box>
	)
}

export default Transactions

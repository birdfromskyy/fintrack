import React, { useState, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
	Box,
	Grid,
	Card,
	CardContent,
	CardActions,
	Typography,
	Button,
	IconButton,
	Chip,
	Menu,
	MenuItem,
	Tooltip,
} from '@mui/material'
import {
	Add as AddIcon,
	Edit as EditIcon,
	Delete as DeleteIcon,
	MoreVert as MoreVertIcon,
	Star as StarIcon,
	StarBorder as StarBorderIcon,
	AccountBalance as AccountBalanceIcon,
} from '@mui/icons-material'

import {
	fetchAccounts,
	deleteAccount,
	setDefaultAccount,
} from '../store/slices/accountSlice'
import { formatCurrency, formatDate } from '../utils/formatters'
import AccountDialog from '../components/accounts/AccountDialog'
import ConfirmDialog from '../components/common/ConfirmDialog'
import LoadingSpinner from '../components/common/LoadingSpinner'
import EmptyState from '../components/common/EmptyState'
import ErrorAlert from '../components/common/ErrorAlert'

const Accounts = () => {
	const dispatch = useDispatch()
	const { accounts, isLoading, error } = useSelector(state => state.accounts)

	const [dialogOpen, setDialogOpen] = useState(false)
	const [editingAccount, setEditingAccount] = useState(null)
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
	const [deletingId, setDeletingId] = useState(null)
	const [anchorEl, setAnchorEl] = useState(null)
	const [selectedAccount, setSelectedAccount] = useState(null)

	useEffect(() => {
		dispatch(fetchAccounts())
	}, [dispatch])

	const handleAddAccount = () => {
		setEditingAccount(null)
		setDialogOpen(true)
	}

	const handleEditAccount = account => {
		setEditingAccount(account)
		setDialogOpen(true)
		handleCloseMenu()
	}

	const handleDeleteClick = account => {
		setDeletingId(account.id)
		setDeleteDialogOpen(true)
		handleCloseMenu()
	}

	const handleDeleteConfirm = async () => {
		if (deletingId) {
			await dispatch(deleteAccount(deletingId))
			setDeleteDialogOpen(false)
			setDeletingId(null)
		}
	}

	const handleSetDefault = async accountId => {
		await dispatch(setDefaultAccount(accountId))
		handleCloseMenu()
	}

	const handleMenuClick = (event, account) => {
		setAnchorEl(event.currentTarget)
		setSelectedAccount(account)
	}

	const handleCloseMenu = () => {
		setAnchorEl(null)
		setSelectedAccount(null)
	}

	const totalBalance = accounts.reduce((sum, acc) => sum + acc.balance, 0)

	if (isLoading && accounts.length === 0) {
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
				<Box>
					<Typography variant='h4' gutterBottom>
						Счета
					</Typography>
					<Typography variant='subtitle1' color='text.secondary'>
						Общий баланс: {formatCurrency(totalBalance)}
					</Typography>
				</Box>
				<Button
					variant='contained'
					startIcon={<AddIcon />}
					onClick={handleAddAccount}
				>
					Добавить счёт
				</Button>
			</Box>

			{error && <ErrorAlert error={error} />}

			{accounts.length > 0 ? (
				<Grid container spacing={3}>
					{accounts.map(account => (
						<Grid item xs={12} sm={6} md={4} key={account.id}>
							<Card
								sx={{
									position: 'relative',
									...(account.is_default && {
										border: '2px solid',
										borderColor: 'primary.main',
									}),
								}}
							>
								{account.is_default && (
									<Chip
										label='Основной'
										size='small'
										color='primary'
										sx={{
											position: 'absolute',
											top: 8,
											right: 8,
										}}
									/>
								)}

								<CardContent>
									<Box display='flex' alignItems='center' mb={2}>
										<AccountBalanceIcon
											sx={{ mr: 1, color: 'text.secondary' }}
										/>
										<Typography variant='h6' component='div'>
											{account.name}
										</Typography>
									</Box>

									<Typography
										variant='h4'
										component='div'
										color={account.balance >= 0 ? 'text.primary' : 'error.main'}
										gutterBottom
									>
										{formatCurrency(account.balance)}
									</Typography>

									{account.stats && (
										<Box mt={2}>
											<Box display='flex' justifyContent='space-between' mb={1}>
												<Typography variant='body2' color='text.secondary'>
													Доходы:
												</Typography>
												<Typography variant='body2' color='success.main'>
													+{formatCurrency(account.stats.total_income)}
												</Typography>
											</Box>
											<Box display='flex' justifyContent='space-between'>
												<Typography variant='body2' color='text.secondary'>
													Расходы:
												</Typography>
												<Typography variant='body2' color='error.main'>
													-{formatCurrency(account.stats.total_expense)}
												</Typography>
											</Box>
										</Box>
									)}

									<Typography
										variant='caption'
										color='text.secondary'
										display='block'
										mt={2}
									>
										Создан: {formatDate(account.created_at)}
									</Typography>
								</CardContent>

								<CardActions>
									{!account.is_default && (
										<Tooltip title='Сделать основным'>
											<IconButton
												size='small'
												onClick={() => handleSetDefault(account.id)}
											>
												<StarBorderIcon />
											</IconButton>
										</Tooltip>
									)}
									{account.is_default && (
										<IconButton size='small' disabled>
											<StarIcon color='primary' />
										</IconButton>
									)}

									<Box flexGrow={1} />

									<IconButton
										size='small'
										onClick={e => handleMenuClick(e, account)}
									>
										<MoreVertIcon />
									</IconButton>
								</CardActions>
							</Card>
						</Grid>
					))}
				</Grid>
			) : (
				<EmptyState
					icon={AccountBalanceIcon}
					title='Нет счетов'
					message='Создайте свой первый счёт для отслеживания финансов'
					actionLabel='Создать счёт'
					onAction={handleAddAccount}
				/>
			)}

			{/* Action Menu */}
			<Menu
				anchorEl={anchorEl}
				open={Boolean(anchorEl)}
				onClose={handleCloseMenu}
			>
				<MenuItem onClick={() => handleEditAccount(selectedAccount)}>
					<EditIcon fontSize='small' sx={{ mr: 1 }} />
					Редактировать
				</MenuItem>
				{!selectedAccount?.is_default && (
					<MenuItem onClick={() => handleSetDefault(selectedAccount?.id)}>
						<StarBorderIcon fontSize='small' sx={{ mr: 1 }} />
						Сделать основным
					</MenuItem>
				)}
				<MenuItem
					onClick={() => handleDeleteClick(selectedAccount)}
					sx={{ color: 'error.main' }}
					disabled={accounts.length === 1}
				>
					<DeleteIcon fontSize='small' sx={{ mr: 1 }} />
					Удалить
				</MenuItem>
			</Menu>

			{/* Account Dialog */}
			<AccountDialog
				open={dialogOpen}
				onClose={() => setDialogOpen(false)}
				account={editingAccount}
				onSave={() => {
					setDialogOpen(false)
					dispatch(fetchAccounts())
				}}
			/>

			{/* Delete Confirmation */}
			<ConfirmDialog
				open={deleteDialogOpen}
				title='Удалить счёт?'
				message='Внимание! Счёт можно удалить только если на нём нет транзакций.'
				confirmText='Удалить'
				cancelText='Отмена'
				confirmColor='error'
				onConfirm={handleDeleteConfirm}
				onCancel={() => setDeleteDialogOpen(false)}
			/>
		</Box>
	)
}

export default Accounts

import React, { useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
	Grid,
	Card,
	CardContent,
	Typography,
	Box,
	Paper,
	IconButton,
	Chip,
	List,
	ListItem,
	ListItemText,
	ListItemAvatar,
	Avatar,
	Skeleton,
	Tooltip,
} from '@mui/material'
import {
	TrendingUp,
	TrendingDown,
	AccountBalance,
	Receipt,
	ArrowUpward,
	ArrowDownward,
	MoreVert,
	Refresh,
} from '@mui/icons-material'
import { Line, Doughnut } from 'react-chartjs-2'
import {
	Chart as ChartJS,
	CategoryScale,
	LinearScale,
	PointElement,
	LineElement,
	ArcElement,
	Title,
	Tooltip as ChartTooltip,
	Legend,
} from 'chart.js'

import { fetchAccounts } from '../store/slices/accountSlice'
import { fetchTransactions } from '../store/slices/transactionSlice'
import { fetchCategories } from '../store/slices/categorySlice'
import {
	fetchOverview,
	fetchTrends,
	fetchInsights,
} from '../store/slices/analyticsSlice'
import {
	formatCurrency,
	formatDate,
	formatShortDate,
	getRelativeTime,
} from '../utils/formatters'
import { CHART_COLORS } from '../utils/constants'
import LoadingSpinner from '../components/common/LoadingSpinner'
import ErrorAlert from '../components/common/ErrorAlert'

// Register ChartJS components
ChartJS.register(
	CategoryScale,
	LinearScale,
	PointElement,
	LineElement,
	ArcElement,
	Title,
	ChartTooltip,
	Legend
)

const Dashboard = () => {
	const dispatch = useDispatch()

	const { accounts } = useSelector(state => state.accounts)
	const { transactions } = useSelector(state => state.transactions)
	const { overview, trends, insights, isLoading, error } = useSelector(
		state => {
			console.log('Current analytics state:', state.analytics)
			return state.analytics
		}
	)

	useEffect(() => {
		console.log('Dashboard mounted, loading data...')
		loadDashboardData()
	}, [])

	useEffect(() => {
		// Log whenever data changes
		console.log('Dashboard data updated:', {
			overview,
			trends: trends?.length || 0,
			insights: insights?.length || 0,
			accounts: accounts?.length || 0,
			transactions: transactions?.length || 0,
		})
	}, [overview, trends, insights, accounts, transactions])

	const loadDashboardData = async () => {
		try {
			console.log('Starting to load dashboard data...')

			// Load data sequentially to debug better
			const accountsResult = await dispatch(fetchAccounts()).unwrap()
			console.log('Accounts loaded:', accountsResult)

			const categoriesResult = await dispatch(fetchCategories()).unwrap()
			console.log('Categories loaded:', categoriesResult)

			const transactionsResult = await dispatch(
				fetchTransactions({ limit: 10 })
			).unwrap()
			console.log('Transactions loaded:', transactionsResult)

			const overviewResult = await dispatch(fetchOverview('month')).unwrap()
			console.log('Overview loaded:', overviewResult)

			const trendsResult = await dispatch(fetchTrends(30)).unwrap()
			console.log('Trends loaded:', trendsResult)

			const insightsResult = await dispatch(fetchInsights()).unwrap()
			console.log('Insights loaded:', insightsResult)
		} catch (error) {
			console.error('Error loading dashboard data:', error)
		}
	}

	const handleRefresh = () => {
		console.log('Refreshing dashboard data...')
		loadDashboardData()
	}

	// Calculate total balance
	const totalBalance = accounts.reduce((sum, acc) => sum + acc.balance, 0)

	// Safely access overview data with defaults
	// Safely access overview data with defaults - проверяем оба варианта названий
	const totalIncome = overview?.totalIncome || overview?.total_income || 0
	const totalExpense = overview?.totalExpense || overview?.total_expense || 0
	const netIncome =
		overview?.netIncome || overview?.net_income || totalIncome - totalExpense
	const savingsRate = overview?.savingsRate || overview?.savings_rate || 0
	const topCategories =
		overview?.topCategories || overview?.top_categories || []
	const monthComparison =
		overview?.monthComparison || overview?.month_comparison || null

	console.log('Computed values:', {
		totalIncome,
		totalExpense,
		netIncome,
		savingsRate,
		categoriesCount: topCategories.length,
	})

	// Prepare chart data with safe access
	const trendChartData = {
		labels: (trends || []).map(t => formatShortDate(t.date || t.Date)),
		datasets: [
			{
				label: 'Доходы',
				data: (trends || []).map(t => t.income || t.Income || 0),
				borderColor: '#4caf50',
				backgroundColor: 'rgba(76, 175, 80, 0.1)',
				tension: 0.4,
			},
			{
				label: 'Расходы',
				data: (trends || []).map(t => t.expense || t.Expense || 0),
				borderColor: '#f44336',
				backgroundColor: 'rgba(244, 67, 54, 0.1)',
				tension: 0.4,
			},
		],
	}

	const categoryChartData = {
		labels: topCategories
			.slice(0, 5)
			.map(c => c.categoryName || c.category_name || 'Без названия'),
		datasets: [
			{
				data: topCategories.slice(0, 5).map(c => c.amount || 0),
				backgroundColor: CHART_COLORS,
				borderWidth: 0,
			},
		],
	}

	const chartOptions = {
		responsive: true,
		plugins: {
			legend: {
				position: 'bottom',
			},
			tooltip: {
				callbacks: {
					label: function (context) {
						let label = context.label || ''
						if (label) {
							label += ': '
						}
						label += formatCurrency(context.parsed || context.parsed.y || 0)
						return label
					},
				},
			},
		},
		maintainAspectRatio: false,
	}

	if (isLoading && !overview) {
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
				<Typography variant='h4'>Дашборд</Typography>
				<IconButton onClick={handleRefresh}>
					<Refresh />
				</IconButton>
			</Box>

			{error && <ErrorAlert error={error} />}

			{/* Stats Cards */}
			<Grid container spacing={3} mb={3}>
				<Grid item xs={12} sm={6} md={3}>
					<Card>
						<CardContent>
							<Box
								display='flex'
								justifyContent='space-between'
								alignItems='start'
							>
								<Box>
									<Typography
										color='textSecondary'
										gutterBottom
										variant='body2'
									>
										Общий баланс
									</Typography>
									<Typography variant='h5' component='div'>
										{formatCurrency(totalBalance)}
									</Typography>
								</Box>
								<Avatar sx={{ bgcolor: 'primary.main' }}>
									<AccountBalance />
								</Avatar>
							</Box>
						</CardContent>
					</Card>
				</Grid>

				<Grid item xs={12} sm={6} md={3}>
					<Card>
						<CardContent>
							<Box
								display='flex'
								justifyContent='space-between'
								alignItems='start'
							>
								<Box>
									<Typography
										color='textSecondary'
										gutterBottom
										variant='body2'
									>
										Доходы за месяц
									</Typography>
									<Typography variant='h5' component='div'>
										{formatCurrency(totalIncome)}
									</Typography>
									{monthComparison && (
										<Box display='flex' alignItems='center' mt={1}>
											{(monthComparison.incomeChange ||
												monthComparison.income_change ||
												0) > 0 ? (
												<TrendingUp color='success' fontSize='small' />
											) : (
												<TrendingDown color='error' fontSize='small' />
											)}
											<Typography
												variant='caption'
												color={
													(monthComparison.incomeChange ||
														monthComparison.income_change ||
														0) > 0
														? 'success.main'
														: 'error.main'
												}
											>
												{Math.abs(
													monthComparison.incomeChange ||
														monthComparison.income_change ||
														0
												).toFixed(1)}
												%
											</Typography>
										</Box>
									)}
								</Box>
								<Avatar sx={{ bgcolor: 'success.main' }}>
									<ArrowUpward />
								</Avatar>
							</Box>
						</CardContent>
					</Card>
				</Grid>

				<Grid item xs={12} sm={6} md={3}>
					<Card>
						<CardContent>
							<Box
								display='flex'
								justifyContent='space-between'
								alignItems='start'
							>
								<Box>
									<Typography
										color='textSecondary'
										gutterBottom
										variant='body2'
									>
										Расходы за месяц
									</Typography>
									<Typography variant='h5' component='div'>
										{formatCurrency(totalExpense)}
									</Typography>
									{monthComparison && (
										<Box display='flex' alignItems='center' mt={1}>
											{(monthComparison.expenseChange ||
												monthComparison.expense_change ||
												0) > 0 ? (
												<TrendingUp color='error' fontSize='small' />
											) : (
												<TrendingDown color='success' fontSize='small' />
											)}
											<Typography
												variant='caption'
												color={
													(monthComparison.expenseChange ||
														monthComparison.expense_change ||
														0) > 0
														? 'error.main'
														: 'success.main'
												}
											>
												{Math.abs(
													monthComparison.expenseChange ||
														monthComparison.expense_change ||
														0
												).toFixed(1)}
												%
											</Typography>
										</Box>
									)}
								</Box>
								<Avatar sx={{ bgcolor: 'error.main' }}>
									<ArrowDownward />
								</Avatar>
							</Box>
						</CardContent>
					</Card>
				</Grid>

				<Grid item xs={12} sm={6} md={3}>
					<Card>
						<CardContent>
							<Box
								display='flex'
								justifyContent='space-between'
								alignItems='start'
							>
								<Box>
									<Tooltip title='Разница между доходами и расходами за период'>
										<Typography
											color='textSecondary'
											gutterBottom
											variant='body2'
										>
											Сбережения
										</Typography>
									</Tooltip>
									<Typography variant='h5' component='div'>
										{formatCurrency(netIncome)}
									</Typography>
									<Box mt={1}>
										<Tooltip title='Процент дохода, который вы сохраняете'>
											<Chip
												label={`${savingsRate.toFixed(1)}%`}
												color={
													savingsRate > 20
														? 'success'
														: savingsRate > 0
														? 'warning'
														: 'error'
												}
												size='small'
											/>
										</Tooltip>
									</Box>
								</Box>
								<Avatar sx={{ bgcolor: 'info.main' }}>
									<Receipt />
								</Avatar>
							</Box>
						</CardContent>
					</Card>
				</Grid>
			</Grid>

			<Grid container spacing={3}>
				{/* Trends Chart */}
				<Grid item xs={12} md={8}>
					<Paper sx={{ p: 2, height: 400 }}>
						<Box
							display='flex'
							justifyContent='space-between'
							alignItems='center'
							mb={2}
						>
							<Typography variant='h6'>Динамика за месяц</Typography>
							<IconButton size='small'>
								<MoreVert />
							</IconButton>
						</Box>
						{trends && trends.length > 0 ? (
							<Box height={320}>
								<Line data={trendChartData} options={chartOptions} />
							</Box>
						) : (
							<Box
								height={320}
								display='flex'
								alignItems='center'
								justifyContent='center'
							>
								<Typography color='textSecondary'>
									Нет данных для отображения
								</Typography>
							</Box>
						)}
					</Paper>
				</Grid>

				{/* Categories Chart */}
				<Grid item xs={12} md={4}>
					<Paper sx={{ p: 2, height: 400 }}>
						<Box
							display='flex'
							justifyContent='space-between'
							alignItems='center'
							mb={2}
						>
							<Typography variant='h6'>Топ категории</Typography>
							<IconButton size='small'>
								<MoreVert />
							</IconButton>
						</Box>
						{topCategories.length > 0 ? (
							<Box height={320}>
								<Doughnut data={categoryChartData} options={chartOptions} />
							</Box>
						) : (
							<Box
								height={320}
								display='flex'
								alignItems='center'
								justifyContent='center'
							>
								<Typography color='textSecondary'>
									Нет данных для отображения
								</Typography>
							</Box>
						)}
					</Paper>
				</Grid>

				{/* Recent Transactions */}
				<Grid item xs={12} md={6}>
					<Paper sx={{ p: 2 }}>
						<Box
							display='flex'
							justifyContent='space-between'
							alignItems='center'
							mb={2}
						>
							<Typography variant='h6'>Последние транзакции</Typography>
							<IconButton size='small'>
								<MoreVert />
							</IconButton>
						</Box>
						<List>
							{transactions.slice(0, 5).map(transaction => (
								<ListItem key={transaction.id} divider>
									<ListItemAvatar>
										<Avatar
											sx={{
												bgcolor:
													transaction.type === 'income'
														? 'success.light'
														: 'error.light',
											}}
										>
											{transaction.category_icon ||
												(transaction.type === 'income' ? '↑' : '↓')}
										</Avatar>
									</ListItemAvatar>
									<ListItemText
										primary={transaction.category_name}
										secondary={
											<Box>
												<Typography variant='caption' display='block'>
													{transaction.description || 'Без описания'}
												</Typography>
												<Typography variant='caption' color='textSecondary'>
													{formatDate(transaction.date)}
												</Typography>
											</Box>
										}
									/>
									<Typography
										variant='subtitle1'
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
								</ListItem>
							))}
							{transactions.length === 0 && (
								<ListItem>
									<ListItemText
										primary='Нет транзакций'
										secondary='Начните добавлять транзакции для отслеживания финансов'
									/>
								</ListItem>
							)}
						</List>
					</Paper>
				</Grid>

				{/* Insights */}
				<Grid item xs={12} md={6}>
					<Paper sx={{ p: 2 }}>
						<Box
							display='flex'
							justifyContent='space-between'
							alignItems='center'
							mb={2}
						>
							<Typography variant='h6'>Аналитика</Typography>
							<IconButton size='small'>
								<MoreVert />
							</IconButton>
						</Box>
						<List>
							{(insights || []).slice(0, 5).map((insight, index) => (
								<ListItem key={index} divider>
									<ListItemAvatar>
										<Avatar
											sx={{
												bgcolor:
													insight.priority === 'high'
														? 'error.light'
														: insight.priority === 'medium'
														? 'warning.light'
														: 'info.light',
											}}
										>
											{insight.priority === 'high'
												? '!'
												: insight.priority === 'medium'
												? '?'
												: 'i'}
										</Avatar>
									</ListItemAvatar>
									<ListItemText
										primary={insight.title}
										secondary={
											<Box>
												<Typography variant='caption' display='block'>
													{insight.description}
												</Typography>
												<Typography variant='caption' color='textSecondary'>
													{getRelativeTime(insight.date)}
												</Typography>
											</Box>
										}
									/>
								</ListItem>
							))}
							{(!insights || insights.length === 0) && (
								<ListItem>
									<ListItemText
										primary='Нет данных для анализа'
										secondary='Добавьте больше транзакций для получения аналитики'
									/>
								</ListItem>
							)}
						</List>
					</Paper>
				</Grid>
			</Grid>
		</Box>
	)
}

export default Dashboard

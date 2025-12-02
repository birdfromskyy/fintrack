import React, { useState, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
	Box,
	Grid,
	Paper,
	Typography,
	Card,
	CardContent,
	Button,
	ButtonGroup,
	Select,
	MenuItem,
	FormControl,
	InputLabel,
	Chip,
	LinearProgress,
	List,
	ListItem,
	ListItemText,
	ListItemAvatar,
	Avatar,
	Divider,
} from '@mui/material'
import {
	TrendingUp,
	TrendingDown,
	Download as DownloadIcon,
	Refresh as RefreshIcon,
	Assessment as AssessmentIcon,
	PieChart as PieChartIcon,
	ShowChart as ShowChartIcon,
	Savings as SavingsIcon,
} from '@mui/icons-material'
import {
	LineChart,
	Line,
	BarChart,
	Bar,
	PieChart,
	Pie,
	Cell,
	XAxis,
	YAxis,
	CartesianGrid,
	Tooltip,
	Legend,
	ResponsiveContainer,
	AreaChart,
	Area,
} from 'recharts'

import {
	fetchOverview,
	fetchTrends,
	fetchForecast,
	fetchInsights,
	fetchCashflow,
	exportTransactions,
} from '../store/slices/analyticsSlice'
import {
	formatCurrency,
	formatPercent,
	formatShortDate,
} from '../utils/formatters'
import { PERIODS, CHART_COLORS } from '../utils/constants'
import LoadingSpinner from '../components/common/LoadingSpinner'
import ErrorAlert from '../components/common/ErrorAlert'

const Analytics = () => {
	const dispatch = useDispatch()
	const { overview, trends, forecast, insights, cashflow, isLoading, error } =
		useSelector(state => state.analytics)

	const [period, setPeriod] = useState('month')
	const [dateRange, setDateRange] = useState({
		start: new Date(new Date().setMonth(new Date().getMonth() - 1)),
		end: new Date(),
	})
	const [chartType, setChartType] = useState('line')

	useEffect(() => {
		loadAnalyticsData()
	}, [period, dateRange])

	const loadAnalyticsData = () => {
		dispatch(fetchOverview(period))
		dispatch(
			fetchTrends(
				period === 'week'
					? 7
					: period === 'month'
					? 30
					: period === 'quarter'
					? 90
					: 365
			)
		)
		dispatch(fetchForecast(3))
		dispatch(fetchInsights())
		dispatch(
			fetchCashflow({
				startDate: dateRange.start.toISOString().split('T')[0],
				endDate: dateRange.end.toISOString().split('T')[0],
			})
		)
	}

	const handlePeriodChange = event => {
		setPeriod(event.target.value)
	}

	const handleExport = async format => {
		const result = await dispatch(
			exportTransactions({
				format,
				date_from: dateRange.start.toISOString().split('T')[0],
				date_to: dateRange.end.toISOString().split('T')[0],
			})
		)

		if (exportTransactions.fulfilled.match(result)) {
			const blob = new Blob([result.payload], {
				type: format === 'csv' ? 'text/csv' : 'application/json',
			})
			const url = window.URL.createObjectURL(blob)
			const link = document.createElement('a')
			link.href = url
			link.download = `analytics_${period}_${
				new Date().toISOString().split('T')[0]
			}.${format}`
			link.click()
			window.URL.revokeObjectURL(url)
		}
	}

	// ✅ ИСПРАВЛЕНО: Используем правильные поля из backend
	const totalIncome = overview?.total_income || 0
	const totalExpense = overview?.total_expense || 0
	const netIncome = overview?.net_income || 0
	const savingsRate = overview?.savings_rate || 0
	const topCategories = overview?.top_categories || []
	const monthComparison = overview?.month_comparison

	console.log('Analytics computed values:', {
		totalIncome,
		totalExpense,
		netIncome,
		savingsRate,
		categoriesCount: topCategories.length,
	})

	// Prepare chart data
	const trendsChartData = Array.isArray(trends)
		? trends.map(t => ({
				date: formatShortDate(t.date),
				income: t.income || 0,
				expense: t.expense || 0,
				balance: t.balance || 0,
		  }))
		: []

	// ✅ ИСПРАВЛЕНО: Фильтруем только расходы для диаграммы
	const expenseCategories = topCategories.filter(c => c.type === 'expense')
	const categoryChartData = expenseCategories.slice(0, 5).map(c => ({
		name: c.category_name || c.name,
		value: c.amount || 0,
		percentage: c.percentage || 0,
	}))

	const cashflowChartData = Array.isArray(cashflow)
		? cashflow.map(c => ({
				date: formatShortDate(c.date),
				inflow: c.total_inflow || 0,
				outflow: c.total_outflow || 0,
				net: c.net_cashflow || 0,
		  }))
		: []

	const CustomTooltip = ({ active, payload, label }) => {
		if (active && payload && payload.length) {
			return (
				<Paper sx={{ p: 1 }}>
					<Typography variant='caption'>{label}</Typography>
					{payload.map((entry, index) => (
						<Typography
							key={index}
							variant='body2'
							style={{ color: entry.color }}
						>
							{entry.name}: {formatCurrency(entry.value)}
						</Typography>
					))}
				</Paper>
			)
		}
		return null
	}

	if (isLoading && !overview) {
		return <LoadingSpinner />
	}

	return (
		<Box>
			{/* Header */}
			<Box
				display='flex'
				justifyContent='space-between'
				alignItems='center'
				mb={3}
			>
				<Typography variant='h4'>Аналитика</Typography>
				<Box display='flex' gap={2}>
					<FormControl size='small' sx={{ minWidth: 120 }}>
						<InputLabel>Период</InputLabel>
						<Select value={period} onChange={handlePeriodChange} label='Период'>
							{PERIODS.map(p => (
								<MenuItem key={p.value} value={p.value}>
									{p.label}
								</MenuItem>
							))}
						</Select>
					</FormControl>
					<ButtonGroup variant='outlined'>
						<Button onClick={() => handleExport('csv')}>
							<DownloadIcon /> CSV
						</Button>
					</ButtonGroup>
					<Button
						variant='contained'
						onClick={loadAnalyticsData}
						startIcon={<RefreshIcon />}
					>
						Обновить
					</Button>
				</Box>
			</Box>

			{error && <ErrorAlert error={error} />}

			{/* Key Metrics */}
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
										Доходы
									</Typography>
									<Typography variant='h5' component='div' color='success.main'>
										{formatCurrency(totalIncome)}
									</Typography>
									{monthComparison && (
										<Box display='flex' alignItems='center' mt={1}>
											{monthComparison.income_change > 0 ? (
												<TrendingUp color='success' fontSize='small' />
											) : (
												<TrendingDown color='error' fontSize='small' />
											)}
											<Typography variant='caption'>
												{formatPercent(Math.abs(monthComparison.income_change))}
											</Typography>
										</Box>
									)}
								</Box>
								<Avatar sx={{ bgcolor: 'success.light' }}>
									<TrendingUp />
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
										Расходы
									</Typography>
									<Typography variant='h5' component='div' color='error.main'>
										{formatCurrency(totalExpense)}
									</Typography>
									{monthComparison && (
										<Box display='flex' alignItems='center' mt={1}>
											{monthComparison.expense_change > 0 ? (
												<TrendingUp color='error' fontSize='small' />
											) : (
												<TrendingDown color='success' fontSize='small' />
											)}
											<Typography variant='caption'>
												{formatPercent(
													Math.abs(monthComparison.expense_change)
												)}
											</Typography>
										</Box>
									)}
								</Box>
								<Avatar sx={{ bgcolor: 'error.light' }}>
									<TrendingDown />
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
										Баланс
									</Typography>
									<Typography variant='h5' component='div'>
										{formatCurrency(netIncome)}
									</Typography>
									<LinearProgress
										variant='determinate'
										value={Math.min(100, Math.max(0, savingsRate))}
										sx={{ mt: 2 }}
										color={savingsRate > 20 ? 'success' : 'warning'}
									/>
								</Box>
								<Avatar sx={{ bgcolor: 'info.light' }}>
									<SavingsIcon />
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
										Норма сбережений
									</Typography>
									<Typography variant='h5' component='div'>
										{formatPercent(savingsRate)}
									</Typography>
									<Chip
										label={
											savingsRate > 30
												? 'Отлично'
												: savingsRate > 20
												? 'Хорошо'
												: savingsRate > 10
												? 'Нормально'
												: savingsRate > 0
												? 'Низко'
												: 'Отрицательно'
										}
										color={
											savingsRate > 30
												? 'success'
												: savingsRate > 20
												? 'primary'
												: savingsRate > 10
												? 'warning'
												: 'error'
										}
										size='small'
										sx={{ mt: 1 }}
									/>
								</Box>
								<Avatar sx={{ bgcolor: 'secondary.light' }}>
									<AssessmentIcon />
								</Avatar>
							</Box>
						</CardContent>
					</Card>
				</Grid>
			</Grid>

			{/* Charts */}
			<Grid container spacing={3}>
				{/* Trends Chart */}
				<Grid item xs={12} md={8}>
					<Paper sx={{ p: 3 }}>
						<Box
							display='flex'
							justifyContent='space-between'
							alignItems='center'
							mb={2}
						>
							<Typography variant='h6'>Динамика</Typography>
							<ButtonGroup size='small'>
								<Button
									variant={chartType === 'line' ? 'contained' : 'outlined'}
									onClick={() => setChartType('line')}
								>
									<ShowChartIcon />
								</Button>
								<Button
									variant={chartType === 'bar' ? 'contained' : 'outlined'}
									onClick={() => setChartType('bar')}
								>
									<AssessmentIcon />
								</Button>
								<Button
									variant={chartType === 'area' ? 'contained' : 'outlined'}
									onClick={() => setChartType('area')}
								>
									<PieChartIcon />
								</Button>
							</ButtonGroup>
						</Box>

						{trendsChartData.length > 0 ? (
							<ResponsiveContainer width='100%' height={300}>
								{chartType === 'line' ? (
									<LineChart data={trendsChartData}>
										<CartesianGrid strokeDasharray='3 3' />
										<XAxis dataKey='date' />
										<YAxis />
										<Tooltip content={<CustomTooltip />} />
										<Legend />
										<Line
											type='monotone'
											dataKey='income'
											stroke='#4caf50'
											name='Доходы'
											strokeWidth={2}
										/>
										<Line
											type='monotone'
											dataKey='expense'
											stroke='#f44336'
											name='Расходы'
											strokeWidth={2}
										/>
									</LineChart>
								) : chartType === 'bar' ? (
									<BarChart data={trendsChartData}>
										<CartesianGrid strokeDasharray='3 3' />
										<XAxis dataKey='date' />
										<YAxis />
										<Tooltip content={<CustomTooltip />} />
										<Legend />
										<Bar dataKey='income' fill='#4caf50' name='Доходы' />
										<Bar dataKey='expense' fill='#f44336' name='Расходы' />
									</BarChart>
								) : (
									<AreaChart data={trendsChartData}>
										<CartesianGrid strokeDasharray='3 3' />
										<XAxis dataKey='date' />
										<YAxis />
										<Tooltip content={<CustomTooltip />} />
										<Legend />
										<Area
											type='monotone'
											dataKey='income'
											stackId='1'
											stroke='#4caf50'
											fill='#4caf50'
											fillOpacity={0.6}
											name='Доходы'
										/>
										<Area
											type='monotone'
											dataKey='expense'
											stackId='2'
											stroke='#f44336'
											fill='#f44336'
											fillOpacity={0.6}
											name='Расходы'
										/>
									</AreaChart>
								)}
							</ResponsiveContainer>
						) : (
							<Typography
								variant='body2'
								color='textSecondary'
								align='center'
								py={4}
							>
								Нет данных для отображения
							</Typography>
						)}
					</Paper>
				</Grid>

				{/* Category Distribution */}
				<Grid item xs={12} md={4}>
					<Paper sx={{ p: 3 }}>
						<Typography variant='h6' gutterBottom>
							Распределение расходов по категориям
						</Typography>
						{categoryChartData.length > 0 ? (
							<ResponsiveContainer width='100%' height={300}>
								<PieChart>
									<Pie
										data={categoryChartData}
										cx='50%'
										cy='50%'
										labelLine={false}
										label={({ name, percentage }) =>
											`${name}: ${formatPercent(percentage)}`
										}
										outerRadius={80}
										fill='#8884d8'
										dataKey='value'
									>
										{categoryChartData.map((entry, index) => (
											<Cell
												key={`cell-${index}`}
												fill={CHART_COLORS[index % CHART_COLORS.length]}
											/>
										))}
									</Pie>
									<Tooltip formatter={value => formatCurrency(value)} />
								</PieChart>
							</ResponsiveContainer>
						) : (
							<Typography
								variant='body2'
								color='textSecondary'
								align='center'
								py={4}
							>
								Нет данных для отображения
							</Typography>
						)}
					</Paper>
				</Grid>

				{/* Forecast */}
				<Grid item xs={12} md={6}>
					<Paper sx={{ p: 3 }}>
						<Typography variant='h6' gutterBottom>
							Прогноз на 3 месяца
						</Typography>
						{forecast ? (
							<Box>
								<Grid container spacing={2}>
									<Grid item xs={6}>
										<Typography variant='body2' color='textSecondary'>
											Прогноз доходов
										</Typography>
										<Typography variant='h6' color='success.main'>
											{formatCurrency(forecast.predicted_income || 0)}
										</Typography>
									</Grid>
									<Grid item xs={6}>
										<Typography variant='body2' color='textSecondary'>
											Прогноз расходов
										</Typography>
										<Typography variant='h6' color='error.main'>
											{formatCurrency(forecast.predicted_expense || 0)}
										</Typography>
									</Grid>
									<Grid item xs={6}>
										<Typography variant='body2' color='textSecondary'>
											Прогноз баланса
										</Typography>
										<Typography variant='h6'>
											{formatCurrency(forecast.predicted_balance || 0)}
										</Typography>
									</Grid>
									<Grid item xs={6}>
										<Typography variant='body2' color='textSecondary'>
											Точность прогноза
										</Typography>
										<Box display='flex' alignItems='center' gap={1}>
											<LinearProgress
												variant='determinate'
												value={forecast.confidence || 0}
												sx={{ flexGrow: 1 }}
											/>
											<Typography variant='body2'>
												{formatPercent(forecast.confidence || 0)}
											</Typography>
										</Box>
									</Grid>
								</Grid>
								<Typography
									variant='caption'
									color='textSecondary'
									display='block'
									mt={2}
								>
									* Прогноз основан на данных за последние{' '}
									{forecast.based_on_months || 0} месяцев
								</Typography>
							</Box>
						) : (
							<Typography variant='body2' color='textSecondary'>
								Недостаточно данных для прогноза
							</Typography>
						)}
					</Paper>
				</Grid>

				{/* Insights */}
				<Grid item xs={12} md={6}>
					<Paper sx={{ p: 3 }}>
						<Typography variant='h6' gutterBottom>
							Рекомендации
						</Typography>
						<List>
							{Array.isArray(insights) && insights.length > 0 ? (
								insights.slice(0, 4).map((insight, index) => (
									<React.Fragment key={index}>
										<ListItem alignItems='flex-start'>
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
													{insight.priority === 'high' ? '!' : '💡'}
												</Avatar>
											</ListItemAvatar>
											<ListItemText
												primary={insight.title}
												secondary={insight.description}
											/>
										</ListItem>
										{index < insights.slice(0, 4).length - 1 && (
											<Divider variant='inset' component='li' />
										)}
									</React.Fragment>
								))
							) : (
								<Typography variant='body2' color='textSecondary'>
									Добавьте больше транзакций для получения рекомендаций
								</Typography>
							)}
						</List>
					</Paper>
				</Grid>
			</Grid>
		</Box>
	)
}

export default Analytics

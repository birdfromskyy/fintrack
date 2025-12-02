export const formatCurrency = (amount, showSign = false) => {
	const formatted = new Intl.NumberFormat('ru-RU', {
		style: 'currency',
		currency: 'RUB',
		minimumFractionDigits: 0,
		maximumFractionDigits: 2,
	}).format(Math.abs(amount))

	if (showSign && amount > 0) {
		return `+${formatted}`
	}
	return formatted
}

export const formatDate = date => {
	if (!date) return ''
	const d = new Date(date)
	return new Intl.DateTimeFormat('ru-RU', {
		day: '2-digit',
		month: 'long',
		year: 'numeric',
	}).format(d)
}

export const formatShortDate = date => {
	if (!date) return ''
	const d = new Date(date)
	return new Intl.DateTimeFormat('ru-RU', {
		day: '2-digit',
		month: '2-digit',
		year: 'numeric',
	}).format(d)
}

export const formatDateTime = date => {
	if (!date) return ''
	const d = new Date(date)
	return new Intl.DateTimeFormat('ru-RU', {
		day: '2-digit',
		month: '2-digit',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
	}).format(d)
}

export const formatPercent = (value, decimals = 1) => {
	return `${value.toFixed(decimals)}%`
}

export const formatNumber = (value, decimals = 0) => {
	return new Intl.NumberFormat('ru-RU', {
		minimumFractionDigits: decimals,
		maximumFractionDigits: decimals,
	}).format(value)
}

export const getMonthName = date => {
	const d = new Date(date)
	return new Intl.DateTimeFormat('ru-RU', { month: 'long' }).format(d)
}

export const getRelativeTime = date => {
	const now = new Date()
	const d = new Date(date)
	const diff = now - d
	const seconds = Math.floor(diff / 1000)
	const minutes = Math.floor(seconds / 60)
	const hours = Math.floor(minutes / 60)
	const days = Math.floor(hours / 24)

	if (days > 0) {
		return `${days} ${days === 1 ? 'день' : 'дней'} назад`
	}
	if (hours > 0) {
		return `${hours} ${hours === 1 ? 'час' : 'часов'} назад`
	}
	if (minutes > 0) {
		return `${minutes} ${minutes === 1 ? 'минуту' : 'минут'} назад`
	}
	return 'только что'
}

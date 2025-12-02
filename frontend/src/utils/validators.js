export const validateEmail = email => {
	const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
	return re.test(email)
}

export const validatePassword = password => {
	return password && password.length >= 6
}

export const validateAmount = amount => {
	return amount && !isNaN(amount) && parseFloat(amount) > 0
}

export const validateRequired = value => {
	return value && value.toString().trim().length > 0
}

export const validateCode = code => {
	return code && /^\d{6}$/.test(code)
}

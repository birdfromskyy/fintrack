import React from 'react'
import { Alert, AlertTitle } from '@mui/material'

const ErrorAlert = ({ error, title = 'Ошибка' }) => {
	if (!error) return null

	return (
		<Alert severity='error' sx={{ mb: 2 }}>
			<AlertTitle>{title}</AlertTitle>
			{error}
		</Alert>
	)
}

export default ErrorAlert

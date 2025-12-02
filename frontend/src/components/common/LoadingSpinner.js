import React from 'react'
import { Box, CircularProgress } from '@mui/material'

const LoadingSpinner = ({ size = 40 }) => {
	return (
		<Box
			display='flex'
			justifyContent='center'
			alignItems='center'
			minHeight={200}
		>
			<CircularProgress size={size} />
		</Box>
	)
}

export default LoadingSpinner

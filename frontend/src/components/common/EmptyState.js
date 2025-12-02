import React from 'react'
import { Box, Typography, Button } from '@mui/material'
import { Add as AddIcon } from '@mui/icons-material'

const EmptyState = ({ title, message, icon: Icon, actionLabel, onAction }) => {
	return (
		<Box
			display='flex'
			flexDirection='column'
			alignItems='center'
			justifyContent='center'
			py={8}
		>
			{Icon && (
				<Box
					sx={{
						width: 80,
						height: 80,
						borderRadius: '50%',
						backgroundColor: 'action.hover',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center',
						mb: 3,
					}}
				>
					<Icon sx={{ fontSize: 40, color: 'text.secondary' }} />
				</Box>
			)}
			<Typography variant='h6' color='text.primary' gutterBottom>
				{title}
			</Typography>
			{message && (
				<Typography
					variant='body2'
					color='text.secondary'
					align='center'
					sx={{ mb: 3 }}
				>
					{message}
				</Typography>
			)}
			{actionLabel && onAction && (
				<Button variant='contained' startIcon={<AddIcon />} onClick={onAction}>
					{actionLabel}
				</Button>
			)}
		</Box>
	)
}

export default EmptyState

import React from 'react'
import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogContentText,
	DialogActions,
	Button,
} from '@mui/material'

const ConfirmDialog = ({
	open,
	title,
	message,
	onConfirm,
	onCancel,
	confirmText = 'Подтвердить',
	cancelText = 'Отмена',
	confirmColor = 'primary',
}) => {
	return (
		<Dialog open={open} onClose={onCancel} maxWidth='sm' fullWidth>
			<DialogTitle>{title}</DialogTitle>
			<DialogContent>
				<DialogContentText>{message}</DialogContentText>
			</DialogContent>
			<DialogActions>
				<Button onClick={onCancel} color='inherit'>
					{cancelText}
				</Button>
				<Button onClick={onConfirm} color={confirmColor} variant='contained'>
					{confirmText}
				</Button>
			</DialogActions>
		</Dialog>
	)
}

export default ConfirmDialog

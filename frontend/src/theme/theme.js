import { createTheme } from '@mui/material/styles'

export const lightTheme = createTheme({
	palette: {
		mode: 'light',
		primary: {
			main: '#1976d2',
			light: '#42a5f5',
			dark: '#1565c0',
		},
		secondary: {
			main: '#dc004e',
			light: '#e33371',
			dark: '#9a0036',
		},
		background: {
			default: '#f5f5f5',
			paper: '#ffffff',
		},
	},
	// ... rest of theme config
})

export const darkTheme = createTheme({
	palette: {
		mode: 'dark',
		primary: {
			main: '#90caf9',
			light: '#e3f2fd',
			dark: '#42a5f5',
		},
		secondary: {
			main: '#f48fb1',
			light: '#ffc0cb',
			dark: '#f06292',
		},
		background: {
			default: '#121212',
			paper: '#1e1e1e',
		},
		text: {
			primary: '#ffffff',
			secondary: 'rgba(255, 255, 255, 0.7)',
		},
	},
	// ... rest of theme config
})

const getTheme = mode => (mode === 'dark' ? darkTheme : lightTheme)

export default getTheme

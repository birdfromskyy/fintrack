import React from 'react'
import ReactDOM from 'react-dom/client'
import { Provider, useSelector } from 'react-redux'
import { BrowserRouter } from 'react-router-dom'
import { ThemeProvider } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'
import { LocalizationProvider } from '@mui/x-date-pickers'
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns'
import ruLocale from 'date-fns/locale/ru'

import './index.css'
import App from './App'
import { store } from './store/store'
import getTheme from './theme/theme'

function Root() {
	const theme = useSelector(state => state.ui.theme)

	return (
		<ThemeProvider theme={getTheme(theme)}>
			<LocalizationProvider
				dateAdapter={AdapterDateFns}
				adapterLocale={ruLocale}
			>
				<CssBaseline />
				<App />
			</LocalizationProvider>
		</ThemeProvider>
	)
}

const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
	<React.StrictMode>
		<Provider store={store}>
			<BrowserRouter>
				<Root />
			</BrowserRouter>
		</Provider>
	</React.StrictMode>
)

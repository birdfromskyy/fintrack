import React, { useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import { Box, CircularProgress } from '@mui/material'

import Layout from './components/layout/Layout'
import PublicLayout from './components/layout/PublicLayout'
import PrivateRoute from './components/common/PrivateRoute'

// Auth pages
import Login from './pages/auth/Login'
import Register from './pages/auth/Register'
import VerifyEmail from './pages/auth/VerifyEmail'

// Main pages
import Dashboard from './pages/Dashboard'
import Transactions from './pages/Transactions'
import Accounts from './pages/Accounts'
import Categories from './pages/Categories'
import Analytics from './pages/Analytics'
import Settings from './pages/Settings'

import { checkAuth } from './store/slices/authSlice'

function App() {
	const dispatch = useDispatch()
	const { isAuthenticated, isLoading } = useSelector(state => state.auth)

	useEffect(() => {
		dispatch(checkAuth())
	}, [dispatch])

	if (isLoading) {
		return (
			<Box
				display='flex'
				justifyContent='center'
				alignItems='center'
				minHeight='100vh'
			>
				<CircularProgress />
			</Box>
		)
	}

	return (
		<Routes>
			{/* Public routes */}
			<Route element={<PublicLayout />}>
				<Route
					path='/login'
					element={!isAuthenticated ? <Login /> : <Navigate to='/dashboard' />}
				/>
				<Route
					path='/register'
					element={
						!isAuthenticated ? <Register /> : <Navigate to='/dashboard' />
					}
				/>
				<Route
					path='/verify-email'
					element={
						!isAuthenticated ? <VerifyEmail /> : <Navigate to='/dashboard' />
					}
				/>
			</Route>

			{/* Private routes */}
			<Route element={<PrivateRoute />}>
				<Route element={<Layout />}>
					<Route path='/dashboard' element={<Dashboard />} />
					<Route path='/transactions' element={<Transactions />} />
					<Route path='/accounts' element={<Accounts />} />
					<Route path='/categories' element={<Categories />} />
					<Route path='/analytics' element={<Analytics />} />
					<Route path='/settings' element={<Settings />} />
				</Route>
			</Route>

			{/* Default redirect */}
			<Route
				path='/'
				element={<Navigate to={isAuthenticated ? '/dashboard' : '/login'} />}
			/>
			<Route
				path='*'
				element={<Navigate to={isAuthenticated ? '/dashboard' : '/login'} />}
			/>
		</Routes>
	)
}

export default App

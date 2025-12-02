import React, { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
	Box,
	Drawer,
	AppBar,
	Toolbar,
	List,
	Typography,
	Divider,
	IconButton,
	ListItem,
	ListItemButton,
	ListItemIcon,
	ListItemText,
	Avatar,
	Menu,
	MenuItem,
} from '@mui/material'
import {
	Menu as MenuIcon,
	Dashboard as DashboardIcon,
	AccountBalance as AccountBalanceIcon,
	Category as CategoryIcon,
	Receipt as ReceiptIcon,
	Analytics as AnalyticsIcon,
	Settings as SettingsIcon,
	Logout as LogoutIcon,
	Person as PersonIcon,
} from '@mui/icons-material'
import { useDispatch, useSelector } from 'react-redux'
import { logout } from '../../store/slices/authSlice'

const drawerWidth = 240

const menuItems = [
	{ text: 'Дашборд', icon: <DashboardIcon />, path: '/dashboard' },
	{ text: 'Транзакции', icon: <ReceiptIcon />, path: '/transactions' },
	{ text: 'Счета', icon: <AccountBalanceIcon />, path: '/accounts' },
	{ text: 'Категории', icon: <CategoryIcon />, path: '/categories' },
	{ text: 'Аналитика', icon: <AnalyticsIcon />, path: '/analytics' },
	{ text: 'Настройки', icon: <SettingsIcon />, path: '/settings' },
]

const Layout = () => {
	const navigate = useNavigate()
	const location = useLocation()
	const dispatch = useDispatch()
	const { user } = useSelector(state => state.auth)
	const [mobileOpen, setMobileOpen] = useState(false)
	const [anchorEl, setAnchorEl] = useState(null)

	const handleDrawerToggle = () => {
		setMobileOpen(!mobileOpen)
	}

	const handleMenuClick = event => {
		setAnchorEl(event.currentTarget)
	}

	const handleMenuClose = () => {
		setAnchorEl(null)
	}

	const handleLogout = () => {
		dispatch(logout())
		navigate('/login')
	}

	const drawer = (
		<div>
			<Toolbar>
				<Typography variant='h6' noWrap component='div'>
					Fintrack
				</Typography>
			</Toolbar>
			<Divider />
			<List>
				{menuItems.map(item => (
					<ListItem key={item.text} disablePadding>
						<ListItemButton
							selected={location.pathname === item.path}
							onClick={() => navigate(item.path)}
						>
							<ListItemIcon>{item.icon}</ListItemIcon>
							<ListItemText primary={item.text} />
						</ListItemButton>
					</ListItem>
				))}
			</List>
		</div>
	)

	return (
		<Box sx={{ display: 'flex' }}>
			<AppBar
				position='fixed'
				sx={{
					width: { sm: `calc(100% - ${drawerWidth}px)` },
					ml: { sm: `${drawerWidth}px` },
				}}
			>
				<Toolbar>
					<IconButton
						color='inherit'
						edge='start'
						onClick={handleDrawerToggle}
						sx={{ mr: 2, display: { sm: 'none' } }}
					>
						<MenuIcon />
					</IconButton>
					<Typography variant='h6' noWrap component='div' sx={{ flexGrow: 1 }}>
						{menuItems.find(item => item.path === location.pathname)?.text ||
							'Fintrack'}
					</Typography>
					<IconButton onClick={handleMenuClick} sx={{ p: 0 }}>
						<Avatar sx={{ bgcolor: 'secondary.main' }}>
							{user?.email?.[0]?.toUpperCase() || 'U'}
						</Avatar>
					</IconButton>
					<Menu
						anchorEl={anchorEl}
						open={Boolean(anchorEl)}
						onClose={handleMenuClose}
						anchorOrigin={{
							vertical: 'bottom',
							horizontal: 'right',
						}}
						transformOrigin={{
							vertical: 'top',
							horizontal: 'right',
						}}
					>
						<MenuItem disabled>
							<ListItemIcon>
								<PersonIcon fontSize='small' />
							</ListItemIcon>
							<Typography variant='body2'>{user?.email}</Typography>
						</MenuItem>
						<Divider />
						<MenuItem onClick={handleLogout}>
							<ListItemIcon>
								<LogoutIcon fontSize='small' />
							</ListItemIcon>
							Выйти
						</MenuItem>
					</Menu>
				</Toolbar>
			</AppBar>
			<Box
				component='nav'
				sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}
			>
				<Drawer
					variant='temporary'
					open={mobileOpen}
					onClose={handleDrawerToggle}
					ModalProps={{
						keepMounted: true,
					}}
					sx={{
						display: { xs: 'block', sm: 'none' },
						'& .MuiDrawer-paper': {
							boxSizing: 'border-box',
							width: drawerWidth,
						},
					}}
				>
					{drawer}
				</Drawer>
				<Drawer
					variant='permanent'
					sx={{
						display: { xs: 'none', sm: 'block' },
						'& .MuiDrawer-paper': {
							boxSizing: 'border-box',
							width: drawerWidth,
						},
					}}
					open
				>
					{drawer}
				</Drawer>
			</Box>
			<Box
				component='main'
				sx={{
					flexGrow: 1,
					p: 3,
					width: { sm: `calc(100% - ${drawerWidth}px)` },
				}}
			>
				<Toolbar />
				<Outlet />
			</Box>
		</Box>
	)
}

export default Layout

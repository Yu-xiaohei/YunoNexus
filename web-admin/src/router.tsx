import { Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { Spin } from 'antd'
import MainLayout from './components/layout/MainLayout'
import Login from './pages/Login'
import Forbidden from './pages/Forbidden'
import Dashboard from './pages/Dashboard'
import TunnelList from './pages/tunnels/TunnelList'
import DeviceList from './pages/devices/DeviceList'
import UserList from './pages/users/UserList'
import TrafficStats from './pages/traffic/TrafficStats'
import SystemSettings from './pages/settings/SystemSettings'
import { useAuthStore } from './store/authStore'

// 受保护的路由组件
function ProtectedRoute({ children, requireAdmin = false }: { children: React.ReactNode; requireAdmin?: boolean }) {
  const { token, user, loadUser } = useAuthStore()
  const location = useLocation()
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (token && !user) {
      loadUser()
    }
    setLoading(false)
  }, [token, user, loadUser])

  if (loading) {
    return <Spin size="large" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }} />
  }

  // 未登录，跳转到登录页
  if (!token) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  // 需要管理员权限但用户不是管理员
  if (requireAdmin && user?.role !== 'admin') {
    return <Navigate to="/forbidden" replace />
  }

  return <>{children}</>
}

function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/forbidden" element={<ProtectedRoute><Forbidden /></ProtectedRoute>} />
      <Route path="/" element={<ProtectedRoute><MainLayout /></ProtectedRoute>}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="tunnels" element={<TunnelList />} />
        <Route path="devices" element={<DeviceList />} />
        <Route path="traffic" element={<TrafficStats />} />
        {/* 需要管理员权限的路由 */}
        <Route path="users" element={<ProtectedRoute requireAdmin><UserList /></ProtectedRoute>} />
        <Route path="settings" element={<ProtectedRoute requireAdmin><SystemSettings /></ProtectedRoute>} />
      </Route>
    </Routes>
  )
}

export default AppRouter

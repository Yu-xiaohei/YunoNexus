import { Routes, Route, Navigate } from 'react-router-dom'
import MainLayout from './components/layout/MainLayout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import TunnelList from './pages/tunnels/TunnelList'
import DeviceList from './pages/devices/DeviceList'
import UserList from './pages/users/UserList'
import TrafficStats from './pages/traffic/TrafficStats'
import SystemSettings from './pages/settings/SystemSettings'

function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<MainLayout />}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="tunnels" element={<TunnelList />} />
        <Route path="devices" element={<DeviceList />} />
        <Route path="users" element={<UserList />} />
        <Route path="traffic" element={<TrafficStats />} />
        <Route path="settings" element={<SystemSettings />} />
      </Route>
    </Routes>
  )
}

export default AppRouter

import { useState, useEffect } from 'react'
import { Layout, Menu, theme, Dropdown, Avatar, Drawer, Button } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  DashboardOutlined,
  ApiOutlined,
  DesktopOutlined,
  UserOutlined,
  BarChartOutlined,
  SettingOutlined,
  LogoutOutlined,
  MenuOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../../store/authStore'

const { Header, Sider, Content } = Layout

function MainLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [drawerVisible, setDrawerVisible] = useState(false)
  const [isMobile, setIsMobile] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { logout, user } = useAuthStore()
  const { token: { colorBgContainer, borderRadiusLG } } = theme.useToken()

  const isAdmin = user?.role === 'admin'

  // 根据角色过滤菜单
  const menuItems = [
    { key: '/dashboard', icon: <DashboardOutlined />, label: '仪表盘' },
    { key: '/tunnels', icon: <ApiOutlined />, label: '隧道管理' },
    { key: '/devices', icon: <DesktopOutlined />, label: '设备管理' },
    ...(isAdmin ? [
      { key: '/users', icon: <UserOutlined />, label: '用户管理' },
    ] : []),
    { key: '/traffic', icon: <BarChartOutlined />, label: '流量统计' },
    ...(isAdmin ? [
      { key: '/settings', icon: <SettingOutlined />, label: '系统设置' },
    ] : []),
  ]

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
      if (window.innerWidth < 768) setCollapsed(true)
    }
    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  const handleMenuClick = (key: string) => {
    navigate(key)
    if (isMobile) setDrawerVisible(false)
  }

  const userMenuItems = [
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: logout },
  ]

  const siderContent = (
    <>
      <div className="logo">{collapsed ? 'YN' : 'YUNO Nexus'}</div>
      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[location.pathname]}
        items={menuItems}
        onClick={({ key }) => handleMenuClick(key)}
      />
    </>
  )

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {isMobile ? (
        <Drawer placement="left" onClose={() => setDrawerVisible(false)} open={drawerVisible} styles={{ body: { padding: 0 } }}>
          {siderContent}
        </Drawer>
      ) : (
        <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>{siderContent}</Sider>
      )}
      <Layout>
        <Header style={{ padding: '0 24px', background: colorBgContainer, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          {isMobile && <Button type="text" icon={<MenuOutlined />} onClick={() => setDrawerVisible(true)} />}
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Avatar style={{ backgroundColor: '#1890ff', cursor: 'pointer' }} icon={<UserOutlined />} />
          </Dropdown>
        </Header>
        <Content style={{ margin: isMobile ? '8px' : '24px 16px', padding: isMobile ? 12 : 24, background: colorBgContainer, borderRadius: borderRadiusLG, overflow: 'auto' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

export default MainLayout

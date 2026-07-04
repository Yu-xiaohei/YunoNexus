import { Button, Result } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

function Forbidden() {
  const navigate = useNavigate()
  const { logout } = useAuthStore()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: '#f0f2f5' }}>
      <Result
        status="403"
        title="访问被拒绝"
        subTitle="您的账户无权访问此系统，请联系管理员。错误码: 403"
        extra={[
          <Button type="primary" key="logout" onClick={handleLogout}>
            退出登录
          </Button>,
        ]}
      />
    </div>
  )
}

export default Forbidden

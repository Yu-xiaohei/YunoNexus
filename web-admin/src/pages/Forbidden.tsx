import { Button, Typography } from 'antd'
import { useNavigate } from 'react-router-dom'
import { StopOutlined } from '@ant-design/icons'

const { Title, Text } = Typography

function Forbidden() {
  const navigate = useNavigate()

  const handleBack = () => {
    navigate(-1)
  }

  return (
    <div style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      height: '100vh', 
      background: '#f0f2f5' 
    }}>
      <div style={{ textAlign: 'center', padding: '40px' }}>
        <StopOutlined style={{ fontSize: 64, color: '#ff4d4f', marginBottom: 24 }} />
        <Title level={3} style={{ marginBottom: 16 }}>访问被拒绝</Title>
        <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
          您的账户无权访问此系统，请联系管理员。
        </Text>
        <Text type="secondary" style={{ display: 'block', marginBottom: 24, fontSize: 12 }}>
          错误码: 403
        </Text>
        <Button type="primary" onClick={handleBack}>
          返回
        </Button>
      </div>
    </div>
  )
}

export default Forbidden

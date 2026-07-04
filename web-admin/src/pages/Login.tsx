import { useState, useEffect } from 'react'
import { Form, Input, Button, Card, message, Typography, Checkbox } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { authAPI } from '../api'
import { useAuthStore } from '../store/authStore'

const { Title } = Typography

function Login() {
  const [loading, setLoading] = useState(false)
  const [rememberMe, setRememberMe] = useState(false)
  const navigate = useNavigate()
  const { setToken, setUser } = useAuthStore()
  const [form] = Form.useForm()

  // 加载保存的凭据
  useEffect(() => {
    const savedEmail = localStorage.getItem('yunonexus_remember_email')
    const savedPassword = localStorage.getItem('yunonexus_remember_password')
    if (savedEmail && savedPassword) {
      form.setFieldsValue({ email: savedEmail, password: savedPassword })
      setRememberMe(true)
    }
  }, [form])

  const onFinish = async (values: { email: string; password: string }) => {
    setLoading(true)
    try {
      const result = await authAPI.login({
        email: values.email,
        password: values.password,
        device_name: 'Web管理界面',
        device_type: 'windows',
        device_fingerprint: 'web-' + Date.now(),
      })
      
      // 保存或清除记住的凭据
      if (rememberMe) {
        localStorage.setItem('yunonexus_remember_email', values.email)
        localStorage.setItem('yunonexus_remember_password', values.password)
      } else {
        localStorage.removeItem('yunonexus_remember_email')
        localStorage.removeItem('yunonexus_remember_password')
      }
      
      setToken(result.data.session.access_token)
      setUser(result.data.user)
      message.success('登录成功')
      navigate('/dashboard')
    } catch (error: unknown) {
      const err = error as { message?: string }
      message.error(err.message || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: '#f0f2f5' }}>
      <Card style={{ width: 400 }}>
        <Title level={2} style={{ textAlign: 'center' }}>YUNO Nexus</Title>
        <p style={{ textAlign: 'center', color: '#666', marginBottom: 24 }}>于的小窝 - 管理界面</p>
        <Form form={form} name="login" onFinish={onFinish} autoComplete="off">
          <Form.Item name="email" rules={[{ required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' }]}>
            <Input prefix={<UserOutlined />} placeholder="邮箱" size="large" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" size="large" />
          </Form.Item>
          <Form.Item>
            <Checkbox checked={rememberMe} onChange={(e) => setRememberMe(e.target.checked)}>
              记住密码
            </Checkbox>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export default Login

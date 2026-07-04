import { useState } from 'react'
import { Card, Form, Input, Switch, Button, message, Tabs } from 'antd'

function SystemSettings() {
  const [loading, setLoading] = useState(false)

  const handleSave = async () => {
    setLoading(true)
    try {
      message.success('保存成功')
    } catch (error) {
      message.error('保存失败')
    } finally {
      setLoading(false)
    }
  }

  const tabItems = [
    {
      key: 'general',
      label: '基本设置',
      children: (
        <Form layout="vertical" style={{ maxWidth: 600 }}>
          <Form.Item label="系统名称">
            <Input defaultValue="YUNO Nexus" />
          </Form.Item>
          <Form.Item label="默认最大隧道数">
            <Input defaultValue="3" type="number" />
          </Form.Item>
          <Form.Item label="会话超时(秒)">
            <Input defaultValue="7200" type="number" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" onClick={handleSave} loading={loading}>保存</Button>
          </Form.Item>
        </Form>
      ),
    },
    {
      key: 'security',
      label: '安全设置',
      children: (
        <Form layout="vertical" style={{ maxWidth: 600 }}>
          <Form.Item label="登录失败锁定">
            <Switch defaultChecked />
          </Form.Item>
          <Form.Item label="最大尝试次数">
            <Input defaultValue="20" type="number" />
          </Form.Item>
          <Form.Item label="锁定时间(分钟)">
            <Input defaultValue="5" type="number" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" onClick={handleSave} loading={loading}>保存</Button>
          </Form.Item>
        </Form>
      ),
    },
  ]

  return (
    <Card title="系统设置">
      <Tabs items={tabItems} />
    </Card>
  )
}

export default SystemSettings

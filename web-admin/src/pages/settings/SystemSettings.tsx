import { Card, Form, Input, Switch, Button, message, Tabs, Alert } from 'antd'

function SystemSettings() {
  const handleSave = async () => {
    try {
      message.success('保存成功')
    } catch (error) {
      message.error('保存失败')
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
          <Form.Item label="会话超时(秒)">
            <Input defaultValue="7200" type="number" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" onClick={handleSave}>保存</Button>
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
            <Button type="primary" onClick={handleSave}>保存</Button>
          </Form.Item>
        </Form>
      ),
    },
    {
      key: 'user-defaults',
      label: '用户默认设置',
      children: (
        <Alert 
          message="提示" 
          description="以下设置为新用户的默认值，管理员可以在用户管理中为每个用户单独调整。" 
          type="info" 
          showIcon 
          style={{ marginBottom: 16 }}
        />
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

import { Card, Form, Input, Switch, Button, message, Tabs, Alert, InputNumber } from 'antd'

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
            <InputNumber defaultValue={7200} min={300} max={86400} style={{ width: '100%' }} />
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
            <InputNumber defaultValue={20} min={1} max={100} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="锁定时间(分钟)">
            <InputNumber defaultValue={5} min={1} max={60} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" onClick={handleSave}>保存</Button>
          </Form.Item>
        </Form>
      ),
    },
    {
      key: 'port',
      label: '端口配置',
      children: (
        <Form layout="vertical" style={{ maxWidth: 600 }}>
          <Alert 
            message="端口范围配置" 
            description="配置隧道可用的端口范围，新隧道将从此范围中分配端口。" 
            type="info" 
            showIcon 
            style={{ marginBottom: 16 }}
          />
          <Form.Item label="起始端口">
            <InputNumber defaultValue={30000} min={1024} max={65535} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="结束端口">
            <InputNumber defaultValue={30100} min={1024} max={65535} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" onClick={handleSave}>保存</Button>
          </Form.Item>
        </Form>
      ),
    },
    {
      key: 'password',
      label: '密码策略',
      children: (
        <Form layout="vertical" style={{ maxWidth: 600 }}>
          <Alert 
            message="密码策略配置" 
            description="设置新用户的密码强度要求。" 
            type="info" 
            showIcon 
            style={{ marginBottom: 16 }}
          />
          <Form.Item label="最小密码长度">
            <InputNumber defaultValue={8} min={6} max={32} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="必须包含大写字母">
            <Switch defaultChecked />
          </Form.Item>
          <Form.Item label="必须包含小写字母">
            <Switch defaultChecked />
          </Form.Item>
          <Form.Item label="必须包含数字">
            <Switch defaultChecked />
          </Form.Item>
          <Form.Item label="必须包含特殊字符">
            <Switch defaultChecked />
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
        <>
          <Alert 
            message="提示" 
            description="以下设置为新用户的默认值，管理员可以在用户管理中为每个用户单独调整。" 
            type="info" 
            showIcon 
            style={{ marginBottom: 16 }}
          />
          <Form layout="vertical" style={{ maxWidth: 600 }}>
            <Form.Item label="默认最大隧道数">
              <InputNumber defaultValue={3} min={1} max={100} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item label="默认最大带宽(MB/s)">
              <InputNumber defaultValue={10} min={1} max={1000} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item>
              <Button type="primary" onClick={handleSave}>保存</Button>
            </Form.Item>
          </Form>
        </>
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

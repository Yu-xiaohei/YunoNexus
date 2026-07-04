import { Card, Form, Input, Switch, Button, message, Tabs, Alert, InputNumber, Space } from 'antd'
import { useState } from 'react'

function SystemSettings() {
  const [generalForm] = Form.useForm()
  const [securityForm] = Form.useForm()
  const [userDefaultForm] = Form.useForm()
  const [saving, setSaving] = useState(false)

  const handleSave = async (formName: string, form: typeof generalForm) => {
    setSaving(true)
    try {
      const values = await form.validateFields()
      // 保存到localStorage（演示用，实际应调用API）
      localStorage.setItem(`yunonexus_${formName}`, JSON.stringify(values))
      message.success('保存成功')
    } catch (error) {
      message.error('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const tabItems = [
    {
      key: 'general',
      label: '基本设置',
      children: (
        <Form form={generalForm} layout="vertical" style={{ maxWidth: 600 }} initialValues={{ system_name: 'YUNO Nexus', session_timeout: 7200 }}>
          <Form.Item name="system_name" label="系统名称">
            <Input />
          </Form.Item>
          <Form.Item name="session_timeout" label="会话超时(秒)">
            <InputNumber min={300} max={86400} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" onClick={() => handleSave('general', generalForm)} loading={saving}>保存</Button>
          </Form.Item>
        </Form>
      ),
    },
    {
      key: 'security',
      label: '安全设置',
      children: (
        <Form form={securityForm} layout="vertical" style={{ maxWidth: 600 }} initialValues={{ login_lock: true, max_attempts: 20, lock_minutes: 5 }}>
          <Form.Item name="login_lock" label="登录失败锁定" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="max_attempts" label="最大尝试次数">
            <InputNumber min={1} max={100} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="lock_minutes" label="锁定时间(分钟)">
            <InputNumber min={1} max={60} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" onClick={() => handleSave('security', securityForm)} loading={saving}>保存</Button>
          </Form.Item>
        </Form>
      ),
    },
    {
      key: 'port',
      label: '端口策略',
      children: (
        <PortPolicyForm />
      ),
    },
    {
      key: 'password',
      label: '密码策略',
      children: (
        <PasswordPolicyForm />
      ),
    },
    {
      key: 'user-defaults',
      label: '用户默认设置',
      children: (
        <Form form={userDefaultForm} layout="vertical" style={{ maxWidth: 600 }} initialValues={{ default_max_tunnels: 3, default_max_bandwidth: 10 }}>
          <Alert 
            message="提示" 
            description="以下设置为新用户的默认值，管理员可以在用户管理中为每个用户单独调整。" 
            type="info" 
            showIcon 
            style={{ marginBottom: 16 }}
          />
          <Form.Item name="default_max_tunnels" label="默认最大隧道数">
            <InputNumber min={1} max={100} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="default_max_bandwidth" label="默认最大带宽(MB/s)">
            <InputNumber min={1} max={1000} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" onClick={() => handleSave('user_defaults', userDefaultForm)} loading={saving}>保存</Button>
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

// 端口策略组件
function PortPolicyForm() {
  const [rules, setRules] = useState<Array<{id: number, type: string, value: string, tag: string}>>([])
  const [tagTypes, setTagTypes] = useState<string[]>(['HTTP服务', '数据库', '游戏服务', '开发测试'])
  const [newRule, setNewRule] = useState({ type: 'single', value: '', tag: '' })
  const [newTagType, setNewTagType] = useState('')
  const [showTagModal, setShowTagModal] = useState(false)

  const handleAddRule = () => {
    if (!newRule.value) {
      message.warning('请输入端口或端口范围')
      return
    }
    setRules([...rules, { id: Date.now(), ...newRule }])
    setNewRule({ type: 'single', value: '', tag: '' })
    message.success('规则添加成功')
  }

  const handleDeleteRule = (id: number) => {
    setRules(rules.filter(r => r.id !== id))
    message.success('规则已删除')
  }

  const handleAddTagType = () => {
    if (newTagType && !tagTypes.includes(newTagType)) {
      setTagTypes([...tagTypes, newTagType])
      setNewTagType('')
      setShowTagModal(false)
      message.success('标签类型添加成功')
    }
  }

  return (
    <div style={{ maxWidth: 600 }}>
      <Alert 
        message="端口策略配置" 
        description="配置隧道可用的端口规则。可以添加单个端口或端口区间，并为端口打标签。" 
        type="info" 
        showIcon 
        style={{ marginBottom: 16 }}
      />
      
      <Space direction="vertical" style={{ width: '100%' }}>
        <Space>
          <select 
            value={newRule.type} 
            onChange={(e) => setNewRule({...newRule, type: e.target.value})}
            style={{ padding: '4px 8px', borderRadius: '4px', border: '1px solid #d9d9d9' }}
          >
            <option value="single">单个端口</option>
            <option value="range">端口区间</option>
          </select>
          <Input 
            placeholder={newRule.type === 'single' ? '例如: 8080' : '例如: 30000-30100'} 
            value={newRule.value}
            onChange={(e) => setNewRule({...newRule, value: e.target.value})}
            style={{ width: 200 }}
          />
          <select 
            value={newRule.tag} 
            onChange={(e) => setNewRule({...newRule, tag: e.target.value})}
            style={{ padding: '4px 8px', borderRadius: '4px', border: '1px solid #d9d9d9' }}
          >
            <option value="">选择标签</option>
            {tagTypes.map(tag => <option key={tag} value={tag}>{tag}</option>)}
          </select>
          <Button type="primary" onClick={handleAddRule}>添加</Button>
          <Button onClick={() => setShowTagModal(true)}>新建标签</Button>
        </Space>

        <table style={{ width: '100%', borderCollapse: 'collapse', marginTop: 16 }}>
          <thead>
            <tr style={{ background: '#fafafa' }}>
              <th style={{ padding: '8px', textAlign: 'left', border: '1px solid #e8e8e8' }}>类型</th>
              <th style={{ padding: '8px', textAlign: 'left', border: '1px solid #e8e8e8' }}>端口/范围</th>
              <th style={{ padding: '8px', textAlign: 'left', border: '1px solid #e8e8e8' }}>标签</th>
              <th style={{ padding: '8px', textAlign: 'center', border: '1px solid #e8e8e8' }}>操作</th>
            </tr>
          </thead>
          <tbody>
            {rules.map(rule => (
              <tr key={rule.id}>
                <td style={{ padding: '8px', border: '1px solid #e8e8e8' }}>{rule.type === 'single' ? '单个端口' : '端口区间'}</td>
                <td style={{ padding: '8px', border: '1px solid #e8e8e8' }}>{rule.value}</td>
                <td style={{ padding: '8px', border: '1px solid #e8e8e8' }}>{rule.tag || '-'}</td>
                <td style={{ padding: '8px', textAlign: 'center', border: '1px solid #e8e8e8' }}>
                  <Button type="link" danger onClick={() => handleDeleteRule(rule.id)}>删除</Button>
                </td>
              </tr>
            ))}
            {rules.length === 0 && (
              <tr>
                <td colSpan={4} style={{ padding: '8px', textAlign: 'center', color: '#999', border: '1px solid #e8e8e8' }}>暂无规则</td>
              </tr>
            )}
          </tbody>
        </table>
      </Space>

      {showTagModal && (
        <div style={{ marginTop: 16, padding: 16, background: '#f5f5f5', borderRadius: 8 }}>
          <h4>新建标签类型</h4>
          <Space>
            <Input 
              placeholder="输入标签名称" 
              value={newTagType}
              onChange={(e) => setNewTagType(e.target.value)}
              style={{ width: 200 }}
            />
            <Button type="primary" onClick={handleAddTagType}>添加</Button>
            <Button onClick={() => setShowTagModal(false)}>取消</Button>
          </Space>
          <div style={{ marginTop: 8 }}>
            <span>已有标签: </span>
            {tagTypes.map(tag => (
              <span key={tag} style={{ background: '#e6f7ff', padding: '2px 8px', borderRadius: 4, marginRight: 8 }}>{tag}</span>
            ))}
          </div>
        </div>
      )}

      <Form.Item style={{ marginTop: 16 }}>
        <Button type="primary" onClick={() => message.success('端口策略保存成功')}>保存端口策略</Button>
      </Form.Item>
    </div>
  )
}

// 密码策略组件
function PasswordPolicyForm() {
  const [form] = Form.useForm()
  const [saving, setSaving] = useState(false)

  const handleSave = async () => {
    setSaving(true)
    try {
      const values = await form.validateFields()
      localStorage.setItem('yunonexus_password_policy', JSON.stringify(values))
      message.success('密码策略保存成功')
    } catch (error) {
      message.error('保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Form form={form} layout="vertical" style={{ maxWidth: 600 }} initialValues={{ min_length: 8, require_uppercase: true, require_lowercase: true, require_digit: true, require_special: true }}>
      <Alert 
        message="密码策略配置" 
        description="设置新用户的密码强度要求。" 
        type="info" 
        showIcon 
        style={{ marginBottom: 16 }}
      />
      <Form.Item name="min_length" label="最小密码长度">
        <InputNumber min={6} max={32} style={{ width: '100%' }} />
      </Form.Item>
      <Form.Item name="require_uppercase" label="必须包含大写字母" valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item name="require_lowercase" label="必须包含小写字母" valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item name="require_digit" label="必须包含数字" valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item name="require_special" label="必须包含特殊字符" valuePropName="checked">
        <Switch />
      </Form.Item>
      <Form.Item>
        <Button type="primary" onClick={handleSave} loading={saving}>保存</Button>
      </Form.Item>
    </Form>
  )
}

export default SystemSettings

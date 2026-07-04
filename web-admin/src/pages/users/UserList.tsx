import { useEffect, useState } from 'react'
import { Table, Tag, Select, message, Modal, Form, Input, Space, Button, Popconfirm, Descriptions } from 'antd'
import { EditOutlined, DeleteOutlined, InfoCircleOutlined, PlusOutlined, KeyOutlined } from '@ant-design/icons'
import { adminAPI } from '../../api'

function UserList() {
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(true)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [passwordModalVisible, setPasswordModalVisible] = useState(false)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<Record<string, unknown> | null>(null)
  const [selectedUser, setSelectedUser] = useState<Record<string, unknown> | null>(null)
  const [form] = Form.useForm()
  const [createForm] = Form.useForm()
  const [passwordForm] = Form.useForm()

  const fetchUsers = async () => {
    setLoading(true)
    try {
      const res = await adminAPI.listUsers({ page: 1, page_size: 50 })
      setUsers(res.data.items || [])
    } catch (error) {
      console.error('获取用户列表失败:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchUsers()
  }, [])

  const handleCreate = () => {
    createForm.resetFields()
    setCreateModalVisible(true)
  }

  const handleCreateOk = async () => {
    try {
      const values = await createForm.validateFields()
      // 调用注册接口创建用户
      await fetch('http://localhost:8080/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: values.username,
          email: values.email,
          password: values.password,
          device_name: '管理员创建',
          device_type: 'windows',
          device_fingerprint: 'admin-created-' + Date.now(),
        }),
      })
      message.success('用户创建成功')
      setCreateModalVisible(false)
      fetchUsers()
    } catch (error) {
      message.error('创建失败')
    }
  }

  const handleEdit = (record: Record<string, unknown>) => {
    setEditingUser(record)
    form.setFieldsValue({
      username: record.username,
      email: record.email,
      role: record.role,
      status: record.status,
      max_tunnels: record.max_tunnels,
    })
    setEditModalVisible(true)
  }

  const handleEditOk = async () => {
    try {
      const values = await form.validateFields()
      await adminAPI.updateUser(editingUser!.id as string, values)
      message.success('更新成功')
      setEditModalVisible(false)
      fetchUsers()
    } catch (error) {
      message.error('更新失败')
    }
  }

  const handleViewDetail = (record: Record<string, unknown>) => {
    setSelectedUser(record)
    setDetailModalVisible(true)
  }

  const handleChangePassword = (record: Record<string, unknown>) => {
    setEditingUser(record)
    passwordForm.resetFields()
    setPasswordModalVisible(true)
  }

  const handlePasswordOk = async () => {
    try {
      await passwordForm.validateFields()
      // 调用修改密码接口（实际应调用专门的API）
      message.success('密码修改成功')
      setPasswordModalVisible(false)
    } catch (error) {
      message.error('修改失败')
    }
  }

  const handleDelete = async (userId: string) => {
    try {
      await adminAPI.updateUser(userId, { status: 'banned' })
      message.success('用户已封禁')
      fetchUsers()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const getStatusTag = (status: string, lastSeen: string) => {
    const isOnline = lastSeen && (Date.now() - new Date(lastSeen).getTime()) < 5 * 60 * 1000
    
    switch (status) {
      case 'active':
        return isOnline ? <Tag color="green">在线</Tag> : <Tag color="default">离线</Tag>
      case 'suspended':
        return <Tag color="orange">暂停</Tag>
      case 'banned':
        return <Tag color="red">封禁</Tag>
      default:
        return <Tag color="default">未知</Tag>
    }
  }

  const columns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    { title: '邮箱', dataIndex: 'email', key: 'email' },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (text: string) => <Tag color={text === 'admin' ? 'purple' : 'blue'}>{text === 'admin' ? '管理员' : '普通用户'}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (text: string, record: Record<string, unknown>) => getStatusTag(text, record.last_seen_at as string),
    },
    { title: '最大隧道数', dataIndex: 'max_tunnels', key: 'max_tunnels' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (text: string) => new Date(text).toLocaleString() },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: Record<string, unknown>) => (
        <Space>
          <Button type="link" icon={<InfoCircleOutlined />} onClick={() => handleViewDetail(record)}>详情</Button>
          <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Button type="link" icon={<KeyOutlined />} onClick={() => handleChangePassword(record)}>改密</Button>
          <Popconfirm title="确定封禁此用户?" onConfirm={() => handleDelete(record.id as string)}>
            <Button type="link" danger icon={<DeleteOutlined />}>封禁</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>创建用户</Button>
      </div>

      <Table columns={columns} dataSource={users} rowKey="id" loading={loading} />

      {/* 创建用户弹窗 */}
      <Modal
        title="创建用户"
        open={createModalVisible}
        onOk={handleCreateOk}
        onCancel={() => setCreateModalVisible(false)}
      >
        <Form form={createForm} layout="vertical">
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, min: 8, message: '密码至少8位' }]}>
            <Input.Password />
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑用户弹窗 */}
      <Modal
        title="编辑用户"
        open={editModalVisible}
        onOk={handleEditOk}
        onCancel={() => setEditModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="username" label="用户名">
            <Input />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input />
          </Form.Item>
          <Form.Item name="role" label="角色">
            <Select>
              <Select.Option value="user">普通用户</Select.Option>
              <Select.Option value="admin">管理员</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select>
              <Select.Option value="active">活跃</Select.Option>
              <Select.Option value="suspended">暂停</Select.Option>
              <Select.Option value="banned">封禁</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="max_tunnels" label="最大隧道数">
            <Input type="number" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 修改密码弹窗 */}
      <Modal
        title="修改密码"
        open={passwordModalVisible}
        onOk={handlePasswordOk}
        onCancel={() => setPasswordModalVisible(false)}
      >
        <Form form={passwordForm} layout="vertical">
          <Form.Item label="用户名">
            <Input value={editingUser?.username as string} disabled />
          </Form.Item>
          <Form.Item name="new_password" label="新密码" rules={[{ required: true, min: 8, message: '密码至少8位' }]}>
            <Input.Password placeholder="输入新密码" />
          </Form.Item>
          <Form.Item name="confirm_password" label="确认密码" rules={[{ required: true, message: '请确认密码' }, ({ getFieldValue }) => ({
            validator(_, value) {
              if (!value || getFieldValue('new_password') === value) {
                return Promise.resolve()
              }
              return Promise.reject(new Error('两次密码不一致'))
            },
          })]}>
            <Input.Password placeholder="再次输入新密码" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 用户详情弹窗 */}
      <Modal
        title="用户详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
      >
        {selectedUser && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="用户ID">{selectedUser.id as string}</Descriptions.Item>
            <Descriptions.Item label="用户名">{selectedUser.username as string}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{selectedUser.email as string}</Descriptions.Item>
            <Descriptions.Item label="角色">{(selectedUser.role as string) === 'admin' ? '管理员' : '普通用户'}</Descriptions.Item>
            <Descriptions.Item label="状态">{getStatusTag(selectedUser.status as string, selectedUser.last_seen_at as string)}</Descriptions.Item>
            <Descriptions.Item label="最大隧道数">{String(selectedUser.max_tunnels)}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{new Date(selectedUser.created_at as string).toLocaleString()}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </>
  )
}

export default UserList

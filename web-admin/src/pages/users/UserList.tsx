import { useEffect, useState } from 'react'
import { Table, Tag, Select, message, Modal, Form, Input, Space, Button, Popconfirm, Descriptions } from 'antd'
import { EditOutlined, DeleteOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { adminAPI } from '../../api'

function UserList() {
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(true)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<Record<string, unknown> | null>(null)
  const [selectedUser, setSelectedUser] = useState<Record<string, unknown> | null>(null)
  const [form] = Form.useForm()

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

  const handleViewDetail = (record: Record<string, unknown>) => {
    setSelectedUser(record)
    setDetailModalVisible(true)
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

  const handleDelete = async (userId: string) => {
    try {
      // 调用后端删除用户接口（如果有的话）
      // 目前先用封禁状态代替
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
          <Popconfirm title="确定封禁此用户?" onConfirm={() => handleDelete(record.id as string)}>
            <Button type="link" danger icon={<DeleteOutlined />}>封禁</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Table columns={columns} dataSource={users} rowKey="id" loading={loading} />

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

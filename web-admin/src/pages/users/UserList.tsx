import { useEffect, useState } from 'react'
import { Table, Tag, Select, message, Modal, Form, Input, Space, Button, Popconfirm, Descriptions, InputNumber } from 'antd'
import { EditOutlined, DeleteOutlined, InfoCircleOutlined, PlusOutlined, KeyOutlined, FilterOutlined } from '@ant-design/icons'
import { adminAPI } from '../../api'

function UserList() {
  const [users, setUsers] = useState([])
  const [filteredUsers, setFilteredUsers] = useState([])
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
  const [filterRole, setFilterRole] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [pageSize] = useState(10)

  const fetchUsers = async () => {
    setLoading(true)
    try { const res = await adminAPI.listUsers({ page: 1, page_size: 100 }); setUsers(res.data.items || []) } catch { console.error('获取用户列表失败') } finally { setLoading(false) }
  }

  useEffect(() => { fetchUsers() }, [])

  useEffect(() => {
    let result = users
    if (filterRole !== 'all') result = result.filter((u: Record<string, unknown>) => u.role === filterRole)
    if (filterStatus !== 'all') {
      if (filterStatus === 'online') result = result.filter((u: Record<string, unknown>) => u.status === 'active' && u.last_seen_at && (Date.now() - new Date(u.last_seen_at as string).getTime()) < 5 * 60 * 1000)
      else if (filterStatus === 'offline') result = result.filter((u: Record<string, unknown>) => u.status === 'active' && (!u.last_seen_at || (Date.now() - new Date(u.last_seen_at as string).getTime()) >= 5 * 60 * 1000))
      else result = result.filter((u: Record<string, unknown>) => u.status === filterStatus)
    }
    setFilteredUsers(result)
  }, [filterRole, filterStatus, users])

  const handleCreate = () => { createForm.resetFields(); setCreateModalVisible(true) }
  const handleCreateOk = async () => {
    try {
      const values = await createForm.validateFields()
      await fetch('http://localhost:8080/api/v1/auth/register', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ username: values.username, email: values.email, password: values.password, device_name: '管理员创建', device_type: 'windows', device_fingerprint: 'admin-created-' + Date.now() }) })
      message.success('用户创建成功'); setCreateModalVisible(false); fetchUsers()
    } catch { message.error('创建失败') }
  }
  const handleEdit = (record: Record<string, unknown>) => { setEditingUser(record); form.setFieldsValue({ username: record.username, email: record.email, role: record.role, status: record.status, max_tunnels: record.max_tunnels }); setEditModalVisible(true) }
  const handleEditOk = async () => { try { const values = await form.validateFields(); await adminAPI.updateUser(editingUser!.id as string, values); message.success('更新成功'); setEditModalVisible(false); fetchUsers() } catch { message.error('更新失败') } }
  const handleViewDetail = (record: Record<string, unknown>) => { setSelectedUser(record); setDetailModalVisible(true) }
  const handleChangePassword = (record: Record<string, unknown>) => { setEditingUser(record); passwordForm.resetFields(); setPasswordModalVisible(true) }
  const handlePasswordOk = async () => { try { await passwordForm.validateFields(); message.success('密码修改成功'); setPasswordModalVisible(false) } catch { message.error('修改失败') } }
  const handleDelete = async (userId: string) => { try { await adminAPI.deleteUser(userId); message.success('用户已删除'); fetchUsers() } catch { message.error('删除失败') } }
  const getStatusTag = (status: string, lastSeen: string) => {
    const isOnline = lastSeen && (Date.now() - new Date(lastSeen).getTime()) < 5 * 60 * 1000
    switch (status) { case 'active': return isOnline ? <Tag color="green">在线</Tag> : <Tag color="default">离线</Tag>; case 'suspended': return <Tag color="orange">暂停</Tag>; case 'banned': return <Tag color="red">封禁</Tag>; default: return <Tag color="default">未知</Tag> }
  }

  const columns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    { title: '邮箱', dataIndex: 'email', key: 'email' },
    { title: '角色', dataIndex: 'role', key: 'role', render: (t: string) => <Tag color={t === 'admin' ? 'purple' : 'blue'}>{t === 'admin' ? '管理员' : '普通用户'}</Tag> },
    { title: '状态', dataIndex: 'status', key: 'status', render: (t: string, r: Record<string, unknown>) => getStatusTag(t, r.last_seen_at as string) },
    { title: '最大隧道数', dataIndex: 'max_tunnels', key: 'max_tunnels' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (t: string) => new Date(t).toLocaleString() },
    { title: '操作', key: 'action', render: (_: unknown, r: Record<string, unknown>) => (
      <Space>
        <Button type="link" icon={<InfoCircleOutlined />} onClick={() => handleViewDetail(r)}>详情</Button>
        <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(r)}>编辑</Button>
        <Button type="link" icon={<KeyOutlined />} onClick={() => handleChangePassword(r)}>改密</Button>
        <Popconfirm title="确定删除此用户?" onConfirm={() => handleDelete(r.id as string)}><Button type="link" danger icon={<DeleteOutlined />}>删除</Button></Popconfirm>
      </Space>
    )},
  ]

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Space>
          <FilterOutlined style={{ color: '#999' }} />
          <Select value={filterRole} onChange={setFilterRole} style={{ width: 120 }}><Select.Option value="all">全部角色</Select.Option><Select.Option value="admin">管理员</Select.Option><Select.Option value="user">普通用户</Select.Option></Select>
          <Select value={filterStatus} onChange={setFilterStatus} style={{ width: 100 }}><Select.Option value="all">全部状态</Select.Option><Select.Option value="online">在线</Select.Option><Select.Option value="offline">离线</Select.Option><Select.Option value="suspended">暂停</Select.Option><Select.Option value="banned">封禁</Select.Option></Select>
          <span style={{ color: '#999' }}>共 {filteredUsers.length} 人</span>
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>创建用户</Button>
      </div>
      <Table columns={columns} dataSource={filteredUsers} rowKey="id" loading={loading} pagination={{ pageSize, showSizeChanger: false }} />
      <Modal title="创建用户" open={createModalVisible} onOk={handleCreateOk} onCancel={() => setCreateModalVisible(false)}>
        <Form form={createForm} layout="vertical">
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, type: 'email' }]}><Input /></Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, min: 8 }]}><Input.Password /></Form.Item>
        </Form>
      </Modal>
      <Modal title="编辑用户" open={editModalVisible} onOk={handleEditOk} onCancel={() => setEditModalVisible(false)}>
        <Form form={form} layout="vertical">
          <Form.Item name="username" label="用户名"><Input /></Form.Item>
          <Form.Item name="email" label="邮箱"><Input /></Form.Item>
          <Form.Item name="role" label="角色"><Select><Select.Option value="user">普通用户</Select.Option><Select.Option value="admin">管理员</Select.Option></Select></Form.Item>
          <Form.Item name="status" label="状态"><Select><Select.Option value="active">活跃</Select.Option><Select.Option value="suspended">暂停</Select.Option><Select.Option value="banned">封禁</Select.Option></Select></Form.Item>
          <Form.Item name="max_tunnels" label="最大隧道数"><InputNumber style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>
      <Modal title="修改密码" open={passwordModalVisible} onOk={handlePasswordOk} onCancel={() => setPasswordModalVisible(false)}>
        <Form form={passwordForm} layout="vertical">
          <Form.Item label="用户名"><Input value={editingUser?.username as string} disabled /></Form.Item>
          <Form.Item name="new_password" label="新密码" rules={[{ required: true, min: 8 }]}><Input.Password placeholder="输入新密码" /></Form.Item>
          <Form.Item name="confirm_password" label="确认密码" rules={[{ required: true }, ({ getFieldValue }) => ({ validator(_, value) { if (!value || getFieldValue('new_password') === value) return Promise.resolve(); return Promise.reject(new Error('两次密码不一致')) } })]}><Input.Password placeholder="再次输入新密码" /></Form.Item>
        </Form>
      </Modal>
      <Modal title="用户详情" open={detailModalVisible} onCancel={() => setDetailModalVisible(false)} footer={null}>
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

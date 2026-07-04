import { useEffect, useState } from 'react'
import { Table, Select, message } from 'antd'
import { adminAPI } from '../../api'

function UserList() {
  const [users, setUsers] = useState([])
  const [loading, setLoading] = useState(true)

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

  const handleRoleChange = async (userId: string, role: string) => {
    try {
      await adminAPI.updateUser(userId, { role })
      message.success('更新成功')
      fetchUsers()
    } catch (error) {
      message.error('更新失败')
    }
  }

  const handleStatusChange = async (userId: string, status: string) => {
    try {
      await adminAPI.updateUser(userId, { status })
      message.success('更新成功')
      fetchUsers()
    } catch (error) {
      message.error('更新失败')
    }
  }

  const columns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    { title: '邮箱', dataIndex: 'email', key: 'email' },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (text: string, record: Record<string, unknown>) => (
        <Select value={text} style={{ width: 100 }} onChange={(value) => handleRoleChange(record.id as string, value)}>
          <Select.Option value="user">普通用户</Select.Option>
          <Select.Option value="admin">管理员</Select.Option>
        </Select>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (text: string, record: Record<string, unknown>) => (
        <Select value={text} style={{ width: 100 }} onChange={(value) => handleStatusChange(record.id as string, value)}>
          <Select.Option value="active">活跃</Select.Option>
          <Select.Option value="suspended">暂停</Select.Option>
          <Select.Option value="banned">封禁</Select.Option>
        </Select>
      ),
    },
    { title: '最大隧道数', dataIndex: 'max_tunnels', key: 'max_tunnels' },
    { title: '创建时间', dataIndex: 'created_at', key: 'created_at', render: (text: string) => new Date(text).toLocaleString() },
  ]

  return (
    <Table columns={columns} dataSource={users} rowKey="id" loading={loading} />
  )
}

export default UserList

import { useEffect, useState } from 'react'
import { Table, Tag, Button, Space, Popconfirm, message, Modal, Form, Input } from 'antd'
import { DeleteOutlined, EditOutlined } from '@ant-design/icons'
import { deviceAPI } from '../../api'

function DeviceList() {
  const [devices, setDevices] = useState([])
  const [loading, setLoading] = useState(true)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [editingDevice, setEditingDevice] = useState<Record<string, unknown> | null>(null)
  const [form] = Form.useForm()

  const fetchDevices = async () => {
    setLoading(true)
    try {
      const res = await deviceAPI.list({ page: 1, page_size: 50 })
      setDevices(res.data.items || [])
    } catch (error) {
      console.error('获取设备列表失败:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchDevices()
  }, [])

  const handleEdit = (record: Record<string, unknown>) => {
    setEditingDevice(record)
    form.setFieldsValue({ device_name: record.device_name })
    setEditModalVisible(true)
  }

  const handleEditOk = async () => {
    try {
      const values = await form.validateFields()
      await deviceAPI.update(editingDevice!.id as string, values)
      message.success('更新成功')
      setEditModalVisible(false)
      fetchDevices()
    } catch (error) {
      message.error('更新失败')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await deviceAPI.delete(id)
      message.success('吊销成功')
      fetchDevices()
    } catch (error) {
      message.error('吊销失败')
    }
  }

  const columns = [
    { title: '设备名称', dataIndex: 'device_name', key: 'device_name' },
    { title: '设备类型', dataIndex: 'device_type', key: 'device_type', render: (text: string) => <Tag>{text}</Tag> },
    { title: '指纹', dataIndex: 'fingerprint', key: 'fingerprint', ellipsis: true },
    { 
      title: '状态', 
      dataIndex: 'status', 
      key: 'status',
      render: (text: string) => <Tag color={text === 'active' ? 'green' : 'red'}>{text === 'active' ? '活跃' : '已吊销'}</Tag>
    },
    { title: '最后在线', dataIndex: 'last_seen_at', key: 'last_seen_at', render: (text: string) => text ? new Date(text).toLocaleString() : '-' },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: Record<string, unknown>) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Popconfirm title="确定吊销此设备?" onConfirm={() => handleDelete(record.id as string)}>
            <Button type="link" danger icon={<DeleteOutlined />}>吊销</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <Table columns={columns} dataSource={devices} rowKey="id" loading={loading} />

      <Modal
        title="编辑设备"
        open={editModalVisible}
        onOk={handleEditOk}
        onCancel={() => setEditModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="device_name" label="设备名称">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}

export default DeviceList

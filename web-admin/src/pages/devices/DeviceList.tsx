import { useEffect, useState } from 'react'
import { Table, Tag, Button, Space, Popconfirm, message, Modal, Form, Input, Descriptions, Select } from 'antd'
import { DeleteOutlined, EditOutlined, InfoCircleOutlined, FilterOutlined } from '@ant-design/icons'
import { deviceAPI } from '../../api'

function DeviceList() {
  const [devices, setDevices] = useState([])
  const [filteredDevices, setFilteredDevices] = useState([])
  const [loading, setLoading] = useState(true)
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [editingDevice, setEditingDevice] = useState<Record<string, unknown> | null>(null)
  const [selectedDevice, setSelectedDevice] = useState<Record<string, unknown> | null>(null)
  const [form] = Form.useForm()
  const [filterType, setFilterType] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [pageSize] = useState(10)

  const fetchDevices = async () => {
    setLoading(true)
    try {
      const res = await deviceAPI.list({ page: 1, page_size: 100 })
      setDevices(res.data.items || [])
    } catch (error) {
      console.error('获取设备列表失败:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchDevices() }, [])

  useEffect(() => {
    let result = devices
    if (filterType !== 'all') result = result.filter((d: Record<string, unknown>) => d.device_type === filterType)
    if (filterStatus !== 'all') result = result.filter((d: Record<string, unknown>) => d.status === filterStatus)
    setFilteredDevices(result)
  }, [filterType, filterStatus, devices])

  const handleEdit = (record: Record<string, unknown>) => { setEditingDevice(record); form.setFieldsValue({ device_name: record.device_name }); setEditModalVisible(true) }
  const handleViewDetail = (record: Record<string, unknown>) => { setSelectedDevice(record); setDetailModalVisible(true) }
  const handleEditOk = async () => { try { const values = await form.validateFields(); await deviceAPI.update(editingDevice!.id as string, values); message.success('更新成功'); setEditModalVisible(false); fetchDevices() } catch { message.error('更新失败') } }
  const handleDelete = async (id: string) => { try { await deviceAPI.delete(id); message.success('吊销成功'); fetchDevices() } catch { message.error('吊销失败') } }

  const columns = [
    { title: '设备名称', dataIndex: 'device_name', key: 'device_name' },
    { title: '设备类型', dataIndex: 'device_type', key: 'device_type', render: (t: string) => <Tag>{t}</Tag> },
    { title: '指纹', dataIndex: 'fingerprint', key: 'fingerprint', ellipsis: true },
    { title: '状态', dataIndex: 'status', key: 'status', render: (t: string) => <Tag color={t === 'active' ? 'green' : 'red'}>{t === 'active' ? '活跃' : '已吊销'}</Tag> },
    { title: '最后在线', dataIndex: 'last_seen_at', key: 'last_seen_at', render: (t: string) => t ? new Date(t).toLocaleString() : '-' },
    { title: '操作', key: 'action', render: (_: unknown, r: Record<string, unknown>) => (
      <Space>
        <Button type="link" icon={<InfoCircleOutlined />} onClick={() => handleViewDetail(r)}>详情</Button>
        <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(r)}>编辑</Button>
        <Popconfirm title="确定吊销此设备?" onConfirm={() => handleDelete(r.id as string)}><Button type="link" danger icon={<DeleteOutlined />}>吊销</Button></Popconfirm>
      </Space>
    )},
  ]

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Space>
          <FilterOutlined style={{ color: '#999' }} />
          <Select value={filterType} onChange={setFilterType} style={{ width: 120 }}><Select.Option value="all">全部类型</Select.Option><Select.Option value="windows">Windows</Select.Option><Select.Option value="linux">Linux</Select.Option><Select.Option value="android">Android</Select.Option></Select>
          <Select value={filterStatus} onChange={setFilterStatus} style={{ width: 100 }}><Select.Option value="all">全部状态</Select.Option><Select.Option value="active">活跃</Select.Option><Select.Option value="revoked">已吊销</Select.Option></Select>
          <span style={{ color: '#999' }}>共 {filteredDevices.length} 台</span>
        </Space>
      </div>
      <Table columns={columns} dataSource={filteredDevices} rowKey="id" loading={loading} pagination={{ pageSize, showSizeChanger: false }} />
      <Modal title="编辑设备" open={editModalVisible} onOk={handleEditOk} onCancel={() => setEditModalVisible(false)}>
        <Form form={form} layout="vertical"><Form.Item name="device_name" label="设备名称"><Input /></Form.Item></Form>
      </Modal>
      <Modal title="设备详情" open={detailModalVisible} onCancel={() => setDetailModalVisible(false)} footer={null}>
        {selectedDevice && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="设备ID">{selectedDevice.id as string}</Descriptions.Item>
            <Descriptions.Item label="设备名称">{selectedDevice.device_name as string}</Descriptions.Item>
            <Descriptions.Item label="设备类型">{selectedDevice.device_type as string}</Descriptions.Item>
            <Descriptions.Item label="指纹">{selectedDevice.fingerprint as string}</Descriptions.Item>
            <Descriptions.Item label="状态">{(selectedDevice.status as string) === 'active' ? '活跃' : '已吊销'}</Descriptions.Item>
            <Descriptions.Item label="最后在线">{selectedDevice.last_seen_at ? new Date(selectedDevice.last_seen_at as string).toLocaleString() : '从未在线'}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{new Date(selectedDevice.created_at as string).toLocaleString()}</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </>
  )
}

export default DeviceList

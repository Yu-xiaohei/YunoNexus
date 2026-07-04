import { useEffect, useState } from 'react'
import { Table, Button, Tag, Space, Modal, Form, Input, Select, message, Popconfirm, InputNumber, Descriptions } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, PlayCircleOutlined, PauseCircleOutlined, InfoCircleOutlined, FilterOutlined } from '@ant-design/icons'
import { tunnelAPI } from '../../api'

function TunnelList() {
  const [tunnels, setTunnels] = useState([])
  const [filteredTunnels, setFilteredTunnels] = useState([])
  const [loading, setLoading] = useState(true)
  const [modalVisible, setModalVisible] = useState(false)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [editingTunnel, setEditingTunnel] = useState<Record<string, unknown> | null>(null)
  const [selectedTunnel, setSelectedTunnel] = useState<Record<string, unknown> | null>(null)
  const [form] = Form.useForm()
  const [filterProtocol, setFilterProtocol] = useState<string>('all')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [pageSize] = useState(10)

  const fetchTunnels = async () => {
    setLoading(true)
    try {
      const res = await tunnelAPI.list({ page: 1, page_size: 100 })
      setTunnels(res.data.items || [])
    } catch (error) {
      console.error('获取隧道列表失败:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchTunnels() }, [])

  useEffect(() => {
    let result = tunnels
    if (filterProtocol !== 'all') result = result.filter((t: Record<string, unknown>) => t.protocol === filterProtocol)
    if (filterStatus !== 'all') result = result.filter((t: Record<string, unknown>) => t.status === filterStatus)
    setFilteredTunnels(result)
  }, [filterProtocol, filterStatus, tunnels])

  const handleCreate = () => { setEditingTunnel(null); form.resetFields(); setModalVisible(true) }
  const handleEdit = (record: Record<string, unknown>) => { setEditingTunnel(record); form.setFieldsValue(record); setModalVisible(true) }
  const handleViewDetail = async (record: Record<string, unknown>) => {
    try { const res = await tunnelAPI.stats(record.id as string); setSelectedTunnel({ ...record, stats: res.data }); setDetailModalVisible(true) } catch { message.error('获取详情失败') }
  }
  const handleDelete = async (id: string) => { try { await tunnelAPI.delete(id); message.success('删除成功'); fetchTunnels() } catch { message.error('删除失败') } }
  const handleStart = async (id: string) => { try { await tunnelAPI.start(id); message.success('启动成功'); fetchTunnels() } catch { message.error('启动失败') } }
  const handleStop = async (id: string) => { try { await tunnelAPI.stop(id); message.success('停止成功'); fetchTunnels() } catch { message.error('停止失败') } }
  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      if (editingTunnel) { await tunnelAPI.update(editingTunnel.id as string, values); message.success('更新成功') }
      else { await tunnelAPI.create(values); message.success('创建成功') }
      setModalVisible(false); fetchTunnels()
    } catch { message.error('操作失败') }
  }

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '协议', dataIndex: 'protocol', key: 'protocol', render: (t: string) => <Tag color="blue">{t.toUpperCase()}</Tag> },
    { title: '本地地址', key: 'local', render: (_: unknown, r: Record<string, unknown>) => `${r.local_host}:${r.local_port}` },
    { title: '远程端口', dataIndex: 'remote_port', key: 'remote_port' },
    { title: '状态', dataIndex: 'status', key: 'status', render: (t: string) => <Tag color={t === 'active' ? 'green' : 'default'}>{t === 'active' ? '已启动' : '离线'}</Tag> },
    { title: '操作', key: 'action', render: (_: unknown, r: Record<string, unknown>) => (
      <Space>
        <Button type="link" icon={<InfoCircleOutlined />} onClick={() => handleViewDetail(r)}>详情</Button>
        {r.status === 'active' ? <Button type="link" icon={<PauseCircleOutlined />} onClick={() => handleStop(r.id as string)}>停止</Button> : <Button type="link" icon={<PlayCircleOutlined />} onClick={() => handleStart(r.id as string)}>启动</Button>}
        <Button type="link" icon={<EditOutlined />} onClick={() => handleEdit(r)}>编辑</Button>
        <Popconfirm title="确定删除?" onConfirm={() => handleDelete(r.id as string)}><Button type="link" danger icon={<DeleteOutlined />}>删除</Button></Popconfirm>
      </Space>
    )},
  ]

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Space>
          <FilterOutlined style={{ color: '#999' }} />
          <Select value={filterProtocol} onChange={setFilterProtocol} style={{ width: 120 }}><Select.Option value="all">全部协议</Select.Option><Select.Option value="tcp">TCP</Select.Option><Select.Option value="udp">UDP</Select.Option><Select.Option value="http">HTTP</Select.Option><Select.Option value="https">HTTPS</Select.Option></Select>
          <Select value={filterStatus} onChange={setFilterStatus} style={{ width: 100 }}><Select.Option value="all">全部状态</Select.Option><Select.Option value="active">已启动</Select.Option><Select.Option value="inactive">离线</Select.Option></Select>
          <span style={{ color: '#999' }}>共 {filteredTunnels.length} 条</span>
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>创建隧道</Button>
      </div>
      <Table columns={columns} dataSource={filteredTunnels} rowKey="id" loading={loading} pagination={{ pageSize, showSizeChanger: false }} />
      <Modal title={editingTunnel ? '编辑隧道' : '创建隧道'} open={modalVisible} onOk={handleModalOk} onCancel={() => setModalVisible(false)}>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="protocol" label="协议" rules={[{ required: true }]}><Select><Select.Option value="tcp">TCP</Select.Option><Select.Option value="udp">UDP</Select.Option><Select.Option value="http">HTTP</Select.Option><Select.Option value="https">HTTPS</Select.Option></Select></Form.Item>
          <Form.Item name="local_host" label="本地地址" rules={[{ required: true }]}><Input defaultValue="127.0.0.1" /></Form.Item>
          <Form.Item name="local_port" label="本地端口" rules={[{ required: true }]}><InputNumber style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="remote_port" label="远程端口"><InputNumber style={{ width: '100%' }} placeholder="留空自动分配" /></Form.Item>
          <Form.Item name="domain" label="域名"><Input placeholder="HTTP/HTTPS协议需要" /></Form.Item>
        </Form>
      </Modal>
      <Modal title="隧道详情" open={detailModalVisible} onCancel={() => setDetailModalVisible(false)} footer={null}>
        {selectedTunnel && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="名称">{selectedTunnel.name as string}</Descriptions.Item>
            <Descriptions.Item label="协议">{(selectedTunnel.protocol as string).toUpperCase()}</Descriptions.Item>
            <Descriptions.Item label="本地地址">{selectedTunnel.local_host as string}:{selectedTunnel.local_port as number}</Descriptions.Item>
            <Descriptions.Item label="远程端口">{selectedTunnel.remote_port as number}</Descriptions.Item>
            <Descriptions.Item label="状态">{selectedTunnel.status === 'active' ? '已启动' : '离线'}</Descriptions.Item>
            <Descriptions.Item label="活跃连接数">{String((selectedTunnel.stats as Record<string, unknown>)?.active_connections || 0)}</Descriptions.Item>
            <Descriptions.Item label="已发送流量">{String((selectedTunnel.stats as Record<string, unknown>)?.bytes_sent || 0)} bytes</Descriptions.Item>
            <Descriptions.Item label="已接收流量">{String((selectedTunnel.stats as Record<string, unknown>)?.bytes_recv || 0)} bytes</Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </>
  )
}

export default TunnelList

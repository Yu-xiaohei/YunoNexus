import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic, Table, Tag, Spin } from 'antd'
import { UserOutlined, ApiOutlined, DesktopOutlined, ThunderboltOutlined, ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { adminAPI, tunnelAPI, trafficAPI } from '../api'

function Dashboard() {
  const [stats, setStats] = useState<Record<string, number>>({})
  const [tunnels, setTunnels] = useState([])
  const [traffic, setTraffic] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [statsRes, tunnelsRes, trafficRes] = await Promise.all([
          adminAPI.systemStats(),
          tunnelAPI.list({ page: 1, page_size: 5 }),
          trafficAPI.overview(),
        ])
        setStats(statsRes.data)
        setTunnels(tunnelsRes.data.items || [])
        setTraffic(trafficRes.data)
      } catch (error) {
        console.error('获取数据失败:', error)
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

  const formatBytes = (bytes: number) => {
    if (!bytes) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const tunnelColumns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '协议', dataIndex: 'protocol', key: 'protocol', render: (text: string) => <Tag color="blue">{text.toUpperCase()}</Tag> },
    { title: '端口', dataIndex: 'remote_port', key: 'remote_port' },
    { 
      title: '状态', 
      dataIndex: 'status', 
      key: 'status',
      render: (text: string) => <Tag color={text === 'active' ? 'green' : 'default'}>{text === 'active' ? '在线' : '离线'}</Tag>
    },
  ]

  return (
    <Spin spinning={loading}>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable onClick={() => navigate('/users')} style={{ cursor: 'pointer' }}>
            <Statistic title="总用户数" value={stats.total_users || 0} prefix={<UserOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable onClick={() => navigate('/tunnels')} style={{ cursor: 'pointer' }}>
            <Statistic title="总隧道数" value={stats.total_tunnels || 0} prefix={<ApiOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable onClick={() => navigate('/devices')} style={{ cursor: 'pointer' }}>
            <Statistic title="在线设备" value={stats.active_users || 0} prefix={<DesktopOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable onClick={() => navigate('/tunnels')} style={{ cursor: 'pointer' }}>
            <Statistic title="活跃隧道" value={stats.active_tunnels || 0} prefix={<ThunderboltOutlined />} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12}>
          <Card hoverable onClick={() => navigate('/traffic')} style={{ cursor: 'pointer' }}>
            <Statistic
              title="总发送流量"
              value={formatBytes(traffic.total_sent || 0)}
              prefix={<ArrowUpOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12}>
          <Card hoverable onClick={() => navigate('/traffic')} style={{ cursor: 'pointer' }}>
            <Statistic
              title="总接收流量"
              value={formatBytes(traffic.total_recv || 0)}
              prefix={<ArrowDownOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <Card 
        title="最近隧道" 
        style={{ marginTop: 16 }}
        extra={<a onClick={() => navigate('/tunnels')}>查看全部</a>}
      >
        <Table 
          columns={tunnelColumns} 
          dataSource={tunnels} 
          rowKey="id"
          pagination={false}
          onRow={() => ({
            onClick: () => navigate('/tunnels'),
            style: { cursor: 'pointer' }
          })}
        />
      </Card>
    </Spin>
  )
}

export default Dashboard

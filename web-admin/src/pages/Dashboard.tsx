import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic, Table, Tag } from 'antd'
import { UserOutlined, ApiOutlined, DesktopOutlined, ThunderboltOutlined } from '@ant-design/icons'
import { adminAPI, tunnelAPI } from '../api'

function Dashboard() {
  const [stats, setStats] = useState<Record<string, number>>({})
  const [tunnels, setTunnels] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [statsRes, tunnelsRes] = await Promise.all([
          adminAPI.systemStats(),
          tunnelAPI.list({ page: 1, page_size: 5 }),
        ])
        setStats(statsRes.data)
        setTunnels(tunnelsRes.data.items || [])
      } catch (error) {
        console.error('获取数据失败:', error)
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [])

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
    <div>
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card>
            <Statistic title="总用户数" value={stats.total_users || 0} prefix={<UserOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="总隧道数" value={stats.total_tunnels || 0} prefix={<ApiOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="在线设备" value={stats.active_users || 0} prefix={<DesktopOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="活跃隧道" value={stats.active_tunnels || 0} prefix={<ThunderboltOutlined />} />
          </Card>
        </Col>
      </Row>

      <Card title="最近隧道" style={{ marginTop: 16 }}>
        <Table 
          columns={tunnelColumns} 
          dataSource={tunnels} 
          rowKey="id"
          loading={loading}
          pagination={false}
        />
      </Card>
    </div>
  )
}

export default Dashboard

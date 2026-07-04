import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic, Select, Spin } from 'antd'
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons'
import { trafficAPI } from '../../api'

function TrafficStats() {
  const [stats, setStats] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('24h')

  useEffect(() => {
    const fetchStats = async () => {
      setLoading(true)
      try {
        const res = await trafficAPI.overview()
        setStats(res.data)
      } catch (error) {
        console.error('获取流量统计失败:', error)
      } finally {
        setLoading(false)
      }
    }
    fetchStats()
  }, [timeRange])

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const timeRangeOptions = [
    { value: '15m', label: '最近15分钟' },
    { value: '1h', label: '最近1小时' },
    { value: '6h', label: '最近6小时' },
    { value: '24h', label: '最近24小时' },
    { value: '3d', label: '最近3天' },
    { value: '30d', label: '最近30天' },
  ]

  return (
    <Spin spinning={loading}>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="总发送流量"
              value={formatBytes(stats.total_sent || 0)}
              prefix={<ArrowUpOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="总接收流量"
              value={formatBytes(stats.total_recv || 0)}
              prefix={<ArrowDownOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="总连接数"
              value={stats.total_connections || 0}
            />
          </Card>
        </Col>
      </Row>

      <Card title="流量趋势" style={{ marginTop: 16 }}>
        <div style={{ marginBottom: 16 }}>
          <Select 
            value={timeRange} 
            onChange={setTimeRange}
            options={timeRangeOptions}
            style={{ width: 200 }}
          />
        </div>
        <div style={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#999' }}>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 24, marginBottom: 8 }}>📊</div>
            <div>流量趋势图</div>
            <div style={{ fontSize: 12 }}>选择时间范围查看趋势</div>
          </div>
        </div>
      </Card>
    </Spin>
  )
}

export default TrafficStats

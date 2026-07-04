import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic } from 'antd'
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons'
import { trafficAPI } from '../../api'

function TrafficStats() {
  const [stats, setStats] = useState<Record<string, number>>({})
  const [_loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchStats = async () => {
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
  }, [])

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col span={8}>
          <Card>
            <Statistic
              title="总发送流量"
              value={formatBytes(stats.total_sent || 0)}
              prefix={<ArrowUpOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="总接收流量"
              value={formatBytes(stats.total_recv || 0)}
              prefix={<ArrowDownOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="总连接数"
              value={stats.total_connections || 0}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default TrafficStats

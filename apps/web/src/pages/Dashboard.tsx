import React, { useState, useEffect } from 'react'
import { Card, Row, Col, Statistic, Progress, Table, Typography, Spin, Alert } from 'antd'
import {
  AppstoreOutlined,
  DownloadOutlined,
  WarningOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons'
import axios from 'axios'

const { Title } = Typography

interface Stats {
  total_skills: number
  total_installations: number
  pending_reviews: number
  active_users: number
}

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [recentSkills, setRecentSkills] = useState<any[]>([])

  useEffect(() => {
    fetchStats()
    fetchRecentSkills()
  }, [])

  const fetchStats = async () => {
    try {
      const token = localStorage.getItem('access_token')
      const response = await axios.get('/api/admin/v1/stats', {
        headers: { Authorization: `Bearer ${token}` },
      })
      setStats(response.data)
    } catch (error) {
      console.error('Failed to fetch stats:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchRecentSkills = async () => {
    try {
      const token = localStorage.getItem('access_token')
      const response = await axios.get('/api/v1/skills?limit=5', {
        headers: { Authorization: `Bearer ${token}` },
      })
      setRecentSkills(response.data.skills || [])
    } catch (error) {
      console.error('Failed to fetch skills:', error)
    }
  }

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    )
  }

  const skillColumns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
    { title: '分类', dataIndex: 'category', key: 'category' },
    { title: '安装次数', dataIndex: 'install_count', key: 'install_count' },
  ]

  return (
    <div>
      <Title level={2}>仪表盘</Title>

      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card bordered={false}>
            <Statistic
              title="Skill 总数"
              value={stats?.total_skills || 0}
              prefix={<AppstoreOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card bordered={false}>
            <Statistic
              title="总安装次数"
              value={stats?.total_installations || 0}
              prefix={<DownloadOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card bordered={false}>
            <Statistic
              title="待审核"
              value={stats?.pending_reviews || 0}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card bordered={false}>
            <Statistic
              title="活跃用户"
              value={stats?.active_users || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={16}>
          <Card title="最近 Skill" bordered={false}>
            <Table
              columns={skillColumns}
              dataSource={recentSkills}
              rowKey="id"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card title="风险分布" bordered={false}>
            <div style={{ padding: '16px 0' }}>
              <div style={{ marginBottom: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span>低风险 (A)</span>
                  <span>70%</span>
                </div>
                <Progress percent={70} strokeColor="#52c41a" showInfo={false} />
              </div>
              <div style={{ marginBottom: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span>中风险 (B-C)</span>
                  <span>25%</span>
                </div>
                <Progress percent={25} strokeColor="#faad14" showInfo={false} />
              </div>
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span>高风险 (D-F)</span>
                  <span>5%</span>
                </div>
                <Progress percent={5} strokeColor="#ff4d4f" showInfo={false} />
              </div>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard

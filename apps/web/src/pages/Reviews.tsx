import React, { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Space, Typography, message, Modal, Form, Input, Badge, Descriptions } from 'antd'
import { CheckCircleOutlined, CloseCircleOutlined, EyeOutlined } from '@ant-design/icons'
import axios from 'axios'

const { Title, Text } = Typography

interface Review {
  id: string
  scan_id: string
  status: string
  decision?: string
  comment: string
  risk_level?: string
  risk_score?: number
  skill_id: string
  skill_name: string
  created_at: string
}

const Reviews: React.FC = () => {
  const [reviews, setReviews] = useState<Review[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedReview, setSelectedReview] = useState<Review | null>(null)
  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false)
  const [isDecisionModalOpen, setIsDecisionModalOpen] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchReviews()
  }, [])

  const fetchReviews = async () => {
    try {
      const token = localStorage.getItem('access_token')
      const response = await axios.get('/api/v1/reviews?status=pending', {
        headers: { Authorization: `Bearer ${token}` },
      })
      setReviews(response.data.reviews || [])
    } catch (error) {
      message.error('加载审核列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleDecision = async (values: { decision: string; comment: string }) => {
    if (!selectedReview) return

    try {
      const token = localStorage.getItem('access_token')
      await axios.put(`/api/v1/reviews/${selectedReview.id}`, values, {
        headers: { Authorization: `Bearer ${token}` },
      })
      message.success('审核决定已保存')
      setIsDecisionModalOpen(false)
      form.resetFields()
      fetchReviews()
    } catch (error) {
      message.error('提交失败')
    }
  }

  const getRiskLevelColor = (level?: string) => {
    const colors: Record<string, string> = {
      A: 'green',
      B: 'lime',
      C: 'orange',
      D: 'red',
      F: 'purple',
    }
    return colors[level || ''] || 'default'
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'orange',
      approved: 'green',
      rejected: 'red',
      escalated: 'purple',
    }
    return colors[status] || 'default'
  }

  const columns = [
    {
      title: 'Skill 名称',
      dataIndex: 'skill_name',
      key: 'skill_name',
    },
    {
      title: '风险等级',
      key: 'risk_level',
      render: (_: any, record: Review) => (
        <Tag color={getRiskLevelColor(record.risk_level)}>{record.risk_level || '待扫描'}</Tag>
      ),
    },
    {
      title: '风险评分',
      dataIndex: 'risk_score',
      key: 'risk_score',
      render: (score?: number) => {
        let color = 'green'
        if (score && score > 50) color = 'orange'
        if (score && score > 70) color = 'red'
        return <Text style={{ color }}>{score ?? '-'}</Text>
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Badge status={getStatusColor(status) === 'green' ? 'success' : getStatusColor(status) === 'red' ? 'error' : 'processing'} text={status} />
      ),
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Review) => (
        <Space>
          <Button
            type="link"
            icon={<EyeOutlined />}
            size="small"
            onClick={() => {
              setSelectedReview(record)
              setIsDetailModalOpen(true)
            }}
          >
            详情
          </Button>
          {record.status === 'pending' && (
            <Button
              type="primary"
              size="small"
              onClick={() => {
                setSelectedReview(record)
                setIsDecisionModalOpen(true)
              }}
            >
              审批
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Title level={2}>审批中心</Title>

      <Card bordered={false}>
        <Table
          columns={columns}
          dataSource={reviews}
          loading={loading}
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </Card>

      {/* 详情弹窗 */}
      <Modal
        title="审核详情"
        open={isDetailModalOpen}
        onCancel={() => setIsDetailModalOpen(false)}
        footer={[
          <Button key="approve" type="primary" onClick={() => { setIsDetailModalOpen(false); setIsDecisionModalOpen(true); }}>
            进行审批
          </Button>,
          <Button key="close" onClick={() => setIsDetailModalOpen(false)}>关闭</Button>,
        ]}
      >
        {selectedReview && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="Skill 名称">{selectedReview.skill_name}</Descriptions.Item>
            <Descriptions.Item label="风险等级">
              <Tag color={getRiskLevelColor(selectedReview.risk_level)}>{selectedReview.risk_level}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="风险评分">{selectedReview.risk_score}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Badge status={getStatusColor(selectedReview.status) === 'green' ? 'success' : 'processing'} text={selectedReview.status} />
            </Descriptions.Item>
            <Descriptions.Item label="申请时间">
              {new Date(selectedReview.created_at).toLocaleString('zh-CN')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* 审批弹窗 */}
      <Modal
        title="审批决定"
        open={isDecisionModalOpen}
        onCancel={() => setIsDecisionModalOpen(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleDecision}>
          <Form.Item name="decision" label="决定" rules={[{ required: true }]}>
            <Select>
              <Select.Option value="approved">
                <CheckCircleOutlined style={{ color: '#52c41a' }} /> 通过
              </Select.Option>
              <Select.Option value="rejected">
                <CloseCircleOutlined style={{ color: '#ff4d4f' }} /> 拒绝
              </Select.Option>
              <Select.Option value="escalated">
                升级审核
              </Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="comment" label="审批意见">
            <Input.TextArea rows={4} placeholder="填写审批意见..." />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">提交</Button>
              <Button onClick={() => setIsDecisionModalOpen(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Reviews

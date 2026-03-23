import React, { useState } from 'react'
import { Card, Table, Input, DatePicker, Select, Space, Typography, Button } from 'antd'
import { SearchOutlined, ExportOutlined } from '@ant-design/icons'

const { Title } = Typography
const { RangePicker } = DatePicker

const AuditLogs: React.FC = () => {
  const [filters, setFilters] = useState({
    eventType: '',
    dateRange: undefined as any,
  })

  // Mock data
  const mockLogs = [
    { id: 1, event_type: 'SKILL_INSTALL', actor: 'zhangsan', resource: 'pdf-processor', result: 'SUCCESS', time: new Date().toISOString() },
    { id: 2, event_type: 'SKILL_UPLOAD', actor: 'admin', resource: 'excel-analyzer', result: 'SUCCESS', time: new Date().toISOString() },
    { id: 3, event_type: 'REVIEW_APPROVED', actor: 'security_admin', resource: 'code-generator', result: 'SUCCESS', time: new Date().toISOString() },
    { id: 4, event_type: 'SKILL_INSTALL', actor: 'lisi', resource: 'unauthorized-skill', result: 'BLOCKED', time: new Date().toISOString() },
  ]

  const columns = [
    { title: '事件类型', dataIndex: 'event_type', key: 'event_type' },
    { title: '操作人', dataIndex: 'actor', key: 'actor' },
    { title: '操作对象', dataIndex: 'resource', key: 'resource' },
    { title: '结果', dataIndex: 'result', key: 'result' },
    {
      title: '时间',
      dataIndex: 'time',
      key: 'time',
      render: (time: string) => new Date(time).toLocaleString('zh-CN'),
    },
  ]

  return (
    <div>
      <Title level={2}>审计日志</Title>

      <Card bordered={false}>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16, flexWrap: 'wrap', gap: 16 }}>
          <Space wrap>
            <Input
              placeholder="搜索..."
              prefix={<SearchOutlined />}
              style={{ width: 200 }}
            />
            <RangePicker />
            <Select
              placeholder="事件类型"
              style={{ width: 150 }}
              allowClear
            >
              <Select.Option value="SKILL_INSTALL">Skill 安装</Select.Option>
              <Select.Option value="SKILL_UPLOAD">Skill 上传</Select.Option>
              <Select.Option value="REVIEW_APPROVED">审批通过</Select.Option>
              <Select.Option value="REVIEW_REJECTED">审批拒绝</Select.Option>
              <Select.Option value="BLOCKED">拦截事件</Select.Option>
            </Select>
          </Space>
          <Button icon={<ExportOutlined />}>导出日志</Button>
        </div>

        <Table
          columns={columns}
          dataSource={mockLogs}
          rowKey="id"
          pagination={{ pageSize: 20, showSizeChanger: true }}
        />
      </Card>
    </div>
  )
}

export default AuditLogs

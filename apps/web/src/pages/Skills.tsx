import React, { useState, useEffect } from 'react'
import { Card, Table, Button, Tag, Input, Space, Modal, Form, Select, Typography, message, Badge } from 'antd'
import { PlusOutlined, SearchOutlined, EyeOutlined, DeleteOutlined } from '@ant-design/icons'
import axios from 'axios'

const { Title } = Typography

interface Skill {
  id: string
  name: string
  description: string
  category?: string
  tags: string[]
  source_type: string
  status: string
  install_count: number
  created_at: string
}

const Skills: React.FC = () => {
  const [skills, setSkills] = useState<Skill[]>([])
  const [loading, setLoading] = useState(true)
  const [searchText, setSearchText] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchSkills()
  }, [])

  const fetchSkills = async () => {
    try {
      const token = localStorage.getItem('access_token')
      const response = await axios.get('/api/v1/skills', {
        headers: { Authorization: `Bearer ${token}` },
      })
      setSkills(response.data.skills || [])
    } catch (error) {
      message.error('加载 Skill 列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = async (values: any) => {
    try {
      const token = localStorage.getItem('access_token')
      await axios.post('/api/v1/skills', values, {
        headers: { Authorization: `Bearer ${token}` },
      })
      message.success('Skill 创建成功')
      setIsModalOpen(false)
      form.resetFields()
      fetchSkills()
    } catch (error: any) {
      message.error(error.response?.data?.error || '创建失败')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      const token = localStorage.getItem('access_token')
      await axios.delete(`/api/v1/skills/${id}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      message.success('Skill 已删除')
      fetchSkills()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      active: 'green',
      deprecated: 'gray',
      blacklisted: 'red',
    }
    return colors[status] || 'default'
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: Skill) => (
        <div>
          <strong>{name}</strong>
          <br />
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>{record.description}</Typography.Text>
        </div>
      ),
    },
    {
      title: '分类',
      dataIndex: 'category',
      key: 'category',
      render: (category?: string) => category || '-',
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      render: (tags: string[]) => (
        <Space size={4}>
          {tags?.slice(0, 3).map((tag) => (
            <Tag key={tag} color="blue">{tag}</Tag>
          ))}
          {tags?.length > 3 && <Tag>+{tags.length - 3}</Tag>}
        </Space>
      ),
    },
    {
      title: '来源',
      dataIndex: 'source_type',
      key: 'source_type',
      render: (type: string) => (
        <Tag color={type === 'internal' ? 'green' : 'orange'}>
          {type === 'internal' ? '内部' : '开源'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Badge status={status === 'active' ? 'success' : status === 'blacklisted' ? 'error' : 'default'} text={status} />
      ),
    },
    {
      title: '安装次数',
      dataIndex: 'install_count',
      key: 'install_count',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Skill) => (
        <Space>
          <Button type="link" icon={<EyeOutlined />} size="small" />
          <Button
            type="link"
            danger
            icon={<DeleteOutlined />}
            size="small"
            onClick={() => handleDelete(record.id)}
          />
        </Space>
      ),
    },
  ]

  const filteredSkills = skills.filter(
    (skill) =>
      skill.name.toLowerCase().includes(searchText.toLowerCase()) ||
      skill.description.toLowerCase().includes(searchText.toLowerCase())
  )

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={2} style={{ margin: 0 }}>Skill 仓库</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsModalOpen(true)}>
          发布 Skill
        </Button>
      </div>

      <Card bordered={false}>
        <div style={{ marginBottom: 16 }}>
          <Input
            placeholder="搜索 Skill..."
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ maxWidth: 300 }}
          />
        </div>
        <Table
          columns={columns}
          dataSource={filteredSkills}
          loading={loading}
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title="发布 Skill"
        open={isModalOpen}
        onCancel={() => setIsModalOpen(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input placeholder="例如：pdf-processor" />
          </Form.Item>
          <Form.Item name="display_name" label="显示名称">
            <Input placeholder="例如：PDF 处理器" />
          </Form.Item>
          <Form.Item name="description" label="描述" rules={[{ required: true }]}>
            <Input.TextArea rows={3} placeholder="描述 Skill 的功能" />
          </Form.Item>
          <Form.Item name="category" label="分类">
            <Select>
              <Select.Option value="document">文档处理</Select.Option>
              <Select.Option value="analysis">数据分析</Select.Option>
              <Select.Option value="code">代码生成</Select.Option>
              <Select.Option value="api">API 集成</Select.Option>
              <Select.Option value="other">其他</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item name="source_type" label="来源类型" initialValue="internal">
            <Select>
              <Select.Option value="internal">内部开发</Select.Option>
              <Select.Option value="opensource">开源引入</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">
              发布
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Skills

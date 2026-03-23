import React from 'react'
import { Card, Form, Input, Switch, Button, message, Divider, Typography, InputNumber } from 'antd'
import { SaveOutlined } from '@ant-design/icons'

const { Title } = Typography

const Settings: React.FC = () => {
  const [form] = Form.useForm()

  const handleSave = async (values: any) => {
    try {
      const token = localStorage.getItem('access_token')
      // await axios.put('/api/admin/v1/settings', values, {
      //   headers: { Authorization: `Bearer ${token}` },
      // })
      message.success('设置已保存')
    } catch (error) {
      message.error('保存失败')
    }
  }

  return (
    <div>
      <Title level={2}>系统设置</Title>

      <Card title="扫描配置" bordered={false} style={{ marginBottom: 16 }}>
        <Form form={form} layout="vertical" onFinish={handleSave} initialValues={{
          autoApproveMaxScore: 30,
          sandboxTimeout: 120000,
          maxConcurrent: 5,
          oidcEnabled: false,
        }}>
          <Form.Item
            name="autoApproveMaxScore"
            label="自动审批最大风险分"
            tooltip="风险评分低于此值的 Skill 将自动审批通过"
          >
            <InputNumber min={0} max={100} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="sandboxTimeout"
            label="沙箱超时时间 (ms)"
          >
            <InputNumber min={30000} max={300000} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="maxConcurrent"
            label="最大并发扫描数"
          >
            <InputNumber min={1} max={20} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item>
            <Button type="primary" icon={<SaveOutlined />} htmlType="submit">
              保存配置
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card title="认证配置" bordered={false} style={{ marginBottom: 16 }}>
        <Form layout="vertical">
          <Form.Item label="OIDC/SSO 集成" valuePropName="checked">
            <Switch disabled />
          </Form.Item>
          <Form.Item label="OIDC Issuer">
            <Input disabled placeholder="https://sso.company.com" />
          </Form.Item>
        </Form>
      </Card>

      <Card title="通知配置" bordered={false}>
        <Form layout="vertical">
          <Form.Item label="SMTP 服务器">
            <Input placeholder="smtp.company.com" />
          </Form.Item>
          <Form.Item label="发件人邮箱">
            <Input placeholder="noreply@company.com" />
          </Form.Item>
          <Form.Item label="Slack Webhook URL">
            <Input placeholder="https://hooks.slack.com/..." />
          </Form.Item>
          <Form.Item>
            <Button type="primary">保存</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export default Settings

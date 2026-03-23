import React, { useState } from 'react'
import { Form, Input, Button, Card, message, Typography } from 'antd'
import { LockOutlined, UserOutlined } from '@ant-design/icons'
import axios from 'axios'

const { Title } = Typography

interface LoginProps {
  onLogin: (user: any, token: string) => void
}

interface LoginForm {
  username: string
  password: string
}

const Login: React.FC<LoginProps> = ({ onLogin }) => {
  const [loading, setLoading] = useState(false)

  const onFinish = async (values: LoginForm) => {
    setLoading(true)
    try {
      const response = await axios.post('/api/v1/auth/login', values)
      const { access_token, user } = response.data
      onLogin(user, access_token)
    } catch (error: any) {
      message.error(error.response?.data?.error || '登录失败，请检查用户名和密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      height: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    }}>
      <Card style={{ width: 400, borderRadius: 8 }} bordered={false}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Title level={2}>SkillsHub Enterprise</Title>
          <Typography.Text type="secondary">AI Skills 私有仓库管理平台</Typography.Text>
        </div>
        <Form onFinish={onFinish} layout="vertical" size="large">
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="用户名"
              autoComplete="username"
            />
          </Form.Item>
          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="密码"
              autoComplete="current-password"
            />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block size="large">
              登录
            </Button>
          </Form.Item>
        </Form>
        <Typography.Text type="secondary" style={{ fontSize: 12 }}>
          默认账号：admin / admin123
        </Typography.Text>
      </Card>
    </div>
  )
}

export default Login

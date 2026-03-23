import React, { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Layout, message } from 'antd'
import {
  DashboardOutlined,
  AppstoreOutlined,
  CheckCircleOutlined,
  FileTextOutlined,
  SettingOutlined,
} from '@ant-design/icons'

import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Skills from './pages/Skills'
import Reviews from './pages/Reviews'
import AuditLogs from './pages/AuditLogs'
import Settings from './pages/Settings'
import SiderMenu from './components/SiderMenu'
import TopBar from './components/TopBar'

const { Content } = Layout

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [user, setUser] = useState<any>(null)
  const [collapsed, setCollapsed] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    const userData = localStorage.getItem('user')
    if (token && userData) {
      setIsAuthenticated(true)
      setUser(JSON.parse(userData))
    }
  }, [])

  const handleLogin = (userData: any, token: string) => {
    localStorage.setItem('access_token', token)
    localStorage.setItem('user', JSON.stringify(userData))
    setIsAuthenticated(true)
    setUser(userData)
    message.success('登录成功')
  }

  const handleLogout = () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('user')
    setIsAuthenticated(false)
    setUser(null)
    message.success('已退出登录')
  }

  if (!isAuthenticated) {
    return <Login onLogin={handleLogin} />
  }

  return (
    <BrowserRouter>
      <Layout style={{ minHeight: '100vh' }}>
        <SiderMenu collapsed={collapsed} />
        <Layout>
          <TopBar
            user={user}
            collapsed={collapsed}
            onToggle={() => setCollapsed(!collapsed)}
            onLogout={handleLogout}
          />
          <Content style={{ padding: '24px', background: '#f0f2f5' }}>
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/skills" element={<Skills />} />
              <Route path="/reviews" element={<Reviews />} />
              <Route path="/audit-logs" element={<AuditLogs />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Content>
        </Layout>
      </Layout>
    </BrowserRouter>
  )
}

export default App

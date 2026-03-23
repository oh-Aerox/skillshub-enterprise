import React from 'react'
import { Layout, Dropdown, Avatar, Button } from 'antd'
import { MenuFoldOutlined, MenuUnfoldOutlined, BellOutlined } from '@ant-design/icons'

const { Header } = Layout

interface TopBarProps {
  user: any
  collapsed: boolean
  onToggle: () => void
  onLogout: () => void
}

const TopBar: React.FC<TopBarProps> = ({ user, collapsed, onToggle, onLogout }) => {
  const menuItems = [
    { key: 'profile', label: '个人信息' },
    { key: 'logout', label: '退出登录', onClick: onLogout },
  ]

  return (
    <Header
      style={{
        padding: '0 24px',
        background: '#fff',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        boxShadow: '0 1px 4px rgba(0,21,41,.08)',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
        <Button type="text" icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={onToggle} />
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
        <Button type="text" icon={<BellOutlined />} />
        <Dropdown menu={{ items: menuItems }} placement="bottomRight" arrow>
          <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
            <Avatar style={{ backgroundColor: '#1890ff' }}>{user?.username?.[0]?.toUpperCase()}</Avatar>
            <span>{user?.username}</span>
          </div>
        </Dropdown>
      </div>
    </Header>
  )
}

export default TopBar

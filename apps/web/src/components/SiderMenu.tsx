import React from 'react'
import { Layout, Menu, type MenuProps } from 'antd'
import {
  DashboardOutlined,
  AppstoreOutlined,
  CheckCircleOutlined,
  FileTextOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'

const { Sider } = Layout

type MenuItem = {
  key: string
  icon: React.ReactNode
  label: string
}

const items: MenuItem[] = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: '/skills', icon: <AppstoreOutlined />, label: 'Skill 仓库' },
  { key: '/reviews', icon: <CheckCircleOutlined />, label: '审批中心' },
  { key: '/audit-logs', icon: <FileTextOutlined />, label: '审计日志' },
  { key: '/settings', icon: <SettingOutlined />, label: '系统设置' },
]

interface SiderMenuProps {
  collapsed: boolean
}

const SiderMenu: React.FC<SiderMenuProps> = ({ collapsed }) => {
  const navigate = useNavigate()
  const location = useLocation()

  const handleClick: MenuProps['onClick'] = (e) => {
    navigate(e.key)
  }

  return (
    <Sider
      collapsible
      collapsed={collapsed}
      onCollapse={() => {}}
      width={200}
      style={{ background: '#001529' }}
    >
      <div
        style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'white',
          fontSize: collapsed ? 14 : 18,
          fontWeight: 'bold',
        }}
      >
        {collapsed ? 'SH' : 'SkillsHub'}
      </div>
      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[location.pathname]}
        items={items}
        onClick={handleClick}
      />
    </Sider>
  )
}

export default SiderMenu

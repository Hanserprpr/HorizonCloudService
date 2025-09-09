import React from 'react';
import { Layout, Menu, Button, Dropdown, Avatar, Breadcrumb } from 'antd';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  DashboardOutlined,
  FileOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  ProfileOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';

const { Header, Sider, Content } = Layout;

const SimpleMainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [collapsed, setCollapsed] = React.useState(false);

  // 菜单项配置
  const menuItems: MenuProps['items'] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: '/files',
      icon: <FileOutlined />,
      label: '文件管理',
    },
    {
      key: '/users',
      icon: <UserOutlined />,
      label: '用户管理',
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: '系统设置',
    },
  ];

  // 用户下拉菜单
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <ProfileOutlined />,
      label: '个人资料',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
    },
  ];

  // 处理菜单点击
  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  // 处理用户菜单点击
  const handleUserMenuClick = ({ key }: { key: string }) => {
    switch (key) {
      case 'profile':
        navigate('/profile');
        break;
      case 'logout':
        localStorage.removeItem('simple-auth');
        navigate('/auth/login');
        break;
    }
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={240}
        style={{
          boxShadow: '2px 0 6px rgba(0, 21, 41, 0.1)',
        }}
      >
        <div style={{
          height: '64px',
          padding: '16px',
          textAlign: 'center',
          borderBottom: '1px solid #f0f0f0',
          marginBottom: '16px',
        }}>
          <h2 style={{
            color: '#1677ff',
            fontWeight: 600,
            margin: 0,
            fontSize: '18px',
          }}>
            {collapsed ? 'CMS' : '云存储管理后台'}
          </h2>
        </div>
        
        <Menu
          theme="light"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>

      <Layout>
        <Header style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          background: '#fff',
          padding: '0 24px',
          boxShadow: '0 1px 4px rgba(0, 21, 41, 0.08)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              style={{
                fontSize: '16px',
                width: 64,
                height: 64,
              }}
            />

            <Breadcrumb style={{ marginLeft: 16 }}>
              <Breadcrumb.Item>首页</Breadcrumb.Item>
              {location.pathname !== '/dashboard' && (
                <Breadcrumb.Item>
                  {location.pathname === '/files' ? '文件管理' :
                   location.pathname === '/users' ? '用户管理' :
                   location.pathname === '/settings' ? '系统设置' :
                   location.pathname === '/profile' ? '个人资料' : '未知页面'}
                </Breadcrumb.Item>
              )}
            </Breadcrumb>
          </div>

          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Dropdown
              menu={{
                items: userMenuItems,
                onClick: handleUserMenuClick,
              }}
              placement="bottomRight"
              arrow
            >
              <div style={{
                display: 'flex',
                alignItems: 'center',
                padding: '8px 12px',
                borderRadius: '4px',
                cursor: 'pointer',
                transition: 'background-color 0.3s',
              }}>
                <Avatar
                  size="small"
                  icon={<UserOutlined />}
                  style={{ marginRight: 8 }}
                />
                <span>系统管理员</span>
              </div>
            </Dropdown>
          </div>
        </Header>

        <Content style={{
          margin: 0,
          padding: 24,
          background: '#f0f2f5',
          minHeight: 'calc(100vh - 64px)',
        }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default SimpleMainLayout;
import React, { useEffect, useState } from 'react';
import { Layout, Menu, Button, Dropdown, Avatar, Badge, Breadcrumb } from 'antd';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  DashboardOutlined,
  FileOutlined,
  UserOutlined,
  SettingOutlined,
  BellOutlined,
  LogoutOutlined,
  ProfileOutlined,
} from '@ant-design/icons';
import { useUIStore } from '@stores/uiStore';
import { ROUTES, APP_NAME, LAYOUT_CONFIG } from '@constants/index';
import type { MenuProps } from 'antd';

const { Header, Sider, Content } = Layout;

const StableMainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  
  const { 
    sidebarCollapsed, 
    setSidebarCollapsed, 
    toggleSidebar,
    breadcrumb,
    setBreadcrumb,
    notificationCount 
  } = useUIStore();

  // 简化的响应式检测
  useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth < 768;
      setIsMobile(mobile);
      if (mobile) {
        setSidebarCollapsed(true);
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [setSidebarCollapsed]);

  // 菜单项配置
  const menuItems: MenuProps['items'] = [
    {
      key: ROUTES.DASHBOARD,
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: ROUTES.FILES,
      icon: <FileOutlined />,
      label: '文件管理',
    },
    {
      key: ROUTES.USERS,
      icon: <UserOutlined />,
      label: '用户管理',
    },
    {
      key: ROUTES.SETTINGS,
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

  // 处理用户菜单点击 - 简化版本
  const handleUserMenuClick = ({ key }: { key: string }) => {
    switch (key) {
      case 'profile':
        navigate(ROUTES.PROFILE);
        break;
      case 'logout':
        // 简化的退出逻辑
        localStorage.removeItem('simple-auth');
        navigate('/auth/login');
        break;
    }
  };

  // 更新面包屑
  useEffect(() => {
    const updateBreadcrumb = () => {
      const pathMap: Record<string, string> = {
        [ROUTES.DASHBOARD]: '仪表盘',
        [ROUTES.FILES]: '文件管理',
        [ROUTES.USERS]: '用户管理',
        [ROUTES.SETTINGS]: '系统设置',
        [ROUTES.PROFILE]: '个人资料',
      };

      const currentPath = location.pathname;
      const breadcrumbItems = [
        { title: '首页', path: ROUTES.DASHBOARD },
      ];

      if (currentPath !== ROUTES.DASHBOARD) {
        const title = pathMap[currentPath] || '未知页面';
        breadcrumbItems.push({ title });
      }

      setBreadcrumb(breadcrumbItems);
    };

    updateBreadcrumb();
  }, [location.pathname, setBreadcrumb]);

  const siderWidth = sidebarCollapsed 
    ? LAYOUT_CONFIG.SIDEBAR_COLLAPSED_WIDTH 
    : LAYOUT_CONFIG.SIDEBAR_WIDTH;

  return (
    <Layout className="main-layout">
      <Sider
        trigger={null}
        collapsible
        collapsed={sidebarCollapsed}
        width={LAYOUT_CONFIG.SIDEBAR_WIDTH}
        collapsedWidth={isMobile ? 0 : LAYOUT_CONFIG.SIDEBAR_COLLAPSED_WIDTH}
        className="main-sider"
        style={{
          position: isMobile ? 'fixed' : 'relative',
          height: '100vh',
          left: 0,
          top: 0,
          zIndex: isMobile ? 1000 : 1,
        }}
      >
        <div className="logo">
          {!sidebarCollapsed ? (
            <h2>{APP_NAME}</h2>
          ) : (
            <h2>CMS</h2>
          )}
        </div>
        
        <Menu
          theme="light"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>

      {/* 移动端遮罩 */}
      {isMobile && !sidebarCollapsed && (
        <div
          className="mobile-mask"
          onClick={() => setSidebarCollapsed(true)}
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: 'rgba(0, 0, 0, 0.45)',
            zIndex: 999,
          }}
        />
      )}

      <Layout
        className="site-layout"
        style={{
          marginLeft: isMobile ? 0 : siderWidth,
          transition: 'margin-left 0.2s',
        }}
      >
        <Header className="site-layout-header" style={{ padding: '0 24px' }}>
          <div className="header-left">
            <Button
              type="text"
              icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={toggleSidebar}
              style={{
                fontSize: '16px',
                width: 64,
                height: 64,
              }}
            />

            <Breadcrumb style={{ marginLeft: 16 }}>
              {breadcrumb.map((item, index) => (
                <Breadcrumb.Item 
                  key={index}
                  onClick={item.path ? () => navigate(item.path!) : undefined}
                  style={{ cursor: item.path ? 'pointer' : 'default' }}
                >
                  {item.title}
                </Breadcrumb.Item>
              ))}
            </Breadcrumb>
          </div>

          <div className="header-right">
            <Badge count={notificationCount} size="small">
              <Button
                type="text"
                icon={<BellOutlined />}
                style={{ marginRight: 16 }}
              />
            </Badge>

            <Dropdown
              menu={{
                items: userMenuItems,
                onClick: handleUserMenuClick,
              }}
              placement="bottomRight"
              arrow
            >
              <div className="user-info" style={{ cursor: 'pointer' }}>
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

        <Content
          className="site-layout-content"
          style={{
            margin: 0,
            padding: 24,
            minHeight: 280,
          }}
        >
          <Outlet />
        </Content>
      </Layout>

      <style jsx>{`
        .main-layout {
          min-height: 100vh;
        }

        .main-sider {
          box-shadow: 2px 0 6px rgba(0, 21, 41, 0.1);
        }

        .logo {
          height: 64px;
          padding: 16px;
          text-align: center;
          border-bottom: 1px solid #f0f0f0;
          margin-bottom: 16px;
        }

        .logo h2 {
          color: #1677ff;
          font-weight: 600;
          margin: 0;
          font-size: 18px;
          line-height: 32px;
        }

        .site-layout-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          background: #fff;
          padding: 0 24px;
          box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
        }

        .header-left {
          display: flex;
          align-items: center;
        }

        .header-right {
          display: flex;
          align-items: center;
        }

        .user-info {
          display: flex;
          align-items: center;
          padding: 8px 12px;
          border-radius: 4px;
          transition: background-color 0.3s;
        }

        .user-info:hover {
          background-color: #f5f5f5;
        }

        .site-layout-content {
          background: #f0f2f5;
          min-height: calc(100vh - 64px);
        }

        /* 响应式样式 */
        @media (max-width: 768px) {
          .site-layout-header {
            padding: 0 16px;
          }

          .header-left {
            flex: 1;
          }

          .header-right {
            flex-shrink: 0;
          }

          .user-info span {
            display: none;
          }

          .site-layout-content {
            padding: 16px;
          }

          .logo {
            padding: 16px 8px;
          }

          .logo h2 {
            font-size: 16px;
          }
        }

        @media (max-width: 480px) {
          .site-layout-header {
            padding: 0 12px;
          }

          .site-layout-content {
            padding: 12px;
          }
        }
      `}</style>
    </Layout>
  );
};

export default StableMainLayout;
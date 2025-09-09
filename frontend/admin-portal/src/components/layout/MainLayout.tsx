import React, { useEffect } from 'react';
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
import { useAuth } from '@hooks/useAuth';
import { useUIStore } from '@stores/uiStore';
import { useResponsive } from '@hooks/useResponsive';
import { ROUTES, APP_NAME, LAYOUT_CONFIG } from '@constants/index';
import type { MenuProps } from 'antd';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuth();
  const { isMobile } = useResponsive();
  const { 
    sidebarCollapsed, 
    setSidebarCollapsed, 
    toggleSidebar,
    breadcrumb,
    setBreadcrumb,
    notificationCount 
  } = useUIStore();

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

  // 处理用户菜单点击
  const handleUserMenuClick = ({ key }: { key: string }) => {
    switch (key) {
      case 'profile':
        navigate(ROUTES.PROFILE);
        break;
      case 'logout':
        logout();
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
        breadcrumbItems.push({ title, path: currentPath });
      }

      setBreadcrumb(breadcrumbItems);
    };

    updateBreadcrumb();
  }, [location.pathname, setBreadcrumb]);

  // 响应式处理侧边栏
  useEffect(() => {
    if (isMobile) {
      setSidebarCollapsed(true);
    }
  }, [isMobile, setSidebarCollapsed]);

  const siderWidth = sidebarCollapsed 
    ? LAYOUT_CONFIG.SIDEBAR_COLLAPSED_WIDTH 
    : LAYOUT_CONFIG.SIDEBAR_WIDTH;

  return (
    <div className="main-layout-wrapper">
      <div
        className="main-sider"
        style={{
          width: siderWidth,
          minWidth: siderWidth,
          maxWidth: siderWidth,
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          zIndex: isMobile ? 1000 : 1,
          backgroundColor: '#fff',
          borderRight: '1px solid #f0f0f0',
          display: isMobile && sidebarCollapsed ? 'none' : 'block',
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
          inlineCollapsed={sidebarCollapsed}
        />
      </div>

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

      <div
        className="main-content"
        style={{
          marginLeft: isMobile ? 0 : siderWidth,
          minHeight: '100vh',
        }}
      >
        <div 
          className="site-layout-header" 
          style={{ 
            padding: '0 24px',
            height: '64px',
            backgroundColor: '#fff',
            borderBottom: '1px solid #f0f0f0',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
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
                {!isMobile && <span>{user?.display_name || user?.username}</span>}
              </div>
            </Dropdown>
          </div>
        </div>

        <div
          className="site-layout-content"
          style={{
            margin: 0,
            padding: 24,
            minHeight: 'calc(100vh - 64px)',
            background: '#f0f2f5',
          }}
        >
          <Outlet />
        </div>
      </div>

      <style>{`
        .main-layout-wrapper {
          min-height: 100vh;
          background: #f0f2f5;
        }

        .main-sider {
          box-shadow: 2px 0 6px rgba(0, 21, 41, 0.1);
          overflow-y: auto;
          transition: width 0.2s, min-width 0.2s, max-width 0.2s;
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

        .main-content {
          transition: margin-left 0.2s;
        }

        .site-layout-header {
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

        /* 响应式样式 */
        @media (max-width: 768px) {
          .site-layout-header {
            padding: 0 16px !important;
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
            padding: 16px !important;
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
            padding: 0 12px !important;
          }

          .site-layout-content {
            padding: 12px !important;
          }
        }
      `}</style>
    </div>
  );
};

export default MainLayout;
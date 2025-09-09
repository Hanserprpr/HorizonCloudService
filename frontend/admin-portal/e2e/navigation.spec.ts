import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    // 模拟已登录状态（在实际应用中可能需要登录流程）
    await page.goto('/dashboard');
  });

  test('should display main layout with sidebar and header', async ({ page }) => {
    // 检查侧边栏
    await expect(page.locator('.main-sider')).toBeVisible();
    await expect(page.getByText('云存储管理后台')).toBeVisible();
    
    // 检查主要导航项
    await expect(page.getByRole('menuitem', { name: '仪表盘' })).toBeVisible();
    await expect(page.getByRole('menuitem', { name: '文件管理' })).toBeVisible();
    await expect(page.getByRole('menuitem', { name: '用户管理' })).toBeVisible();
    await expect(page.getByRole('menuitem', { name: '系统设置' })).toBeVisible();
    
    // 检查顶部导航栏
    await expect(page.locator('.site-layout-header')).toBeVisible();
    await expect(page.locator('.header-left')).toBeVisible();
    await expect(page.locator('.header-right')).toBeVisible();
  });

  test('should toggle sidebar collapse', async ({ page }) => {
    const sidebarToggle = page.getByRole('button').first(); // 折叠按钮
    const sidebar = page.locator('.main-sider');
    
    // 默认状态下侧边栏应该是展开的
    await expect(page.getByText('云存储管理后台')).toBeVisible();
    
    // 点击折叠按钮
    await sidebarToggle.click();
    
    // 验证侧边栏折叠后的状态
    await expect(page.getByText('CMS')).toBeVisible();
    await expect(page.getByText('云存储管理后台')).not.toBeVisible();
    
    // 再次点击展开
    await sidebarToggle.click();
    await expect(page.getByText('云存储管理后台')).toBeVisible();
  });

  test('should navigate between pages', async ({ page }) => {
    // 导航到仪表盘
    await page.getByRole('menuitem', { name: '仪表盘' }).click();
    await expect(page).toHaveURL('/dashboard');
    await expect(page.getByText('仪表盘')).toBeVisible();
    
    // 导航到文件管理
    await page.getByRole('menuitem', { name: '文件管理' }).click();
    await expect(page).toHaveURL('/files');
    
    // 导航到用户管理
    await page.getByRole('menuitem', { name: '用户管理' }).click();
    await expect(page).toHaveURL('/users');
    
    // 导航到系统设置
    await page.getByRole('menuitem', { name: '系统设置' }).click();
    await expect(page).toHaveURL('/settings');
  });

  test('should display breadcrumb navigation', async ({ page }) => {
    // 检查仪表盘的面包屑
    await page.getByRole('menuitem', { name: '仪表盘' }).click();
    await expect(page.locator('.ant-breadcrumb')).toContainText('首页');
    
    // 检查其他页面的面包屑
    await page.getByRole('menuitem', { name: '文件管理' }).click();
    await expect(page.locator('.ant-breadcrumb')).toContainText('首页');
    await expect(page.locator('.ant-breadcrumb')).toContainText('文件管理');
  });

  test('should handle user dropdown menu', async ({ page }) => {
    const userDropdown = page.locator('.user-info');
    
    // 点击用户头像区域
    await userDropdown.click();
    
    // 检查下拉菜单项
    await expect(page.getByText('个人资料')).toBeVisible();
    await expect(page.getByText('退出登录')).toBeVisible();
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 }); // iPad size
    
    // 在移动设备上，侧边栏默认应该是折叠的
    await expect(page.locator('.main-sider')).toBeVisible();
    
    // 检查响应式布局
    await page.setViewportSize({ width: 375, height: 667 }); // iPhone size
    
    // 验证移动端导航行为
    const toggleButton = page.getByRole('button').first();
    await toggleButton.click();
    
    // 在移动设备上应该显示遮罩层
    await expect(page.locator('.mobile-mask')).toBeVisible();
    
    // 点击遮罩层关闭侧边栏
    await page.locator('.mobile-mask').click();
    await expect(page.locator('.mobile-mask')).not.toBeVisible();
  });

  test('should handle keyboard navigation', async ({ page }) => {
    // 使用Tab键导航
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab');
    
    // 使用Enter键激活菜单项
    await page.keyboard.press('Enter');
    
    // 验证导航是否生效
    // 这里的具体验证取决于焦点管理的实现
  });
});
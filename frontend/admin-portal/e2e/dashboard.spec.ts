import { test, expect } from '@playwright/test';

test.describe('Dashboard Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/dashboard');
  });

  test('should display dashboard with all stat cards', async ({ page }) => {
    // 检查页面标题
    await expect(page.getByRole('heading', { name: '仪表盘' })).toBeVisible();
    
    // 检查统计卡片
    await expect(page.getByText('总用户数')).toBeVisible();
    await expect(page.getByText('总文件数')).toBeVisible();
    await expect(page.getByText('今日上传')).toBeVisible();
    await expect(page.getByText('存储使用率')).toBeVisible();
    
    // 检查统计数据
    await expect(page.getByText('156')).toBeVisible();
    await expect(page.getByText('2,345')).toBeVisible();
    await expect(page.getByText('23 个文件')).toBeVisible();
    await expect(page.getByText('29.04%')).toBeVisible();
  });

  test('should display quick actions section', async ({ page }) => {
    await expect(page.getByText('快速操作')).toBeVisible();
    
    // 检查快速操作按钮
    await expect(page.getByText('上传文件')).toBeVisible();
    await expect(page.getByText('新建文件夹')).toBeVisible();
    await expect(page.getByText('添加用户')).toBeVisible();
    await expect(page.getByText('系统设置')).toBeVisible();
    await expect(page.getByText('数据导出')).toBeVisible();
    await expect(page.getByText('系统监控')).toBeVisible();
  });

  test('should display storage usage chart', async ({ page }) => {
    await expect(page.getByText('存储使用情况')).toBeVisible();
    
    // 检查存储使用统计
    await expect(page.getByText('145.2 GB')).toBeVisible();
    await expect(page.getByText('500.0 GB')).toBeVisible();
    await expect(page.getByText('使用率: 29.04%')).toBeVisible();
    
    // 检查存储类型分类
    await expect(page.getByText('热存储')).toBeVisible();
    await expect(page.getByText('9.1 GB')).toBeVisible();
    await expect(page.getByText('冷存储')).toBeVisible();
    await expect(page.getByText('354.8 GB')).toBeVisible();
  });

  test('should display recent activity', async ({ page }) => {
    await expect(page.getByText('最近活动')).toBeVisible();
    
    // 检查活动列表项
    await expect(page.getByText('上传了文件')).toBeVisible();
    await expect(page.getByText('新用户注册')).toBeVisible();
    await expect(page.getByText('批量上传完成')).toBeVisible();
    await expect(page.getByText('删除了文件')).toBeVisible();
  });

  test('should handle quick action clicks', async ({ page }) => {
    // 测试上传文件按钮
    await page.getByText('上传文件').click();
    // 在实际应用中，这应该触发文件上传对话框或导航到文件页面
    
    // 测试添加用户按钮
    await page.getByText('添加用户').click();
    // 在实际应用中，这应该导航到用户管理页面或打开用户创建表单
    
    // 测试系统设置按钮
    await page.getByText('系统设置').click();
    // 在实际应用中，这应该导航到设置页面
  });

  test('should be responsive on different screen sizes', async ({ page }) => {
    // 测试平板尺寸
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.getByText('仪表盘')).toBeVisible();
    await expect(page.getByText('总用户数')).toBeVisible();
    
    // 测试手机尺寸
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.getByText('仪表盘')).toBeVisible();
    
    // 在移动设备上，统计卡片应该垂直堆叠
    const statCards = page.locator('.ant-card');
    await expect(statCards.first()).toBeVisible();
  });

  test('should update data when refreshed', async ({ page }) => {
    // 获取初始数据
    const initialUserCount = await page.getByText('156').textContent();
    
    // 刷新页面
    await page.reload();
    
    // 验证数据加载
    await expect(page.getByText('仪表盘')).toBeVisible();
    await expect(page.getByText('156')).toBeVisible();
  });

  test('should handle loading states', async ({ page }) => {
    // 在页面加载时检查loading状态
    await page.goto('/dashboard');
    
    // 检查页面是否正常加载
    await expect(page.getByText('仪表盘')).toBeVisible();
    
    // 在实际应用中，这里可以测试skeleton loading或spinner
  });
});
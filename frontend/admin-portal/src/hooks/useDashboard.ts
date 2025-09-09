import { useQuery } from '@tanstack/react-query';
import { systemService } from '@services/systemService';
import { userService } from '@services/userService';
import { fileService } from '@services/fileService';

// Dashboard 统计数据类型
export interface DashboardStats {
  users: {
    total: number;
    active: number;
    newToday: number;
    trend: number;
  };
  files: {
    total: number;
    totalSize: string;
    uploadedToday: number;
    sizeToday: string;
    trend: number;
  };
  storage: {
    used: string;
    total: string;
    usagePercent: number;
    trend: number;
  };
  system: {
    uptime: string;
    cpuUsage: number;
    memoryUsage: number;
    status: 'healthy' | 'warning' | 'error';
  };
}

// 最近活动类型
export interface RecentActivity {
  id: string;
  type: 'file_upload' | 'user_register' | 'file_delete' | 'user_login';
  title: string;
  description: string;
  user: string;
  timestamp: string;
  icon?: string;
  color?: string;
}

// 获取仪表盘统计数据
export const useDashboardStats = () => {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: async (): Promise<DashboardStats> => {
      // 模拟数据 - 实际项目中会调用真实API
      try {
        // 并行调用多个API获取数据
        const [usersResult, filesResult, systemResult] = await Promise.allSettled([
          userService.getUsers({ page: 1, limit: 1 }), // 获取用户总数
          fileService.getFiles({ page: 1, limit: 1 }), // 获取文件总数
          systemService.getSystemInfo(), // 获取系统信息
        ]);

        // 处理真实API结果或返回模拟数据
        return {
          users: {
            total: 156,
            active: 89,
            newToday: 12,
            trend: 8.5,
          },
          files: {
            total: 2345,
            totalSize: '45.6 GB',
            uploadedToday: 23,
            sizeToday: '1.2 GB',
            trend: 12.3,
          },
          storage: {
            used: '145.2 GB',
            total: '500.0 GB',
            usagePercent: 29.04,
            trend: 5.8,
          },
          system: {
            uptime: '15天 3小时',
            cpuUsage: 23,
            memoryUsage: 67,
            status: 'healthy',
          },
        };
      } catch (error) {
        console.error('获取仪表盘数据失败:', error);
        // 返回模拟数据作为后备
        return {
          users: {
            total: 156,
            active: 89,
            newToday: 12,
            trend: 8.5,
          },
          files: {
            total: 2345,
            totalSize: '45.6 GB',
            uploadedToday: 23,
            sizeToday: '1.2 GB',
            trend: 12.3,
          },
          storage: {
            used: '145.2 GB',
            total: '500.0 GB',
            usagePercent: 29.04,
            trend: 5.8,
          },
          system: {
            uptime: '15天 3小时',
            cpuUsage: 23,
            memoryUsage: 67,
            status: 'healthy',
          },
        };
      }
    },
    staleTime: 5 * 60 * 1000, // 5分钟内数据不过期
    refetchInterval: 30 * 1000, // 30秒自动刷新
    retry: 2,
  });
};

// 获取最近活动
export const useRecentActivities = () => {
  return useQuery({
    queryKey: ['dashboard', 'activities'],
    queryFn: async (): Promise<RecentActivity[]> => {
      // 模拟数据 - 实际项目中会调用真实API
      return [
        {
          id: '1',
          type: 'file_upload',
          title: '上传了新文件',
          description: 'summer-vacation.jpg (2.5MB)',
          user: '张三',
          timestamp: '2024-01-15 14:32',
          icon: 'file-image',
          color: '#52c41a',
        },
        {
          id: '2',
          type: 'user_register',
          title: '新用户注册',
          description: '用户 "李四" 注册了账户',
          user: '系统',
          timestamp: '2024-01-15 14:15',
          icon: 'user-add',
          color: '#1677ff',
        },
        {
          id: '3',
          type: 'file_upload',
          title: '批量上传完成',
          description: '家庭照片文件夹 (15个文件)',
          user: '王五',
          timestamp: '2024-01-15 13:45',
          icon: 'folder',
          color: '#722ed1',
        },
        {
          id: '4',
          type: 'user_login',
          title: '用户登录',
          description: '管理员登录系统',
          user: '管理员',
          timestamp: '2024-01-15 13:20',
          icon: 'login',
          color: '#fa8c16',
        },
        {
          id: '5',
          type: 'file_delete',
          title: '删除了文件',
          description: 'old-backup.zip',
          user: '管理员',
          timestamp: '2024-01-15 12:58',
          icon: 'delete',
          color: '#ff4d4f',
        },
      ];
    },
    staleTime: 2 * 60 * 1000, // 2分钟内数据不过期
    refetchInterval: 60 * 1000, // 1分钟自动刷新
    retry: 1,
  });
};

// 获取存储使用趋势数据 (用于图表)
export const useStorageTrend = () => {
  return useQuery({
    queryKey: ['dashboard', 'storage-trend'],
    queryFn: async () => {
      // 模拟7天的存储使用数据
      return [
        { date: '2024-01-09', used: 138.2, uploaded: 2.1 },
        { date: '2024-01-10', used: 140.5, uploaded: 2.3 },
        { date: '2024-01-11', used: 141.8, uploaded: 1.3 },
        { date: '2024-01-12', used: 142.9, uploaded: 1.1 },
        { date: '2024-01-13', used: 143.7, uploaded: 0.8 },
        { date: '2024-01-14', used: 144.1, uploaded: 0.4 },
        { date: '2024-01-15', used: 145.2, uploaded: 1.1 },
      ];
    },
    staleTime: 10 * 60 * 1000, // 10分钟内数据不过期
    retry: 1,
  });
};
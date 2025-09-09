# 测试指南

本项目包含完整的单元测试和E2E测试套件，确保管理端前端系统的质量和稳定性。

## 🧪 测试技术栈

### 单元测试
- **Vitest**: 基于Vite的快速单元测试框架
- **React Testing Library**: React组件测试工具
- **Jest DOM**: DOM断言匹配器
- **User Event**: 用户交互模拟

### E2E测试
- **Playwright**: 跨浏览器端到端测试
- **多浏览器支持**: Chromium, Firefox, WebKit
- **移动端测试**: 包含响应式测试

## 📁 测试文件结构

```
src/test/                     # 单元测试目录
├── components/               # 组件测试
│   ├── common/              # 通用组件测试
│   │   ├── StatCard.test.tsx
│   │   ├── ErrorBoundary.test.tsx
│   │   └── LoadingSpinner.test.tsx
│   └── ...
├── hooks/                    # Hooks测试
│   ├── useAuth.test.tsx
│   └── ...
├── pages/                    # 页面测试
│   ├── auth/
│   │   └── LoginPage.test.tsx
│   └── ...
├── setup.ts                  # 测试环境配置
├── test-utils.tsx           # 测试工具和Provider
└── sample.test.ts           # 示例测试

e2e/                         # E2E测试目录
├── auth.spec.ts             # 认证功能测试
├── navigation.spec.ts       # 导航功能测试
├── dashboard.spec.ts        # 仪表盘测试
└── ...
```

## 🚀 运行测试

### 单元测试

```bash
# 运行所有单元测试
npm run test:run

# 启动测试监视模式
npm test

# 启动可视化测试界面
npm run test:ui

# 生成测试覆盖率报告
npm run test:coverage
```

### E2E测试

```bash
# 运行所有E2E测试
npm run test:e2e

# 启动可视化E2E测试界面
npm run test:e2e:ui

# 在有头模式下运行E2E测试（显示浏览器窗口）
npm run test:e2e:headed

# 运行所有测试（单元测试 + E2E测试）
npm run test:all
```

## 🎯 测试覆盖范围

### 单元测试覆盖

✅ **通用组件**
- StatCard: 统计卡片组件渲染和交互
- ErrorBoundary: 错误边界处理和恢复
- LoadingSpinner: 加载状态和数据包装器

✅ **Hooks测试**
- useAuth: 用户认证逻辑
- 状态管理和副作用

✅ **页面组件**
- LoginPage: 登录页面表单验证和提交
- 路由和导航逻辑

### E2E测试覆盖

✅ **认证流程**
- 登录表单验证
- 认证状态管理
- 密码可见性切换
- 移动端响应式

✅ **导航功能**
- 侧边栏折叠/展开
- 页面间导航
- 面包屑导航
- 移动端导航

✅ **仪表盘功能**
- 统计数据显示
- 快速操作按钮
- 存储使用图表
- 最近活动列表

## 🛠 测试工具功能

### 测试工具 (test-utils.tsx)

```typescript
// 自定义渲染函数，包含所有Provider
import { render } from '../test-utils';

// Mock数据
import { mockUser, mockFile } from '../test-utils';

// Mock localStorage
import { mockLocalStorage } from '../test-utils';
```

### 测试环境配置

- **全局Mock**: window.matchMedia, IntersectionObserver, ResizeObserver
- **存储Mock**: localStorage, sessionStorage
- **Provider包装**: Router, QueryClient, ConfigProvider
- **国际化**: 中文语言包配置

## 📊 测试最佳实践

### 单元测试
1. **测试用户行为**: 关注用户如何与组件交互
2. **Mock外部依赖**: API调用、第三方库、浏览器API
3. **测试边界情况**: 加载状态、错误状态、空数据状态
4. **可读性**: 使用描述性的测试名称和清晰的断言

### E2E测试
1. **用户场景**: 测试完整的用户工作流程
2. **跨浏览器**: 确保在不同浏览器中的兼容性
3. **响应式**: 测试不同屏幕尺寸的行为
4. **性能**: 关注页面加载和交互性能

## 🔧 配置文件

- **vitest.config.ts**: Vitest配置，包含路径别名和测试环境
- **playwright.config.ts**: Playwright配置，包含浏览器设置和并发选项
- **src/test/setup.ts**: 测试环境初始化，全局Mock和扩展

## 📈 持续集成

测试已配置为可以在CI/CD环境中运行：

```bash
# CI环境中的测试命令
npm run test:run        # 单元测试
npm run test:e2e       # E2E测试（会自动启动开发服务器）
npm run lint           # 代码质量检查
```

## 🚨 故障排除

### 常见问题

1. **测试超时**: 增加timeout配置或检查异步操作
2. **Mock问题**: 确保正确Mock外部依赖
3. **浏览器依赖**: E2E测试需要系统支持图形界面
4. **路径问题**: 检查别名配置和导入路径

### 调试技巧

```bash
# 调试单元测试
npm test -- --reporter=verbose

# 调试E2E测试
npm run test:e2e:headed    # 显示浏览器窗口
npm run test:e2e:ui        # 使用可视化界面
```

## 📝 贡献指南

添加新测试时请遵循以下规范：

1. **文件命名**: `*.test.tsx` (单元测试) 或 `*.spec.ts` (E2E测试)
2. **测试组织**: 使用describe和it组织测试结构
3. **断言**: 使用语义化的断言方法
4. **清理**: 在适当的地方使用beforeEach/afterEach清理
5. **覆盖率**: 新功能应包含对应的测试用例

---

通过这套完整的测试体系，我们确保了管理端前端系统的质量、稳定性和用户体验。
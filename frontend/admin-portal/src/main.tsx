import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ConfigProvider } from 'antd'
import { QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import zhCN from 'antd/locale/zh_CN'
import 'antd/dist/reset.css'
import './index.css'
import App from './App.tsx'
import { THEME_CONFIG } from '@constants/index'
import { queryClient } from './lib/queryClient'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ConfigProvider
        locale={zhCN}
        theme={{
          token: {
            colorPrimary: THEME_CONFIG.PRIMARY_COLOR,
            colorSuccess: THEME_CONFIG.SUCCESS_COLOR,
            colorWarning: THEME_CONFIG.WARNING_COLOR,
            colorError: THEME_CONFIG.ERROR_COLOR,
            colorInfo: THEME_CONFIG.INFO_COLOR,
            borderRadius: 6,
            wireframe: false,
          },
          components: {
            Layout: {
              headerBg: '#fff',
              headerPadding: '0 24px',
              siderBg: '#fff',
            },
            Menu: {
              itemBg: 'transparent',
              itemSelectedBg: '#e6f7ff',
              itemSelectedColor: THEME_CONFIG.PRIMARY_COLOR,
            },
            Button: {
              borderRadius: 6,
            },
            Input: {
              borderRadius: 6,
            },
            Card: {
              borderRadius: 8,
            },
            Table: {
              // 响应式表格配置
              cellPaddingBlock: 12,
              cellPaddingInline: 16,
            },
            Form: {
              // 响应式表单配置
              itemMarginBottom: 20,
            },
          },
        }}
      >
        <App />
        {/* React Query开发者工具 - 仅在开发环境显示 */}
        <ReactQueryDevtools initialIsOpen={false} />
      </ConfigProvider>
    </QueryClientProvider>
  </StrictMode>,
)

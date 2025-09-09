import { describe, it, expect, vi, beforeAll, afterAll } from 'vitest';
import { render, screen } from '../../test-utils';
import userEvent from '@testing-library/user-event';
import { ErrorBoundary } from '@components/common/ErrorBoundary';
import React from 'react';

// 创建一个会抛出错误的组件用于测试
const ThrowError: React.FC<{ shouldThrow?: boolean }> = ({ shouldThrow = false }) => {
  if (shouldThrow) {
    throw new Error('测试错误');
  }
  return <div>正常组件</div>;
};

describe('ErrorBoundary', () => {
  // 抑制控制台错误输出以保持测试输出清洁
  const originalError = console.error;
  beforeAll(() => {
    console.error = vi.fn();
  });
  afterAll(() => {
    console.error = originalError;
  });

  it('should render children when no error occurs', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={false} />
      </ErrorBoundary>
    );

    expect(screen.getByText('正常组件')).toBeInTheDocument();
  });

  it('should render error UI when error occurs', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('页面出现错误')).toBeInTheDocument();
    expect(screen.getByText('很抱歉，页面遇到了一个错误。请尝试刷新页面，或联系技术支持。')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /刷新页面/ })).toBeInTheDocument();
  });

  it('should show error details in development mode', () => {
    const originalEnv = import.meta.env.MODE;
    (import.meta.env as any).MODE = 'development';

    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('错误详情')).toBeInTheDocument();
    expect(screen.getByText('测试错误')).toBeInTheDocument();

    // 恢复环境变量
    (import.meta.env as any).MODE = originalEnv;
  });

  it('should handle refresh button click', async () => {
    // Mock window.location.reload
    const mockReload = vi.fn();
    Object.defineProperty(window, 'location', {
      value: { reload: mockReload },
      writable: true,
    });

    const user = userEvent.setup();
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    const refreshButton = screen.getByRole('button', { name: /刷新页面/ });
    await user.click(refreshButton);

    expect(mockReload).toHaveBeenCalledOnce();
  });
});
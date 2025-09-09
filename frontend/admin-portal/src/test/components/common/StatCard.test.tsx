import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '../../test-utils';
import userEvent from '@testing-library/user-event';
import { StatCard } from '@components/common/StatCard';

describe('StatCard', () => {
  it('should render stat card with basic props', () => {
    render(
      <StatCard
        title="总用户数"
        value="156"
        icon="user"
        color="blue"
      />
    );

    expect(screen.getByText('总用户数')).toBeInTheDocument();
    expect(screen.getByText('156')).toBeInTheDocument();
  });

  it('should render trend when provided', () => {
    render(
      <StatCard
        title="文件数量"
        value="2,345"
        icon="file"
        color="green"
        trend={{ value: 12, type: 'up' }}
      />
    );

    expect(screen.getByText('文件数量')).toBeInTheDocument();
    expect(screen.getByText('2,345')).toBeInTheDocument();
    expect(screen.getByText('12%')).toBeInTheDocument();
  });

  it('should render description when provided', () => {
    const description = '相比上个月增长了12%';
    render(
      <StatCard
        title="存储使用率"
        value="29.04%"
        icon="database"
        color="orange"
        description={description}
      />
    );

    expect(screen.getByText('存储使用率')).toBeInTheDocument();
    expect(screen.getByText('29.04%')).toBeInTheDocument();
    expect(screen.getByText(description)).toBeInTheDocument();
  });

  it('should handle click events', async () => {
    const handleClick = vi.fn();
    const user = userEvent.setup();
    render(
      <StatCard
        title="点击测试"
        value="100"
        icon="user"
        color="blue"
        onClick={handleClick}
      />
    );

    const card = screen.getByRole('button');
    await user.click(card);
    
    expect(handleClick).toHaveBeenCalledOnce();
  });
});
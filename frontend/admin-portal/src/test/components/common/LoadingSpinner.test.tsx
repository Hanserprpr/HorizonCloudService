import { describe, it, expect } from 'vitest';
import { render, screen } from '../../test-utils';
import { LoadingSpinner, DataLoadingWrapper } from '@components/common/LoadingSpinner';

describe('LoadingSpinner', () => {
  it('should render loading spinner with default props', () => {
    render(<LoadingSpinner />);
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render loading spinner with custom tip', () => {
    const tip = '正在加载数据...';
    render(<LoadingSpinner tip={tip} />);
    expect(screen.getByText(tip)).toBeInTheDocument();
  });

  it('should render large size spinner', () => {
    render(<LoadingSpinner size="large" />);
    const spinner = screen.getByRole('progressbar');
    expect(spinner).toBeInTheDocument();
  });
});

describe('DataLoadingWrapper', () => {
  const mockData = [{ id: 1, name: '测试数据' }];
  const MockComponent = ({ data }: { data: typeof mockData }) => (
    <div>
      {data.map(item => (
        <div key={item.id}>{item.name}</div>
      ))}
    </div>
  );

  it('should render loading state when loading', () => {
    render(
      <DataLoadingWrapper
        loading={true}
        data={mockData}
        component={MockComponent}
        loadingTip="加载中..."
      />
    );

    expect(screen.getByText('加载中...')).toBeInTheDocument();
    expect(screen.queryByText('测试数据')).not.toBeInTheDocument();
  });

  it('should render error state when error occurs', () => {
    const error = new Error('加载失败');
    render(
      <DataLoadingWrapper
        loading={false}
        data={null}
        error={error}
        component={MockComponent}
      />
    );

    expect(screen.getByText('加载失败')).toBeInTheDocument();
    expect(screen.getByText('加载失败')).toBeInTheDocument();
  });

  it('should render empty state when no data', () => {
    render(
      <DataLoadingWrapper
        loading={false}
        data={[]}
        component={MockComponent}
      />
    );

    expect(screen.getByText('暂无数据')).toBeInTheDocument();
  });

  it('should render component when data is available', () => {
    render(
      <DataLoadingWrapper
        loading={false}
        data={mockData}
        component={MockComponent}
      />
    );

    expect(screen.getByText('测试数据')).toBeInTheDocument();
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
  });
});
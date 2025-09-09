import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '../../test-utils';
import { LoginPage } from '@pages/auth/LoginPage';
import userEvent from '@testing-library/user-event';

// Mock navigation
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => ({
  ...await vi.importActual('react-router-dom'),
  useNavigate: () => mockNavigate,
}));

// Mock useAuth hook
const mockLogin = vi.fn();
vi.mock('@hooks/useAuth', () => ({
  useAuth: () => ({
    login: mockLogin,
    loading: false,
    isAuthenticated: false,
  }),
}));

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render login form', () => {
    render(<LoginPage />);

    expect(screen.getByText('云存储管理后台')).toBeInTheDocument();
    expect(screen.getByLabelText('用户名')).toBeInTheDocument();
    expect(screen.getByLabelText('密码')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '登录' })).toBeInTheDocument();
  });

  it('should validate required fields', async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    const loginButton = screen.getByRole('button', { name: '登录' });
    await user.click(loginButton);

    expect(await screen.findByText('请输入用户名')).toBeInTheDocument();
    expect(await screen.findByText('请输入密码')).toBeInTheDocument();
  });

  it('should handle successful login', async () => {
    const user = userEvent.setup();
    mockLogin.mockResolvedValue(undefined);

    render(<LoginPage />);

    const usernameInput = screen.getByLabelText('用户名');
    const passwordInput = screen.getByLabelText('密码');
    const loginButton = screen.getByRole('button', { name: '登录' });

    await user.type(usernameInput, 'admin');
    await user.type(passwordInput, 'password123');
    await user.click(loginButton);

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('admin', 'password123');
    });

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
  });

  it('should handle login failure', async () => {
    const user = userEvent.setup();
    const errorMessage = '用户名或密码错误';
    mockLogin.mockRejectedValue(new Error(errorMessage));

    render(<LoginPage />);

    const usernameInput = screen.getByLabelText('用户名');
    const passwordInput = screen.getByLabelText('密码');
    const loginButton = screen.getByRole('button', { name: '登录' });

    await user.type(usernameInput, 'admin');
    await user.type(passwordInput, 'wrongpassword');
    await user.click(loginButton);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });

    expect(mockNavigate).not.toHaveBeenCalled();
  });

  it('should handle remember me checkbox', async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    const rememberCheckbox = screen.getByRole('checkbox', { name: '记住我' });
    
    expect(rememberCheckbox).not.toBeChecked();
    
    await user.click(rememberCheckbox);
    
    expect(rememberCheckbox).toBeChecked();
  });
});
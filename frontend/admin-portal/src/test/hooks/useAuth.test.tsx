import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAuth } from '@hooks/useAuth';
import { mockUser, mockLocalStorage } from '../test-utils';

// Mock API calls
vi.mock('@services/authService', () => ({
  AuthService: {
    login: vi.fn(),
    logout: vi.fn(),
    getCurrentUser: vi.fn(),
    refreshToken: vi.fn(),
  },
}));

const mockAuthService = vi.mocked(await import('@services/authService')).AuthService;

describe('useAuth', () => {
  const mockStorage = mockLocalStorage();

  beforeEach(() => {
    vi.clearAllMocks();
    // Mock localStorage
    global.localStorage = mockStorage as any;
  });

  it('should initialize with logged out state', () => {
    const { result } = renderHook(() => useAuth());

    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.user).toBe(null);
    expect(result.current.loading).toBe(false);
  });

  it('should handle successful login', async () => {
    const loginResponse = {
      token: 'mock-token',
      refresh_token: 'mock-refresh-token',
      user: mockUser,
    };

    mockAuthService.login.mockResolvedValue(loginResponse);

    const { result } = renderHook(() => useAuth());

    await act(async () => {
      await result.current.login('admin', 'password');
    });

    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.user).toEqual(mockUser);
    expect(mockStorage.getItem('auth-token')).toBe('mock-token');
  });

  it('should handle login failure', async () => {
    const error = new Error('Invalid credentials');
    mockAuthService.login.mockRejectedValue(error);

    const { result } = renderHook(() => useAuth());

    await expect(
      act(async () => {
        await result.current.login('admin', 'wrong-password');
      })
    ).rejects.toThrow('Invalid credentials');

    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.user).toBe(null);
  });

  it('should handle logout', async () => {
    // Setup logged in state
    mockStorage.setItem('auth-token', 'mock-token');
    mockStorage.setItem('user-info', JSON.stringify(mockUser));

    const { result } = renderHook(() => useAuth());

    await act(async () => {
      await result.current.logout();
    });

    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.user).toBe(null);
    expect(mockStorage.getItem('auth-token')).toBe(null);
  });

  it('should restore auth state from localStorage', async () => {
    mockStorage.setItem('auth-token', 'stored-token');
    mockStorage.setItem('user-info', JSON.stringify(mockUser));
    mockAuthService.getCurrentUser.mockResolvedValue(mockUser);

    const { result } = renderHook(() => useAuth());

    // Wait for initialization
    await act(async () => {
      // Trigger initialization by accessing isAuthenticated
      result.current.isAuthenticated;
    });

    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.user).toEqual(mockUser);
  });
});
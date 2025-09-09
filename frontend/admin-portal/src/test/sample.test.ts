import { describe, it, expect } from 'vitest';

describe('Sample Test', () => {
  it('should run basic tests', () => {
    expect(1 + 1).toBe(2);
    expect('hello').toBe('hello');
    expect([1, 2, 3]).toHaveLength(3);
  });

  it('should handle async operations', async () => {
    const promise = Promise.resolve('async result');
    const result = await promise;
    expect(result).toBe('async result');
  });

  it('should test objects', () => {
    const user = {
      id: 1,
      name: 'Test User',
      email: 'test@example.com'
    };

    expect(user).toMatchObject({
      id: expect.any(Number),
      name: expect.any(String),
      email: expect.stringContaining('@')
    });
  });
});
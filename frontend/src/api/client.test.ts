import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('axios', () => {
  const mockAxiosInstance = {
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
    post: vi.fn(),
    get: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  };

  return {
    default: {
      create: vi.fn(() => mockAxiosInstance),
    },
  };
});

describe('ApiClient', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('should export apiClient instance', async () => {
    const { apiClient } = await import('../api/client');
    expect(apiClient).toBeDefined();
  });

  it('should have login method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.login).toBe('function');
  });

  it('should have getFunds method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.getFunds).toBe('function');
  });

  it('should have getFund method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.getFund).toBe('function');
  });

  it('should have createFund method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.createFund).toBe('function');
  });

  it('should have updateFund method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.updateFund).toBe('function');
  });

  it('should have deleteFund method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.deleteFund).toBe('function');
  });

  it('should have getFinancials method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.getFinancials).toBe('function');
  });

  it('should have getLLMSettings method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.getLLMSettings).toBe('function');
  });

  it('should have updateLLMSettings method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.updateLLMSettings).toBe('function');
  });
});

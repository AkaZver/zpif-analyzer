import { describe, it, expect, vi, beforeEach } from 'vitest';

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

vi.mock('axios', () => {
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

  it('should call login with correct parameters', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { token: 'test-token', user: { id: 1, username: 'admin' } } };
    mockAxiosInstance.post.mockResolvedValue(mockResponse);

    const result = await apiClient.login({ username: 'admin', password: 'password' });

    expect(mockAxiosInstance.post).toHaveBeenCalledWith('/auth/login', { username: 'admin', password: 'password' });
    expect(result).toEqual(mockResponse.data);
  });

  it('should call getFunds', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: [{ id: 1, name: 'Test Fund' }] };
    mockAxiosInstance.get.mockResolvedValue(mockResponse);

    const result = await apiClient.getFunds();

    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/funds');
    expect(result).toEqual(mockResponse.data);
  });

  it('should have getFundsWithFinancials method', async () => {
    const { apiClient } = await import('../api/client');
    expect(typeof apiClient.getFundsWithFinancials).toBe('function');
  });

  it('should call getFundsWithFinancials', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: [{ id: 1, name: 'Test Fund', financials: [{ id: 1, fund_id: 1 }] }] };
    mockAxiosInstance.get.mockResolvedValue(mockResponse);

    const result = await apiClient.getFundsWithFinancials();

    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/funds-with-financials');
    expect(result).toEqual(mockResponse.data);
  });

  it('should handle getFundsWithFinancials error', async () => {
    const { apiClient } = await import('../api/client');
    const mockError = new Error('Network error');
    mockAxiosInstance.get.mockRejectedValue(mockError);

    await expect(apiClient.getFundsWithFinancials()).rejects.toThrow('Network error');
    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/funds-with-financials');
  });

  it('should call getFund with correct ID', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { id: 1, name: 'Test Fund' } };
    mockAxiosInstance.get.mockResolvedValue(mockResponse);

    const result = await apiClient.getFund(1);

    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/funds/1');
    expect(result).toEqual(mockResponse.data);
  });

  it('should call createFund with correct data', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { id: 1, name: 'New Fund' } };
    mockAxiosInstance.post.mockResolvedValue(mockResponse);

    const fundData = { name: 'New Fund', isin: 'RU000TEST001' };
    const result = await apiClient.createFund(fundData);

    expect(mockAxiosInstance.post).toHaveBeenCalledWith('/funds', fundData);
    expect(result).toEqual(mockResponse.data);
  });

  it('should call updateFund with correct parameters', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { id: 1, name: 'Updated Fund' } };
    mockAxiosInstance.put.mockResolvedValue(mockResponse);

    const fundData = { name: 'Updated Fund' };
    const result = await apiClient.updateFund(1, fundData);

    expect(mockAxiosInstance.put).toHaveBeenCalledWith('/funds/1', fundData);
    expect(result).toEqual(mockResponse.data);
  });

  it('should call deleteFund with correct ID', async () => {
    const { apiClient } = await import('../api/client');
    mockAxiosInstance.delete.mockResolvedValue({ data: {} });

    await apiClient.deleteFund(1);

    expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/funds/1');
  });

  it('should call getFinancials with correct fund ID', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: [{ id: 1, fund_id: 1, nav: 100 }] };
    mockAxiosInstance.get.mockResolvedValue(mockResponse);

    const result = await apiClient.getFinancials(1);

    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/funds/1/financials');
    expect(result).toEqual(mockResponse.data);
  });

  it('should call getLLMSettings', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { search_model_name: 'gpt-4', analysis_model_name: 'gpt-4' } };
    mockAxiosInstance.get.mockResolvedValue(mockResponse);

    const result = await apiClient.getLLMSettings();

    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/llm/settings');
    expect(result).toEqual(mockResponse.data);
  });

  it('should call updateLLMSettings with correct data', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { search_model_name: 'gpt-4', analysis_model_name: 'gpt-4' } };
    mockAxiosInstance.put.mockResolvedValue(mockResponse);

    const settings = { search_model_name: 'gpt-4', analysis_model_name: 'gpt-4' };
    const result = await apiClient.updateLLMSettings(settings);

    expect(mockAxiosInstance.put).toHaveBeenCalledWith('/llm/settings', settings);
    expect(result).toEqual(mockResponse.data);
  });

  it('should call testLLMConnection', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { search_model: { success: true }, analysis_model: { success: true } } };
    mockAxiosInstance.post.mockResolvedValue(mockResponse);

    const result = await apiClient.testLLMConnection();

    expect(mockAxiosInstance.post).toHaveBeenCalledWith('/llm/test');
    expect(result).toEqual(mockResponse.data);
  });

  it('should call getLLMModels', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: ['gpt-4', 'gpt-3.5-turbo'] };
    mockAxiosInstance.get.mockResolvedValue(mockResponse);

    const result = await apiClient.getLLMModels();

    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/llm/models');
    expect(result).toEqual(mockResponse.data);
  });

  it('should call analyzeFund with document IDs', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { id: 1, fund_id: 1, summary: 'Analysis' } };
    mockAxiosInstance.post.mockResolvedValue(mockResponse);

    const result = await apiClient.analyzeFund(1, [1, 2, 3]);

    expect(mockAxiosInstance.post).toHaveBeenCalledWith('/funds/1/analyze', { document_ids: [1, 2, 3] });
    expect(result).toEqual(mockResponse.data);
  });

  it('should call analyzeFund without document IDs', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { id: 1, fund_id: 1, summary: 'Analysis' } };
    mockAxiosInstance.post.mockResolvedValue(mockResponse);

    const result = await apiClient.analyzeFund(1);

    expect(mockAxiosInstance.post).toHaveBeenCalledWith('/funds/1/analyze', { document_ids: [] });
    expect(result).toEqual(mockResponse.data);
  });

  it('should call uploadDocument', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { id: 1, file_name: 'test.pdf' } };
    mockAxiosInstance.post.mockResolvedValue(mockResponse);

    const file = new File(['test'], 'test.pdf', { type: 'application/pdf' });
    const result = await apiClient.uploadDocument(1, file);

    expect(mockAxiosInstance.post).toHaveBeenCalled();
    expect(result).toEqual(mockResponse.data);
  });

  it('should call deleteDocument with correct parameters', async () => {
    const { apiClient } = await import('../api/client');
    mockAxiosInstance.delete.mockResolvedValue({ data: {} });

    await apiClient.deleteDocument(1, 2);

    expect(mockAxiosInstance.delete).toHaveBeenCalledWith('/funds/1/documents/2');
  });

  it('should call getDocuments with correct fund ID', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: [{ id: 1, file_name: 'test.pdf' }] };
    mockAxiosInstance.get.mockResolvedValue(mockResponse);

    const result = await apiClient.getDocuments(1);

    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/funds/1/documents');
    expect(result).toEqual(mockResponse.data);
  });

  it('should call fetchMarketData with correct fund ID', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { status: 'success' } };
    mockAxiosInstance.post.mockResolvedValue(mockResponse);

    const result = await apiClient.fetchMarketData(1);

    expect(mockAxiosInstance.post).toHaveBeenCalledWith('/funds/1/fetch-market-data');
    expect(result).toEqual(mockResponse.data);
  });

  it('should call fetchAllMarketData', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { status: 'success' } };
    mockAxiosInstance.post.mockResolvedValue(mockResponse);

    const result = await apiClient.fetchAllMarketData();

    expect(mockAxiosInstance.post).toHaveBeenCalledWith('/funds/fetch-all-market-data');
    expect(result).toEqual(mockResponse.data);
  });

  it('should call getAnalysis with correct fund ID', async () => {
    const { apiClient } = await import('../api/client');
    const mockResponse = { data: { id: 1, summary: 'Analysis' } };
    mockAxiosInstance.get.mockResolvedValue(mockResponse);

    const result = await apiClient.getAnalysis(1);

    expect(mockAxiosInstance.get).toHaveBeenCalledWith('/funds/1/analysis');
    expect(result).toEqual(mockResponse.data);
  });

  it('should call discoverDocuments with correct fund ID', async () => {
    const { apiClient } = await import('../api/client');
    mockAxiosInstance.post.mockResolvedValue({ data: {} });

    await apiClient.discoverDocuments(1);

    expect(mockAxiosInstance.post).toHaveBeenCalledWith('/funds/1/discover');
  });
});

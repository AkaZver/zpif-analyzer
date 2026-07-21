import axios, { AxiosError } from 'axios';
import type { AxiosInstance } from 'axios';
import type { LoginRequest, LoginResponse, Fund, FundFinancials, FundDocument, LLMAnalysis, LLMSettings } from '../types';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: '/api',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Interceptor для добавления JWT токена
    this.client.interceptors.request.use((config) => {
      const token = localStorage.getItem('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    // Interceptor для обработки ошибок
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('token');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // Auth
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await this.client.post<LoginResponse>('/auth/login', data);
    return response.data;
  }

  async getMe() {
    const response = await this.client.get('/auth/me');
    return response.data;
  }

  // Funds
  async getFunds(): Promise<Fund[]> {
    const response = await this.client.get<Fund[]>('/funds');
    return response.data;
  }

  async getFund(id: number): Promise<Fund> {
    const response = await this.client.get<Fund>(`/funds/${id}`);
    return response.data;
  }

  async createFund(data: Partial<Fund>): Promise<Fund> {
    const response = await this.client.post<Fund>('/funds', data);
    return response.data;
  }

  async updateFund(id: number, data: Partial<Fund>): Promise<Fund> {
    const response = await this.client.put<Fund>(`/funds/${id}`, data);
    return response.data;
  }

  async deleteFund(id: number): Promise<void> {
    await this.client.delete(`/funds/${id}`);
  }

  // Financials
  async getFinancials(fundId: number): Promise<FundFinancials[]> {
    const response = await this.client.get<FundFinancials[]>(`/funds/${fundId}/financials`);
    return response.data;
  }

  async addFinancials(fundId: number, data: Partial<FundFinancials>): Promise<FundFinancials> {
    const response = await this.client.post<FundFinancials>(`/funds/${fundId}/financials`, data);
    return response.data;
  }

  // Documents
  async getDocuments(fundId: number): Promise<FundDocument[]> {
    const response = await this.client.get<FundDocument[]>(`/funds/${fundId}/documents`);
    return response.data;
  }

  async uploadDocument(fundId: number, file: File): Promise<FundDocument> {
    const formData = new FormData();
    formData.append('file', file);
    const response = await this.client.post<FundDocument>(`/funds/${fundId}/documents`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return response.data;
  }

  async deleteDocument(fundId: number, documentId: number): Promise<void> {
    await this.client.delete(`/funds/${fundId}/documents/${documentId}`);
  }

  // Discovery
  async discoverDocuments(fundId: number): Promise<void> {
    await this.client.post(`/funds/${fundId}/discover`);
  }

  async discoverAll(): Promise<void> {
    await this.client.post('/funds/discover-all');
  }

  async getDiscoveryStatus(fundId: number): Promise<{ status: string; found: number; downloaded: number; errors: number }> {
    const response = await this.client.get(`/funds/${fundId}/discovery-status`);
    return response.data;
  }

  // Analysis
  async analyzeFund(fundId: number): Promise<LLMAnalysis> {
    const response = await this.client.post<LLMAnalysis>(`/funds/${fundId}/analyze`);
    return response.data;
  }

  async getAnalysis(fundId: number): Promise<LLMAnalysis> {
    const response = await this.client.get<LLMAnalysis>(`/funds/${fundId}/analysis`);
    return response.data;
  }

  // LLM Settings
  async getLLMSettings(): Promise<LLMSettings> {
    const response = await this.client.get<LLMSettings>('/llm/settings');
    return response.data;
  }

  async updateLLMSettings(data: Partial<LLMSettings>): Promise<LLMSettings> {
    const response = await this.client.put<LLMSettings>('/llm/settings', data);
    return response.data;
  }

  async testLLMConnection(): Promise<{ success: boolean; message: string }> {
    const response = await this.client.post('/llm/test');
    return response.data;
  }

  async getLLMModels(): Promise<string[]> {
    const response = await this.client.get<string[]>('/llm/models');
    return response.data;
  }

  // Export
  async exportExcel(): Promise<Blob> {
    const response = await this.client.get('/export/excel', { responseType: 'blob' });
    return response.data;
  }
}

export const apiClient = new ApiClient();

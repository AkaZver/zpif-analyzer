import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Settings from './index';
import { apiClient } from '../../api/client';

vi.mock('../../api/client', () => ({
  apiClient: {
    getLLMSettings: vi.fn(),
    getLLMModels: vi.fn(),
    updateLLMSettings: vi.fn(),
    testLLMConnection: vi.fn(),
  },
}));

describe('Settings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(apiClient.getLLMSettings).mockResolvedValue({
      id: 1,
      api_key_encrypted: '',
      base_url: 'https://api.openai.com/v1',
      model_name: 'gpt-4o-mini',
      proxy_enabled: false,
      proxy_url: '',
      proxy_username: '',
      proxy_password: '',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    });
    vi.mocked(apiClient.getLLMModels).mockResolvedValue(['gpt-4o-mini', 'gpt-4o']);
  });

  it('should render settings page', async () => {
    render(
      <MemoryRouter>
        <Settings />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Настройки')).toBeInTheDocument();
    });
  });

  it('should load LLM settings on mount', async () => {
    render(
      <MemoryRouter>
        <Settings />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(apiClient.getLLMSettings).toHaveBeenCalled();
    });
  });

  it('should load models on mount', async () => {
    render(
      <MemoryRouter>
        <Settings />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(apiClient.getLLMModels).toHaveBeenCalled();
    });
  });

  it('should show save and test buttons', async () => {
    render(
      <MemoryRouter>
        <Settings />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Сохранить')).toBeInTheDocument();
      expect(screen.getByText('Тест LLM')).toBeInTheDocument();
      expect(screen.getByText('Загрузить модели')).toBeInTheDocument();
    });
  });

  it('should handle settings load error', async () => {
    vi.mocked(apiClient.getLLMSettings).mockRejectedValue(new Error('Not configured'));

    render(
      <MemoryRouter>
        <Settings />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Настройки')).toBeInTheDocument();
    });
  });

  it('should handle models load error', async () => {
    vi.mocked(apiClient.getLLMModels).mockRejectedValue(new Error('Failed'));

    render(
      <MemoryRouter>
        <Settings />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Настройки')).toBeInTheDocument();
    });
  });

  it('should render proxy section', async () => {
    render(
      <MemoryRouter>
        <Settings />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Прокси')).toBeInTheDocument();
      expect(screen.getByText('Использовать прокси')).toBeInTheDocument();
    });
  });

  it('should have proxy fields disabled by default', async () => {
    render(
      <MemoryRouter>
        <Settings />
      </MemoryRouter>
    );

    await waitFor(() => {
      const proxyUrlInput = screen.getByPlaceholderText('http://proxy.example.com:8080');
      expect(proxyUrlInput).toBeDisabled();
    });
  });

  it('should enable proxy fields when checkbox is checked', async () => {
    vi.mocked(apiClient.getLLMSettings).mockResolvedValue({
      id: 1,
      api_key_encrypted: '',
      base_url: 'https://api.openai.com/v1',
      model_name: 'gpt-4o-mini',
      proxy_enabled: true,
      proxy_url: 'http://proxy.example.com:8080',
      proxy_username: 'user',
      proxy_password: '****',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    });

    render(
      <MemoryRouter>
        <Settings />
      </MemoryRouter>
    );

    await waitFor(() => {
      const proxyUrlInput = screen.getByPlaceholderText('http://proxy.example.com:8080');
      expect(proxyUrlInput).not.toBeDisabled();
    });
  });
});

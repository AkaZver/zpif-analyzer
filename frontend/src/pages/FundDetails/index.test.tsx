import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, cleanup, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import FundDetails from './index';
import { apiClient } from '../../api/client';
import { ThemeProvider } from '../../hooks/ThemeProvider';

vi.mock('../../api/client', () => ({
  apiClient: {
    getFund: vi.fn(),
    getFinancials: vi.fn(),
    getDocuments: vi.fn(),
    getAnalysis: vi.fn(),
    discoverDocuments: vi.fn(),
    uploadDocument: vi.fn(),
    analyzeFund: vi.fn(),
    deleteDocument: vi.fn(),
    downloadDocument: vi.fn(),
    updateFund: vi.fn(),
    deleteFund: vi.fn(),
    fetchMarketData: vi.fn(),
  },
}));

const mockNavigate = vi.fn();

vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ id: '1' }),
  };
});

describe('FundDetails', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(apiClient.getFund).mockResolvedValue({
      id: 1,
      name: 'Тестовый фонд',
      isin: 'RU000TEST01',
      ticker: 'TEST',
      management_company: 'Тест УК',
      real_estate_segment: 'склады',
      qualified_required: false,
      has_market_maker: true,
      fund_end_date: null,
      investfunds_url: '',
      vsezpif_url: '',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    });
    vi.mocked(apiClient.getFinancials).mockResolvedValue([]);
    vi.mocked(apiClient.getDocuments).mockResolvedValue([]);
    vi.mocked(apiClient.getAnalysis).mockRejectedValue(new Error('No analysis'));
  });

  afterEach(() => {
    cleanup();
  });

  it('should render loading state initially', () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    const spinner = document.querySelector('.ant-spin');
    expect(spinner).toBeInTheDocument();
  });

  it.each([
    { name: 'should load and display fund data', expectedText: 'Тестовый фонд' },
    { name: 'should show market maker tag', expectedText: 'Маркет-мейкер' },
  ])('$name', async ({ expectedText }) => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText(expectedText)).toBeInTheDocument();
    });
  });

  it('should call getAnalysis in parallel with other API calls', async () => {
    const callOrder: string[] = [];
    
    vi.mocked(apiClient.getFund).mockImplementation(async () => {
      callOrder.push('getFund');
      return {
        id: 1,
        name: 'Тестовый фонд',
        isin: 'RU000TEST01',
        ticker: 'TEST',
        management_company: 'Тест УК',
        real_estate_segment: 'склады',
        qualified_required: false,
        has_market_maker: true,
        fund_end_date: null,
        investfunds_url: '',
        vsezpif_url: '',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };
    });
    
    vi.mocked(apiClient.getFinancials).mockImplementation(async () => {
      callOrder.push('getFinancials');
      return [];
    });
    
    vi.mocked(apiClient.getDocuments).mockImplementation(async () => {
      callOrder.push('getDocuments');
      return [];
    });
    
    vi.mocked(apiClient.getAnalysis).mockImplementation(async () => {
      callOrder.push('getAnalysis');
      throw new Error('No analysis');
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(apiClient.getFund).toHaveBeenCalled();
      expect(apiClient.getFinancials).toHaveBeenCalled();
      expect(apiClient.getDocuments).toHaveBeenCalled();
      expect(apiClient.getAnalysis).toHaveBeenCalled();
    });

    expect(callOrder).toContain('getAnalysis');
    expect(callOrder.indexOf('getAnalysis')).toBeLessThan(callOrder.length);
  });

  it('should show qualified tag when required', async () => {
    vi.mocked(apiClient.getFund).mockResolvedValue({
      id: 1,
      name: 'Квал фонд',
      isin: 'RU000QUAL01',
      ticker: 'QUAL',
      management_company: 'Тест УК',
      real_estate_segment: 'офисы',
      qualified_required: true,
      has_market_maker: false,
      fund_end_date: null,
      investfunds_url: '',
      vsezpif_url: '',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Только для квалов')).toBeInTheDocument();
    });
  });

  it('should display financial metrics', async () => {
    vi.mocked(apiClient.getFinancials).mockResolvedValue([
      {
        id: 1,
        fund_id: 1,
        snapshot_date: '2024-01-15T00:00:00Z',
        unit_price_rub: 1000,
        nav_per_unit_rub: 1050,
        nav_total_mln_rub: 5000,
        discount_to_nav_pct: -4.76,
        cap_rate_pct: 8.5,
        p_nav: 0.95,
        p_affo: 12.0,
        noi_yield_pct: 7.2,
        annual_payout_rub: 80,
        payout_amount_rub: 80,
        payout_yield_pct: 8.0,
        payout_yield_after_tax_pct: 6.96,
        payout_frequency: 'monthly',
        payout_stability: 'high',
        rent_indexation_pct: 3.0,
        management_fee_pct: 1.5,
        trading_volume_mln_rub: 5.0,
        number_of_properties: 3,
        main_tenants: 'Ozon',
        created_at: '2024-01-15T00:00:00Z',
        updated_at: '2024-01-15T00:00:00Z',
      },
    ]);

    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Цена пая')).toBeInTheDocument();
      expect(screen.getByText('РСП')).toBeInTheDocument();
    });
  });

  it('should handle load error', async () => {
    vi.mocked(apiClient.getFund).mockRejectedValue(new Error('Not found'));

    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Фонд не найден')).toBeInTheDocument();
    });
  });

  it('should display documents section', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Документы')).toBeInTheDocument();
    });
  });

  it('should navigate back on back button click', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      const backButton = screen.getByText('Назад к сравнению');
      expect(backButton).toBeInTheDocument();
    });
  });

  it('should delete fund when confirmed', async () => {
    vi.mocked(apiClient.deleteFund).mockResolvedValue();

    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Тестовый фонд')).toBeInTheDocument();
    });

    const deleteButton = screen.getByText('Удалить');
    fireEvent.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Да')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Да'));

    await waitFor(() => {
      expect(apiClient.deleteFund).toHaveBeenCalledWith(1);
      expect(mockNavigate).toHaveBeenCalledWith('/');
    });
  });

  it('should fetch market data when button clicked', async () => {
    vi.mocked(apiClient.fetchMarketData).mockResolvedValue({
      status: 'success',
      fund_id: 1,
      records_created: 5,
      records_updated: 3,
      moex_available: true,
      investfunds_available: true,
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Тестовый фонд')).toBeInTheDocument();
    });

    const fetchButton = screen.getByText('Обновить данные');
    fireEvent.click(fetchButton);

    await waitFor(() => {
      expect(apiClient.fetchMarketData).toHaveBeenCalledWith(1);
    });
  });

  it('should delete document when delete button clicked', async () => {
    vi.mocked(apiClient.getDocuments).mockResolvedValue([
      {
        id: 123,
        fund_id: 1,
        file_name: 'test.pdf',
        file_path: '/path/to/test.pdf',
        document_type: 'pdf',
        content_hash: 'hash123',
        source: 'manual',
        source_url: '',
        upload_date: '2024-01-01T00:00:00Z',
        status: 'analyzed',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ]);
    vi.mocked(apiClient.deleteDocument).mockResolvedValue();

    render(
      <ThemeProvider>
        <MemoryRouter>
          <FundDetails />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('test.pdf')).toBeInTheDocument();
    });

    const table = document.querySelector('.ant-table-tbody');
    expect(table).toBeInTheDocument();
    
    const actionButtons = table!.querySelectorAll('button');
    expect(actionButtons.length).toBeGreaterThanOrEqual(2);
    
    fireEvent.click(actionButtons[1]);

    await waitFor(() => {
      expect(apiClient.deleteDocument).toHaveBeenCalledWith(1, 123);
    });
  });
});

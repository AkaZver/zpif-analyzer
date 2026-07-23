import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import FundDetails from './index';
import { apiClient } from '../../api/client';

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

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
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

  it('should render loading state initially', () => {
    render(
      <MemoryRouter>
        <FundDetails />
      </MemoryRouter>
    );

    const spinner = document.querySelector('.ant-spin');
    expect(spinner).toBeInTheDocument();
  });

  it('should load and display fund data', async () => {
    render(
      <MemoryRouter>
        <FundDetails />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Тестовый фонд')).toBeInTheDocument();
    });
  });

  it('should show market maker tag', async () => {
    render(
      <MemoryRouter>
        <FundDetails />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Маркет-мейкер')).toBeInTheDocument();
    });
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
      <MemoryRouter>
        <FundDetails />
      </MemoryRouter>
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
      <MemoryRouter>
        <FundDetails />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Цена пая')).toBeInTheDocument();
      expect(screen.getByText('РСП')).toBeInTheDocument();
    });
  });

  it('should handle load error', async () => {
    vi.mocked(apiClient.getFund).mockRejectedValue(new Error('Not found'));

    render(
      <MemoryRouter>
        <FundDetails />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Фонд не найден')).toBeInTheDocument();
    });
  });

  it('should display documents section', async () => {
    render(
      <MemoryRouter>
        <FundDetails />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Документы')).toBeInTheDocument();
    });
  });

  it('should navigate back on back button click', async () => {
    render(
      <MemoryRouter>
        <FundDetails />
      </MemoryRouter>
    );

    await waitFor(() => {
      const backButton = screen.getByText('Назад к сравнению');
      expect(backButton).toBeInTheDocument();
    });
  });
});

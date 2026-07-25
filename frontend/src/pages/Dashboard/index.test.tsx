import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Dashboard from './index';
import { apiClient } from '../../api/client';
import { ThemeProvider } from '../../hooks/ThemeProvider';

vi.mock('../../api/client', () => ({
  apiClient: {
    getFunds: vi.fn(),
    getFinancials: vi.fn(),
    exportExcel: vi.fn(),
    enrichAndCreateFund: vi.fn(),
    fetchAllMarketData: vi.fn(),
  },
}));

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockFunds = [
  {
    id: 1,
    name: 'Бета Фонд',
    isin: 'RU000BETA01',
    ticker: 'BETA',
    management_company: 'Бета УК',
    real_estate_segment: 'склады',
    qualified_required: false,
    has_market_maker: false,
    fund_end_date: null,
    investfunds_url: '',
    vsezpif_url: '',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    name: 'Альфа Фонд',
    isin: 'RU000ALFA01',
    ticker: 'ALFA',
    management_company: 'Альфа УК',
    real_estate_segment: 'офисы',
    qualified_required: true,
    has_market_maker: true,
    fund_end_date: null,
    investfunds_url: '',
    vsezpif_url: '',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 3,
    name: 'Гамма Фонд',
    isin: 'RU000GAMM01',
    ticker: 'GAMM',
    management_company: 'Гамма УК',
    real_estate_segment: 'ТЦ',
    qualified_required: false,
    has_market_maker: false,
    fund_end_date: null,
    investfunds_url: '',
    vsezpif_url: '',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

const mockFinancials = [
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
];

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(apiClient.getFunds).mockResolvedValue(mockFunds);
    vi.mocked(apiClient.getFinancials).mockResolvedValue(mockFinancials);
  });

  it('should render dashboard with funds table', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Сравнение ЗПИФ')).toBeInTheDocument();
    });
  });

  it('should display all fund names', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
      expect(screen.getByText('Бета Фонд')).toBeInTheDocument();
      expect(screen.getByText('Гамма Фонд')).toBeInTheDocument();
    });
  });

  it('should have default sort by name ascending', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      const rows = screen.getAllByRole('row');
      const dataRows = rows.slice(1);
      
      const names = dataRows.map(row => {
        const cells = row.querySelectorAll('td');
        return cells[0]?.textContent;
      });

      expect(names[0]).toContain('Альфа Фонд');
      expect(names[1]).toContain('Бета Фонд');
      expect(names[2]).toContain('Гамма Фонд');
    });
  });

  it('should have sorter on all columns', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      const headers = screen.getAllByRole('columnheader');
      const columnNames = ['Название', 'ISIN', 'УК', 'Сегмент', 'Цена пая', 'РСП', 'Дисконт', 'Доходность выплат', 'Квал'];
      
      columnNames.forEach(name => {
        const header = headers.find(h => h.textContent?.includes(name));
        expect(header).toBeDefined();
      });
    });
  });

  it('should sort by ISIN when clicking ISIN header', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
    });

    const isinHeaders = screen.getAllByText('ISIN');
    fireEvent.click(isinHeaders[0]);

    await waitFor(() => {
      const rows = screen.getAllByRole('row');
      const dataRows = rows.slice(1);
      
      const isins = dataRows.map(row => {
        const cells = row.querySelectorAll('td');
        return cells[1]?.textContent;
      });

      expect(isins[0]).toContain('RU000ALFA01');
    });
  });

  it('should sort by УК when clicking УК header', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
    });

    const ukHeaders = screen.getAllByText('УК');
    const ukHeader = ukHeaders.find(el => el.classList.contains('ant-table-column-title'));
    fireEvent.click(ukHeader!);

    await waitFor(() => {
      const rows = screen.getAllByRole('row');
      const dataRows = rows.slice(1);
      
      const companies = dataRows.map(row => {
        const cells = row.querySelectorAll('td');
        return cells[2]?.textContent;
      });

      expect(companies[0]).toContain('Альфа УК');
    });
  });

  it('should sort by Сегмент when clicking Сегмент header', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
    });

    const segmentHeaders = screen.getAllByText('Сегмент');
    const segmentHeader = segmentHeaders.find(el => el.classList.contains('ant-table-column-title'));
    fireEvent.click(segmentHeader!);

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
    });
  });

  it('should sort by РСП when clicking РСП header', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
    });

    const navHeaders = screen.getAllByText('РСП');
    fireEvent.click(navHeaders[0]);

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
    });
  });

  it('should sort by Квал when clicking Квал header', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
    });

    const qualHeaders = screen.getAllByText('Квал');
    fireEvent.click(qualHeaders[0]);

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
    });
  });

  it('should show filter controls', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Сегмент')).toBeInTheDocument();
      expect(screen.getByText('Только для квалов')).toBeInTheDocument();
    });
  });

  it('should show action buttons', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Добавить')).toBeInTheDocument();
      expect(screen.getByText('Обновить данные')).toBeInTheDocument();
      expect(screen.getByText('Экспорт')).toBeInTheDocument();
    });
  });

  it('should navigate to fund details on row click', async () => {
    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Альфа Фонд')).toBeInTheDocument();
    });

    const fundRow = screen.getByText('Альфа Фонд').closest('tr');
    fireEvent.click(fundRow!);

    expect(mockNavigate).toHaveBeenCalledWith('/funds/2');
  });

  it('should handle load error', async () => {
    vi.mocked(apiClient.getFunds).mockRejectedValue(new Error('Failed'));

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      </ThemeProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Сравнение ЗПИФ')).toBeInTheDocument();
    });
  });
});

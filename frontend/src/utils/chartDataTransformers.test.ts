import { describe, it, expect } from 'vitest';
import {
  groupFinancialsByMonth,
  buildPayoutChartData,
  getTradingStartFormatted,
} from './chartDataTransformers';
import type { FundFinancials } from '../types';

const createFinancial = (overrides: Partial<FundFinancials> = {}): FundFinancials => ({
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
  payout_amount_rub: 0,
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
  ...overrides,
});

describe('groupFinancialsByMonth', () => {
  it('should group by month and take latest entry', () => {
    const financials = [
      createFinancial({ id: 1, snapshot_date: '2024-07-01T00:00:00Z', unit_price_rub: 1000 }),
      createFinancial({ id: 2, snapshot_date: '2024-07-15T00:00:00Z', unit_price_rub: 1010 }),
      createFinancial({ id: 3, snapshot_date: '2024-08-01T00:00:00Z', unit_price_rub: 1020 }),
    ];

    const result = groupFinancialsByMonth(financials);

    expect(result).toHaveLength(2);
    expect(result[0].id).toBe(2);
    expect(result[1].id).toBe(3);
  });

  it('should apply filter function', () => {
    const financials = [
      createFinancial({ id: 1, snapshot_date: '2024-07-01T00:00:00Z', payout_amount_rub: 100 }),
      createFinancial({ id: 2, snapshot_date: '2024-07-15T00:00:00Z', payout_amount_rub: 0 }),
      createFinancial({ id: 3, snapshot_date: '2024-08-01T00:00:00Z', payout_amount_rub: 200 }),
    ];

    const result = groupFinancialsByMonth(financials, (f) => f.payout_amount_rub > 0);

    expect(result).toHaveLength(2);
    expect(result[0].id).toBe(1);
    expect(result[1].id).toBe(3);
  });

  it('should return empty array for empty input', () => {
    const result = groupFinancialsByMonth([]);
    expect(result).toHaveLength(0);
  });

  it('should not update entry when new date is not later', () => {
    const financials = [
      createFinancial({ id: 1, snapshot_date: '2024-07-15T00:00:00Z', unit_price_rub: 1010 }),
      createFinancial({ id: 2, snapshot_date: '2024-07-01T00:00:00Z', unit_price_rub: 1000 }),
    ];

    const result = groupFinancialsByMonth(financials);

    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
  });
});

describe('buildPayoutChartData', () => {
  it('should build payout chart data with correct format', () => {
    const financials = [
      createFinancial({
        id: 1,
        snapshot_date: '2024-07-01T00:00:00Z',
        payout_amount_rub: 4200,
      }),
      createFinancial({
        id: 2,
        snapshot_date: '2024-08-01T00:00:00Z',
        payout_amount_rub: 4300,
      }),
    ];

    const result = buildPayoutChartData(financials);

    expect(result).toHaveLength(2);
    expect(result[0]).toEqual({
      date: '07.24',
      'Выплата': 4200,
    });
    expect(result[1]).toEqual({
      date: '08.24',
      'Выплата': 4300,
    });
  });

  it('should filter out entries with zero payout', () => {
    const financials = [
      createFinancial({ id: 1, snapshot_date: '2024-07-01T00:00:00Z', payout_amount_rub: 4200 }),
      createFinancial({ id: 2, snapshot_date: '2024-08-01T00:00:00Z', payout_amount_rub: 0 }),
      createFinancial({ id: 3, snapshot_date: '2024-09-01T00:00:00Z', payout_amount_rub: 4400 }),
    ];

    const result = buildPayoutChartData(financials);

    expect(result).toHaveLength(2);
    expect(result[0].date).toBe('07.24');
    expect(result[1].date).toBe('09.24');
  });

  it('should group multiple payouts in same month', () => {
    const financials = [
      createFinancial({
        id: 1,
        snapshot_date: '2024-07-01T00:00:00Z',
        payout_amount_rub: 4200,
      }),
      createFinancial({
        id: 2,
        snapshot_date: '2024-07-15T00:00:00Z',
        payout_amount_rub: 4200,
      }),
      createFinancial({
        id: 3,
        snapshot_date: '2024-07-20T00:00:00Z',
        payout_amount_rub: 4200,
      }),
    ];

    const result = buildPayoutChartData(financials);

    expect(result).toHaveLength(1);
    expect(result[0].date).toBe('07.24');
    expect(result[0]['Выплата']).toBe(4200);
  });
});

describe('getTradingStartFormatted', () => {
  it('should return formatted date of first trading entry', () => {
    const financials = [
      createFinancial({ id: 1, snapshot_date: '2024-06-01T00:00:00Z', unit_price_rub: 1000 }),
      createFinancial({ id: 2, snapshot_date: '2024-07-01T00:00:00Z', unit_price_rub: 1010 }),
    ];

    const result = getTradingStartFormatted(financials);

    expect(result).toBe('06.24');
  });

  it('should return null when no trading entries exist', () => {
    const financials = [
      createFinancial({ id: 1, snapshot_date: '2024-06-01T00:00:00Z', unit_price_rub: 0 }),
      createFinancial({ id: 2, snapshot_date: '2024-07-01T00:00:00Z', unit_price_rub: 0 }),
    ];

    const result = getTradingStartFormatted(financials);

    expect(result).toBeNull();
  });

  it('should use all financials to find first trading date', () => {
    const allFinancials = [
      createFinancial({ id: 1, snapshot_date: '2024-01-01T00:00:00Z', unit_price_rub: 1000 }),
      createFinancial({ id: 2, snapshot_date: '2024-06-01T00:00:00Z', unit_price_rub: 1010 }),
    ];

    const result = getTradingStartFormatted(allFinancials);

    expect(result).toBe('01.24');
  });

  it('should return null for empty financials', () => {
    const result = getTradingStartFormatted([]);
    expect(result).toBeNull();
  });
});

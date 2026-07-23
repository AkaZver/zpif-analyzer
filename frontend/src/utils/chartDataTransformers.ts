import type { FundFinancials } from '../types';
import { formatMonthYear } from './dateFormatters';

export interface PayoutChartPoint {
  date: string;
  'Выплата': number;
}

export interface PriceChartPoint {
  date: string;
  'Цена пая': number | null;
  'РСП': number;
}

export const groupFinancialsByMonth = (
  financials: FundFinancials[],
  filterFn?: (f: FundFinancials) => boolean
): FundFinancials[] => {
  const grouped = new Map<string, FundFinancials>();

  const filtered = filterFn ? financials.filter(filterFn) : financials;

  filtered.forEach((f) => {
    const date = new Date(f.snapshot_date);
    const key = `${date.getFullYear()}-${date.getMonth()}`;

    const existing = grouped.get(key);
    if (!existing || new Date(f.snapshot_date) > new Date(existing.snapshot_date)) {
      grouped.set(key, f);
    }
  });

  return Array.from(grouped.values()).sort(
    (a, b) => new Date(a.snapshot_date).getTime() - new Date(b.snapshot_date).getTime()
  );
};

export const buildPayoutChartData = (financials: FundFinancials[]): PayoutChartPoint[] => {
  const grouped = groupFinancialsByMonth(financials, (f) => f.payout_amount_rub > 0);

  return grouped.map((f) => ({
    date: formatMonthYear(new Date(f.snapshot_date)),
    'Выплата': f.payout_amount_rub,
  }));
};

export const getTradingStartFormatted = (
  financials: FundFinancials[]
): string | null => {
  const firstTradingDate = financials
    .filter((f) => f.unit_price_rub > 0)
    .sort((a, b) => new Date(a.snapshot_date).getTime() - new Date(b.snapshot_date).getTime())[0];

  if (!firstTradingDate) {
    return null;
  }

  const tradingStartDate = new Date(firstTradingDate.snapshot_date);
  return formatMonthYear(tradingStartDate);
};

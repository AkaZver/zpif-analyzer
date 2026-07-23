import { describe, it, expect } from 'vitest';
import { formatMonthYear } from './dateFormatters';

describe('formatMonthYear', () => {
  it('should format date as MM.YY', () => {
    const date = new Date(2024, 6, 15); // July 15, 2024
    expect(formatMonthYear(date)).toBe('07.24');
  });

  it('should pad single-digit months with zero', () => {
    const date = new Date(2025, 0, 1); // January 1, 2025
    expect(formatMonthYear(date)).toBe('01.25');
  });

  it('should handle December correctly', () => {
    const date = new Date(2023, 11, 31); // December 31, 2023
    expect(formatMonthYear(date)).toBe('12.23');
  });

  it('should handle year 2000', () => {
    const date = new Date(2000, 5, 10); // June 10, 2000
    expect(formatMonthYear(date)).toBe('06.00');
  });
});

export const formatMonthYear = (date: Date): string =>
  `${String(date.getMonth() + 1).padStart(2, '0')}.${String(date.getFullYear()).slice(-2)}`;

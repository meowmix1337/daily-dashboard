export function formatStockPrice(price: number, symbol: string): string {
  if (symbol === 'BTC') {
    return `${(price / 1000).toFixed(1)}k`;
  }
  return price.toFixed(2);
}

export function cn(...classes: (string | undefined | false | null)[]): string {
  return classes.filter(Boolean).join(' ');
}

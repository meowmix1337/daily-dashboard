import type { DashboardResponse, NewsCategory, Task, StockQuote, SymbolSearchResult } from '../types/dashboard';

const BASE = '/api';

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const { headers: extraHeaders, ...rest } = options ?? {};
  const res = await fetch(`${BASE}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(extraHeaders as Record<string, string>) },
    ...rest,
  });
  if (!res.ok) {
    throw new Error(`API error ${res.status}: ${res.statusText}`);
  }
  return res.json() as Promise<T>;
}

export function fetchDashboard(): Promise<DashboardResponse> {
  return apiFetch<DashboardResponse>('/dashboard');
}

export function fetchNews(): Promise<NewsCategory[]> {
  return apiFetch<NewsCategory[]>('/news');
}

export function toggleTask(id: string, done: boolean): Promise<Task> {
  return apiFetch<Task>(`/tasks/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ done }),
  });
}

export function createTask(text: string, priority: string): Promise<Task> {
  return apiFetch<Task>('/tasks', {
    method: 'POST',
    body: JSON.stringify({ text, priority }),
  });
}

export function fetchStocks(): Promise<StockQuote[]> {
  return apiFetch<StockQuote[]>('/stocks');
}

export function addStockSymbol(symbol: string): Promise<{ symbols: string[] }> {
  return apiFetch<{ symbols: string[] }>('/stocks/watchlist', {
    method: 'POST',
    body: JSON.stringify({ symbol }),
  });
}

export function removeStockSymbol(symbol: string): Promise<{ symbols: string[] }> {
  return apiFetch<{ symbols: string[] }>(`/stocks/watchlist/${encodeURIComponent(symbol)}`, {
    method: 'DELETE',
  });
}

export function searchSymbols(query: string): Promise<{ results: SymbolSearchResult[] }> {
  return apiFetch<{ results: SymbolSearchResult[] }>(`/stocks/search?q=${encodeURIComponent(query)}`);
}

export function deleteTask(id: string): Promise<void> {
  return apiFetch(`/tasks/${id}`, { method: 'DELETE' }).then(() => undefined);
}

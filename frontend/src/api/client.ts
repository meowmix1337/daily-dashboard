import type { DashboardResponse, Task, StockQuote, SymbolSearchResult } from '../types/dashboard';

const BASE = '/api';

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    throw new Error(`API error ${res.status}: ${res.statusText}`);
  }
  return res.json() as Promise<T>;
}

export function fetchDashboard(): Promise<DashboardResponse> {
  return apiFetch<DashboardResponse>('/dashboard');
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

export async function deleteTask(id: string): Promise<void> {
  const res = await fetch(`${BASE}/tasks/${id}`, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
  });
  if (!res.ok) {
    throw new Error(`API error ${res.status}: ${res.statusText}`);
  }
}

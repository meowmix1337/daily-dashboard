import type { DashboardResponse, NewsCategory, Task, StockQuote, SymbolSearchResult, TaskLabel, UserSettings, NewsCategoriesResponse } from '../types/dashboard';

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
  if (res.status === 204 || res.headers.get('content-length') === '0') {
    return undefined as T;
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

export function fetchUserSettings(): Promise<UserSettings> {
  return apiFetch<UserSettings>('/settings');
}

export function upsertUserSettings(settings: Partial<UserSettings>): Promise<UserSettings> {
  return apiFetch<UserSettings>('/settings', { method: 'PUT', body: JSON.stringify(settings) });
}

export function fetchNewsCategories(): Promise<NewsCategoriesResponse> {
  return apiFetch<NewsCategoriesResponse>('/settings/news-categories');
}

export function setNewsCategories(categoryIds: string[]): Promise<void> {
  return apiFetch('/settings/news-categories', { method: 'PUT', body: JSON.stringify({ category_ids: categoryIds }) }).then(() => undefined);
}
export function fetchLabels(): Promise<TaskLabel[]> {
  return apiFetch<TaskLabel[]>('/labels');
}
export function createLabel(name: string, color: string): Promise<TaskLabel> {
  return apiFetch<TaskLabel>('/labels', { method: 'POST', body: JSON.stringify({ name, color }) });
}
export function updateLabel(id: string, name?: string, color?: string): Promise<TaskLabel> {
  return apiFetch<TaskLabel>(`/labels/${id}`, { method: 'PATCH', body: JSON.stringify({ name, color }) });
}
export function deleteLabel(id: string): Promise<void> {
  return apiFetch(`/labels/${id}`, { method: 'DELETE' }).then(() => undefined);
}
export function fetchTaskLabels(taskId: string): Promise<TaskLabel[]> {
  return apiFetch<TaskLabel[]>(`/tasks/${encodeURIComponent(taskId)}/labels`);
}
export function assignLabelToTask(taskId: string, labelId: string): Promise<void> {
  return apiFetch(`/tasks/${encodeURIComponent(taskId)}/labels`, { method: 'POST', body: JSON.stringify({ label_id: labelId }) }).then(() => undefined);
}
export function removeLabelFromTask(taskId: string, labelId: string): Promise<void> {
  return apiFetch(`/tasks/${encodeURIComponent(taskId)}/labels/${encodeURIComponent(labelId)}`, { method: 'DELETE' }).then(() => undefined);
}

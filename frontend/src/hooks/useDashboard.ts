import { useQuery } from '@tanstack/react-query';
import { fetchDashboard } from '../api/client';
import type { DashboardResponse } from '../types/dashboard';

export function useDashboard() {
  return useQuery<DashboardResponse, Error>({
    queryKey: ['dashboard'],
    queryFn: fetchDashboard,
    staleTime: 60_000,
    refetchInterval: 30_000,
    retry: 2,
  });
}

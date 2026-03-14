import { useQuery } from '@tanstack/react-query';
import { fetchDashboard } from '../api/client';
import type { DashboardResponse } from '../types/dashboard';
import { useAuth } from './useAuth';

export function useDashboard() {
  const { isAuthenticated } = useAuth();
  return useQuery<DashboardResponse, Error>({
    queryKey: ['dashboard'],
    queryFn: fetchDashboard,
    staleTime: 60_000,
    refetchInterval: 30_000,
    retry: 2,
    enabled: isAuthenticated,
  });
}

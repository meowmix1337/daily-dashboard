import { useQuery } from '@tanstack/react-query';
import { fetchNews } from '../api/client';

export function useNews() {
  return useQuery({
    queryKey: ['news'],
    queryFn: fetchNews,
    staleTime: Infinity,
    refetchOnWindowFocus: false,
    refetchInterval: false,
    retry: 1,
  });
}
